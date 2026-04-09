package engine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"relay-judge/internal/subject"
)

const (
	defaultCCompiler = "cc"
	cCompileTimeout  = 10 * time.Second
	cReturnMarker    = "__RELAY_RETURN__:"
)

var (
	cPrototypePattern = regexp.MustCompile(`^\s*(.+?)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	intPattern        = regexp.MustCompile(`-?\d+`)
)

type cMode string

const (
	cModeStdoutInts cMode = "stdout_ints"
	cModeReturnInt  cMode = "return_int"
	cModeReturnBool cMode = "return_bool"
)

func runC(spec subject.Subject, options Options, report Report, startedAt time.Time) (Report, error) {
	mode, err := determineCMode(spec)
	if err != nil {
		report.Status = "load_error"
		report.Message = err.Error()
		report.DurationMs = time.Since(startedAt).Seconds() * 1000
		return report, nil
	}

	tempDir, err := os.MkdirTemp("", "relay-judge-c-*")
	if err != nil {
		return report, err
	}
	defer os.RemoveAll(tempDir)

	harnessSource, err := buildCHarness(spec, mode)
	if err != nil {
		report.Status = "load_error"
		report.Message = err.Error()
		report.DurationMs = time.Since(startedAt).Seconds() * 1000
		return report, nil
	}

	harnessPath := filepath.Join(tempDir, "runner.c")
	binaryPath := filepath.Join(tempDir, "runner")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	if err := os.WriteFile(harnessPath, []byte(harnessSource), 0o644); err != nil {
		return report, err
	}

	compiler := options.CCompiler
	if strings.TrimSpace(compiler) == "" {
		compiler = defaultCCompiler
	}

	compileCtx, cancelCompile := context.WithTimeout(context.Background(), cCompileTimeout)
	defer cancelCompile()

	compileCmd := exec.CommandContext(
		compileCtx,
		compiler,
		"-std=c11",
		"-O2",
		"-o",
		binaryPath,
		harnessPath,
		options.SubmissionPath,
	)

	var compileStdout bytes.Buffer
	var compileStderr bytes.Buffer
	compileCmd.Stdout = &compileStdout
	compileCmd.Stderr = &compileStderr

	if err := compileCmd.Run(); err != nil {
		report.Status = "load_error"
		if compileCtx.Err() == context.DeadlineExceeded {
			report.Message = "c compilation timed out"
		} else {
			report.Message = strings.TrimSpace(joinNonEmpty(compileStderr.String(), compileStdout.String(), err.Error()))
		}
		report.DurationMs = time.Since(startedAt).Seconds() * 1000
		return report, nil
	}

	results := make([]wrapperTestResult, 0, len(spec.Tests))
	for index, test := range spec.Tests {
		timeout := time.Duration(spec.TimeLimitMs) * time.Millisecond
		runCtx, cancelRun := context.WithTimeout(context.Background(), timeout)

		cmd := exec.CommandContext(runCtx, binaryPath, strconv.Itoa(index))
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		testStartedAt := time.Now()
		err := cmd.Run()
		durationMs := time.Since(testStartedAt).Seconds() * 1000
		cancelRun()

		if runCtx.Err() == context.DeadlineExceeded {
			report.Status = "timeout"
			report.Message = fmt.Sprintf("c process exceeded %dms", spec.TimeLimitMs)
			report.DurationMs = time.Since(startedAt).Seconds() * 1000
			return report, nil
		}

		if err != nil {
			results = append(results, wrapperTestResult{
				Name:       test.Name,
				Group:      test.Group,
				Status:     "runtime_error",
				Error:      strings.TrimSpace(joinNonEmpty(stderr.String(), stdout.String(), err.Error())),
				DurationMs: durationMs,
				Stdout:     stdout.String(),
				Stderr:     stderr.String(),
			})
			break
		}

		actual, err := decodeCActual(stdout.String(), mode)
		if err != nil {
			results = append(results, wrapperTestResult{
				Name:       test.Name,
				Group:      test.Group,
				Status:     "runtime_error",
				Error:      err.Error(),
				DurationMs: durationMs,
				Stdout:     stdout.String(),
				Stderr:     stderr.String(),
			})
			break
		}

		results = append(results, wrapperTestResult{
			Name:       test.Name,
			Group:      test.Group,
			Status:     "ok",
			Actual:     actual,
			DurationMs: durationMs,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
		})
	}

	report.DurationMs = time.Since(startedAt).Seconds() * 1000
	report = evaluateTestResults(spec, report, results)
	return report, nil
}

func determineCMode(spec subject.Subject) (cMode, error) {
	switch spec.NormalizedResultSource() {
	case "stdout_ints":
		return cModeStdoutInts, nil
	case "return":
		switch inferCReturnType(spec) {
		case "int":
			return cModeReturnInt, nil
		case "bool":
			return cModeReturnBool, nil
		default:
			return "", fmt.Errorf("unsupported C return type in prototype %q", spec.Prototype)
		}
	default:
		return "", fmt.Errorf("unsupported C result_source %q", spec.NormalizedResultSource())
	}
}

func inferCReturnType(spec subject.Subject) string {
	matches := cPrototypePattern.FindStringSubmatch(sanitizeCPrototype(spec.Prototype))
	if len(matches) != 3 {
		return ""
	}

	if matches[2] != spec.FunctionName {
		return ""
	}

	switch strings.TrimSpace(matches[1]) {
	case "void", "int", "bool":
		return strings.TrimSpace(matches[1])
	default:
		return ""
	}
}

func sanitizeCPrototype(prototype string) string {
	value := strings.TrimSpace(prototype)
	value = strings.TrimSuffix(value, ":")
	value = strings.TrimSuffix(value, ";")
	return value
}

func buildCHarness(spec subject.Subject, mode cMode) (string, error) {
	prototype := sanitizeCPrototype(spec.Prototype)
	if prototype == "" {
		return "", fmt.Errorf("empty C prototype")
	}

	var cases strings.Builder
	for index, test := range spec.Tests {
		caseSource, err := buildCTestCase(index, spec.FunctionName, test.Args, mode)
		if err != nil {
			return "", fmt.Errorf("generate C harness for test %q: %w", test.Name, err)
		}
		cases.WriteString(caseSource)
	}

	return fmt.Sprintf(`#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

%s;

static int run_test(int index) {
  switch (index) {
%s    default:
      fprintf(stderr, "unknown test index %%d\n", index);
      return 64;
  }
}

int main(int argc, char **argv) {
  if (argc != 2) {
    fprintf(stderr, "expected test index argument\n");
    return 64;
  }

  char *end = NULL;
  long index = strtol(argv[1], &end, 10);
  if (end == argv[1] || *end != '\0') {
    fprintf(stderr, "invalid test index: %%s\n", argv[1]);
    return 64;
  }

  return run_test((int) index);
}
`, prototype, cases.String()), nil
}

func buildCTestCase(index int, functionName string, args []any, mode cMode) (string, error) {
	declarations := make([]string, 0, len(args))
	argNames := make([]string, 0, len(args))

	for argIndex, arg := range args {
		declaration, argName, err := renderCArg(argIndex, arg)
		if err != nil {
			return "", err
		}
		declarations = append(declarations, declaration)
		argNames = append(argNames, argName)
	}

	call := fmt.Sprintf("%s(%s)", functionName, strings.Join(argNames, ", "))

	var body strings.Builder
	fmt.Fprintf(&body, "    case %d: {\n", index)
	for _, declaration := range declarations {
		body.WriteString(indentCBlock(declaration, "      "))
	}

	switch mode {
	case cModeStdoutInts:
		fmt.Fprintf(&body, "      %s;\n", call)
	case cModeReturnInt:
		fmt.Fprintf(&body, "      printf(\"\\n%s%%d\\n\", %s);\n", cReturnMarker, call)
	case cModeReturnBool:
		fmt.Fprintf(&body, "      printf(\"\\n%s%%s\\n\", %s ? \"true\" : \"false\");\n", cReturnMarker, call)
	default:
		return "", fmt.Errorf("unsupported C mode %q", mode)
	}

	body.WriteString("      return 0;\n")
	body.WriteString("    }\n")
	return body.String(), nil
}

func renderCArg(index int, value any) (string, string, error) {
	name := fmt.Sprintf("arg%d", index)

	if scalar, ok := toCInt(value); ok {
		return fmt.Sprintf("int %s = %d;\n", name, scalar), name, nil
	}

	switch typed := value.(type) {
	case bool:
		return fmt.Sprintf("bool %s = %s;\n", name, strconv.FormatBool(typed)), name, nil
	case string:
		return fmt.Sprintf("const char *%s = %s;\n", name, strconv.Quote(typed)), name, nil
	}

	if items, ok := toCIntSlice(value); ok {
		if len(items) == 0 {
			return fmt.Sprintf("int *%s = NULL;\n", name), name, nil
		}
		return fmt.Sprintf("int %s[] = {%s};\n", name, joinCInts(items)), name, nil
	}

	return "", "", fmt.Errorf("unsupported C argument type %T", value)
}

func decodeCActual(stdout string, mode cMode) (any, error) {
	switch mode {
	case cModeStdoutInts:
		matches := intPattern.FindAllString(stdout, -1)
		values := make([]int, 0, len(matches))
		for _, match := range matches {
			value, err := strconv.Atoi(match)
			if err != nil {
				return nil, fmt.Errorf("parse stdout integers: %w", err)
			}
			values = append(values, value)
		}
		return values, nil
	case cModeReturnInt:
		raw, err := extractMarkedReturn(stdout)
		if err != nil {
			return nil, err
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("parse C int return value %q: %w", raw, err)
		}
		return value, nil
	case cModeReturnBool:
		raw, err := extractMarkedReturn(stdout)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("parse C bool return value %q", raw)
		}
	default:
		return nil, fmt.Errorf("unsupported C mode %q", mode)
	}
}

func extractMarkedReturn(stdout string) (string, error) {
	index := strings.LastIndex(stdout, cReturnMarker)
	if index < 0 {
		return "", fmt.Errorf("missing C return marker in stdout")
	}

	value := stdout[index+len(cReturnMarker):]
	value = strings.SplitN(value, "\n", 2)[0]
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty C return marker payload")
	}
	return value, nil
}

func indentCBlock(value, prefix string) string {
	lines := strings.Split(strings.TrimSuffix(value, "\n"), "\n")
	var builder strings.Builder
	for _, line := range lines {
		if line == "" {
			builder.WriteString(prefix)
			builder.WriteString("\n")
			continue
		}
		builder.WriteString(prefix)
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return builder.String()
}

func joinCInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ", ")
}

func toCInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), float64(int(typed)) == typed
	default:
		return 0, false
	}
}

func toCIntSlice(value any) ([]int, bool) {
	switch typed := value.(type) {
	case []int:
		return append([]int(nil), typed...), true
	case []any:
		values := make([]int, 0, len(typed))
		for _, item := range typed {
			next, ok := toCInt(item)
			if !ok {
				return nil, false
			}
			values = append(values, next)
		}
		return values, true
	default:
		return nil, false
	}
}
