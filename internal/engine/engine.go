package engine

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"relay-judge/internal/checker"
	"relay-judge/internal/subject"
)

type Options struct {
	PythonBin      string
	CCompiler      string
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
	report := newReport(spec, options.SubmissionPath)

	if strings.TrimSpace(options.SubmissionPath) == "" {
		markLoadError(&report, startedAt, "submission path is required")
		return report, nil
	}

	if _, err := os.Stat(options.SubmissionPath); err != nil {
		markLoadError(&report, startedAt, fmt.Sprintf("submission file not found: %s", options.SubmissionPath))
		return report, nil
	}

	switch spec.NormalizedLanguage() {
	case "python":
		return runPython(spec, options, report, startedAt)
	case "c":
		return runC(spec, options, report, startedAt)
	default:
		markLoadError(&report, startedAt, fmt.Sprintf("unsupported language %q", spec.NormalizedLanguage()))
		return report, nil
	}
}

func newReport(spec subject.Subject, submissionPath string) Report {
	return Report{
		SubjectID:      spec.ID,
		SubjectTitle:   spec.Title,
		SubmissionPath: submissionPath,
		Groups:         buildGroupReports(spec.Tests),
	}
}

func markLoadError(report *Report, startedAt time.Time, message string) {
	report.Status = "load_error"
	report.Message = message
	report.DurationMs = time.Since(startedAt).Seconds() * 1000
}

func evaluateTestResults(spec subject.Subject, report Report, results []wrapperTestResult) Report {
	groupIndex := indexGroupReports(report.Groups)
	var hasFailure bool

	for index, testResult := range results {
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
			return report
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
		return report
	}

	report.Status = "passed"
	return report
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
