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
	"sort"
	"strings"
	"time"

	"relay-judge/internal/checker"
	"relay-judge/internal/subject"
)

//go:embed wrapper.py
var wrapperSource string

type Options struct {
	PythonBin      string
	SubmissionPath string
}

type Report struct {
	SubjectID      string        `json:"subject_id"`
	SubjectTitle   string        `json:"subject_title"`
	SubmissionPath string        `json:"submission_path"`
	Status         string        `json:"status"`
	DurationMs     float64       `json:"duration_ms"`
	Message        string        `json:"message,omitempty"`
	Groups         []GroupReport `json:"groups"`
	Failures       []Failure     `json:"failures,omitempty"`
}

type GroupReport struct {
	Name     string `json:"name"`
	Total    int    `json:"total"`
	Executed int    `json:"executed"`
	Passed   int    `json:"passed"`
}

type Failure struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Message string `json:"message"`
}

type wrapperPayload struct {
	FunctionName string             `json:"function_name"`
	Tests        []subject.TestCase `json:"tests"`
}

type wrapperResponse struct {
	Status string              `json:"status"`
	Error  string              `json:"error,omitempty"`
	Tests  []wrapperTestResult `json:"tests,omitempty"`
}

type wrapperTestResult struct {
	Name       string  `json:"name"`
	Group      string  `json:"group"`
	Status     string  `json:"status"`
	Actual     any     `json:"actual,omitempty"`
	Error      string  `json:"error,omitempty"`
	DurationMs float64 `json:"duration_ms,omitempty"`
	Stdout     string  `json:"stdout,omitempty"`
	Stderr     string  `json:"stderr,omitempty"`
}

func Run(spec subject.Subject, options Options) (Report, error) {
	startedAt := time.Now()
	report := Report{
		SubjectID:      spec.ID,
		SubjectTitle:   spec.Title,
		SubmissionPath: options.SubmissionPath,
		Groups:         buildGroupReports(spec.Tests),
	}

	if strings.TrimSpace(options.SubmissionPath) == "" {
		report.Status = "load_error"
		report.Message = "submission path is required"
		report.DurationMs = time.Since(startedAt).Seconds() * 1000
		return report, nil
	}

	if _, err := os.Stat(options.SubmissionPath); err != nil {
		report.Status = "load_error"
		report.Message = fmt.Sprintf("submission file not found: %s", options.SubmissionPath)
		report.DurationMs = time.Since(startedAt).Seconds() * 1000
		return report, nil
	}

	tempDir, err := os.MkdirTemp("", "relay-judge-*")
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

	groupIndex := indexGroupReports(report.Groups)
	var hasFailure bool

	for index, testResult := range wrapped.Tests {
		if index >= len(spec.Tests) {
			break
		}

		test := spec.Tests[index]
		group := report.Groups[groupIndex[test.Group]]
		group.Executed++
		report.Groups[groupIndex[test.Group]] = group

		if testResult.Status == "runtime_error" {
			report.Status = "runtime_error"
			report.Message = "submission raised an exception"
			report.Failures = append(report.Failures, Failure{
				Name:    test.Name,
				Group:   test.Group,
				Message: compactMessage(testResult.Error),
			})
			return report, nil
		}

		result := checker.Evaluate(spec.Checker, test, testResult.Actual)
		if result.Passed {
			group = report.Groups[groupIndex[test.Group]]
			group.Passed++
			report.Groups[groupIndex[test.Group]] = group
			continue
		}

		hasFailure = true
		report.Failures = append(report.Failures, Failure{
			Name:    test.Name,
			Group:   test.Group,
			Message: result.Message,
		})
	}

	if hasFailure {
		report.Status = "failed"
		return report, nil
	}

	report.Status = "passed"
	return report, nil
}

func buildGroupReports(tests []subject.TestCase) []GroupReport {
	counts := map[string]int{}
	for _, test := range tests {
		counts[test.Group]++
	}

	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return groupPriority(names[i]) < groupPriority(names[j])
	})

	groups := make([]GroupReport, 0, len(names))
	for _, name := range names {
		groups = append(groups, GroupReport{
			Name:  name,
			Total: counts[name],
		})
	}

	return groups
}

func groupPriority(name string) int {
	switch name {
	case "core":
		return 0
	case "edge":
		return 1
	case "anti-hardcode":
		return 2
	case "perf":
		return 3
	default:
		return 100
	}
}

func indexGroupReports(groups []GroupReport) map[string]int {
	index := make(map[string]int, len(groups))
	for i, group := range groups {
		index[group.Name] = i
	}
	return index
}

func joinNonEmpty(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return strings.Join(filtered, "\n")
}

func compactMessage(message string) string {
	lines := strings.Split(strings.TrimSpace(message), "\n")
	if len(lines) == 0 {
		return "unknown error"
	}
	return lines[len(lines)-1]
}
