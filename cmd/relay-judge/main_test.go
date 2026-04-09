package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDirectArgsParsesStressMode(t *testing.T) {
	t.Parallel()

	submissionPath, subjectsDir, pythonBin, cCompiler, jsonOutput, detailedOutput, stressMode, handled, err := parseDirectArgs([]string{
		"--stress",
		"--json",
		"--python",
		"python-custom",
		"--cc",
		"clang",
		"/tmp/two-sum.py",
	})
	if err != nil {
		t.Fatalf("parseDirectArgs returned error: %v", err)
	}
	if !handled {
		t.Fatalf("expected handled=true")
	}
	if submissionPath != "/tmp/two-sum.py" {
		t.Fatalf("unexpected submission path: %q", submissionPath)
	}
	if subjectsDir == "" {
		t.Fatalf("expected non-empty subjects dir")
	}
	if pythonBin != "python-custom" {
		t.Fatalf("unexpected python bin: %q", pythonBin)
	}
	if cCompiler != "clang" {
		t.Fatalf("unexpected c compiler: %q", cCompiler)
	}
	if !jsonOutput {
		t.Fatalf("expected jsonOutput=true")
	}
	if detailedOutput {
		t.Fatalf("expected detailedOutput=false")
	}
	if !stressMode {
		t.Fatalf("expected stressMode=true")
	}
}

func TestParseDirectArgsParsesCSubmission(t *testing.T) {
	t.Parallel()

	submissionPath, subjectsDir, pythonBin, cCompiler, jsonOutput, detailedOutput, stressMode, handled, err := parseDirectArgs([]string{
		"--cc",
		"clang-custom",
		"/tmp/sort_the_stack.c",
	})
	if err != nil {
		t.Fatalf("parseDirectArgs returned error: %v", err)
	}
	if !handled {
		t.Fatalf("expected handled=true")
	}
	if submissionPath != "/tmp/sort_the_stack.c" {
		t.Fatalf("unexpected submission path: %q", submissionPath)
	}
	if subjectsDir == "" {
		t.Fatalf("expected non-empty subjects dir")
	}
	if pythonBin != defaultPython {
		t.Fatalf("unexpected python bin: %q", pythonBin)
	}
	if cCompiler != "clang-custom" {
		t.Fatalf("unexpected c compiler: %q", cCompiler)
	}
	if jsonOutput || detailedOutput || stressMode {
		t.Fatalf("expected optional output flags to remain false")
	}
}

func TestApplyTrailingRunFlagsSetsStressMode(t *testing.T) {
	t.Parallel()

	var jsonOutput bool
	var detailedOutput bool
	var stressMode bool

	err := applyTrailingRunFlags([]string{"./examples/two_sum.py", "--stress", "--json"}, &jsonOutput, &detailedOutput, &stressMode)
	if err != nil {
		t.Fatalf("applyTrailingRunFlags returned error: %v", err)
	}
	if !jsonOutput {
		t.Fatalf("expected jsonOutput=true")
	}
	if detailedOutput {
		t.Fatalf("expected detailedOutput=false")
	}
	if !stressMode {
		t.Fatalf("expected stressMode=true")
	}
}

func TestRunWithInferredSubjectStressModeSupportsHyphenatedFileNames(t *testing.T) {
	tempDir := t.TempDir()
	subjectsDir := filepath.Join(tempDir, "subjects")
	subjectDir := filepath.Join(subjectsDir, "two-sum")
	if err := os.MkdirAll(subjectDir, 0o755); err != nil {
		t.Fatalf("mkdir subject dir: %v", err)
	}

	subjectJSON := `{
  "id": "two-sum",
  "title": "Two Sum",
  "file_name": "two_sum.py",
  "function_name": "two_sum",
  "checker": "two_sum_pair",
  "time_limit_ms": 1500,
  "tests": [
    {
      "name": "basic_pair",
      "group": "core",
      "args": [[2, 7, 11, 15], 9]
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(subjectDir, "subject.json"), []byte(subjectJSON), 0o644); err != nil {
		t.Fatalf("write subject.json: %v", err)
	}

	submissionPath := filepath.Join(tempDir, "two-sum.py")
	submissionSource := "def two_sum(nums, target):\n    seen = {}\n    for index, value in enumerate(nums):\n        wanted = target - value\n        if wanted in seen:\n            return [seen[wanted], index]\n        seen[value] = index\n    return []\n"
	if err := os.WriteFile(submissionPath, []byte(submissionSource), 0o644); err != nil {
		t.Fatalf("write submission: %v", err)
	}

	stdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = writer

	code, runErr := runWithInferredSubject(subjectsDir, submissionPath, defaultPython, defaultC, true, false, true)

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = stdout

	var output bytes.Buffer
	if _, err := output.ReadFrom(reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if runErr != nil {
		t.Fatalf("runWithInferredSubject returned error: %v", runErr)
	}
	if code != exitPassed {
		t.Fatalf("unexpected exit code: got %d want %d", code, exitPassed)
	}
	if !strings.Contains(output.String(), `"mode": "stress"`) {
		t.Fatalf("expected stress json output, got: %s", output.String())
	}
	if !strings.Contains(output.String(), `"subject_id": "two-sum"`) {
		t.Fatalf("expected subject id in output, got: %s", output.String())
	}
}
