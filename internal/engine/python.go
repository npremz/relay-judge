package engine

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"relay-judge/internal/subject"
)

//go:embed wrapper.py
var wrapperSource string

func runPython(spec subject.Subject, options Options, report Report, startedAt time.Time) (Report, error) {
	tempDir, err := os.MkdirTemp("", "relay-judge-python-*")
	if err != nil {
		return report, err
	}
	defer os.RemoveAll(tempDir)

	wrapperPath := filepath.Join(tempDir, "wrapper.py")
	if err := os.WriteFile(wrapperPath, []byte(wrapperSource), 0o700); err != nil {
		return report, err
	}

	payload, err := json.Marshal(wrapperPayload{
		FunctionName: spec.FunctionName,
		Tests:        spec.Tests,
	})
	if err != nil {
		return report, err
	}

	timeout := time.Duration(spec.TimeLimitMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pythonBin := options.PythonBin
	if strings.TrimSpace(pythonBin) == "" {
		pythonBin = "python3"
	}

	cmd := exec.CommandContext(ctx, pythonBin, wrapperPath, options.SubmissionPath)
	cmd.Stdin = bytes.NewReader(payload)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	report.DurationMs = time.Since(startedAt).Seconds() * 1000

	if ctx.Err() == context.DeadlineExceeded {
		report.Status = "timeout"
		report.Message = fmt.Sprintf("python process exceeded %dms", spec.TimeLimitMs)
		return report, nil
	}

	if err != nil {
		report.Status = "load_error"
		report.Message = strings.TrimSpace(joinNonEmpty(stderr.String(), err.Error()))
		return report, nil
	}

	var wrapped wrapperResponse
	if err := json.Unmarshal(stdout.Bytes(), &wrapped); err != nil {
		report.Status = "load_error"
		report.Message = fmt.Sprintf("invalid wrapper output: %v", err)
		if stderr.Len() > 0 {
			report.Message = joinNonEmpty(report.Message, stderr.String())
		}
		return report, nil
	}

	if wrapped.Status != "ok" {
		report.Status = "load_error"
		report.Message = wrapped.Error
		return report, nil
	}

	report = evaluateTestResults(spec, report, wrapped.Tests)
	return report, nil
}
