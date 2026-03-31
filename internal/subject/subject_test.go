package subject

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveByFileNameMatchesNormalizedNames(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writeTestSubject(t, tempDir, Subject{
		ID:           "two-sum",
		Title:        "Two Sum",
		FileName:     "two_sum.py",
		FunctionName: "two_sum",
		Checker:      "two_sum_pair",
		TimeLimitMs:  1000,
		Tests: []TestCase{
			{Name: "basic_pair", Group: "core", Args: []any{[]int{2, 7}, 9}},
		},
	})

	summary, err := ResolveByFileName(tempDir, "two-sum.py")
	if err != nil {
		t.Fatalf("ResolveByFileName returned error: %v", err)
	}
	if summary.ID != "two-sum" {
		t.Fatalf("unexpected summary id: %q", summary.ID)
	}
}

func TestResolveByFileNameErrorsOnAmbiguousNormalizedNames(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writeTestSubject(t, tempDir, Subject{
		ID:           "alpha",
		Title:        "Alpha",
		FileName:     "alpha_beta.py",
		FunctionName: "alpha_beta",
		Checker:      "exact_int",
		TimeLimitMs:  1000,
		Tests: []TestCase{
			{Name: "basic", Group: "core", Args: []any{}, Expected: 1},
		},
	})
	writeTestSubject(t, tempDir, Subject{
		ID:           "beta",
		Title:        "Beta",
		FileName:     "alpha-beta.py",
		FunctionName: "alpha_beta",
		Checker:      "exact_int",
		TimeLimitMs:  1000,
		Tests: []TestCase{
			{Name: "basic", Group: "core", Args: []any{}, Expected: 1},
		},
	})

	_, err := ResolveByFileName(tempDir, "alpha beta.py")
	if err == nil {
		t.Fatalf("expected ambiguous match error")
	}
	if !strings.Contains(err.Error(), "matches multiple subjects") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeFileName(t *testing.T) {
	t.Parallel()

	if got := normalizeFileName(" Two-Sum.py "); got != "twosum" {
		t.Fatalf("unexpected normalized file name: %q", got)
	}
	if got := normalizeFileName("two_sum.py"); got != "twosum" {
		t.Fatalf("unexpected normalized file name: %q", got)
	}
}

func writeTestSubject(t *testing.T, subjectsDir string, spec Subject) {
	t.Helper()

	subjectDir := filepath.Join(subjectsDir, spec.ID)
	if err := os.MkdirAll(subjectDir, 0o755); err != nil {
		t.Fatalf("mkdir subject dir: %v", err)
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal subject: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subjectDir, "subject.json"), data, 0o644); err != nil {
		t.Fatalf("write subject.json: %v", err)
	}
}
