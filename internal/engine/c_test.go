package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"relay-judge/internal/subject"
)

func TestRunCSupportsStdoutIntArrays(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath(defaultCCompiler); err != nil {
		t.Skipf("skip: %s not available: %v", defaultCCompiler, err)
	}

	tempDir := t.TempDir()
	submissionPath := filepath.Join(tempDir, "sort_the_stack.c")
	submissionSource := `#include <stdio.h>
#include <stdlib.h>

void sort_the_stack(int *stack1, int len_stack1, int *stack2, int len_stack2) {
    int total = len_stack1 + len_stack2;
    int *values = malloc((size_t) total * sizeof(int));
    int index = 0;

    for (int i = 0; i < len_stack1; i++) {
        values[index++] = stack1[i];
    }
    for (int i = 0; i < len_stack2; i++) {
        values[index++] = stack2[i];
    }

    for (int i = 0; i < total; i++) {
        for (int j = i + 1; j < total; j++) {
            if (values[j] < values[i]) {
                int tmp = values[i];
                values[i] = values[j];
                values[j] = tmp;
            }
        }
    }

    for (int i = 0; i < total; i++) {
        if (i > 0) {
            printf(" ");
        }
        printf("%d", values[i]);
    }

    free(values);
}
`
	if err := os.WriteFile(submissionPath, []byte(submissionSource), 0o644); err != nil {
		t.Fatalf("write submission: %v", err)
	}

	spec := subject.Subject{
		ID:           "sort-the-stack",
		Title:        "Sort the stack",
		Language:     "c",
		Prototype:    "void sort_the_stack(int *stack1, int len_stack1, int *stack2, int len_stack2):",
		FileName:     "sort_the_stack.c",
		FunctionName: "sort_the_stack",
		Checker:      "exact_array",
		TimeLimitMs:  1200,
		Tests: []subject.TestCase{
			{
				Name:     "test1",
				Group:    "core",
				Args:     []any{[]int{1, 3, 5}, 3, []int{2, 4, 6}, 3},
				Expected: []int{1, 2, 3, 4, 5, 6},
			},
			{
				Name:     "test2",
				Group:    "core",
				Args:     []any{[]int{12, 4, 3}, 3, []int{6, 90, 8}, 3},
				Expected: []int{3, 4, 6, 8, 12, 90},
			},
		},
	}

	if err := spec.Validate(); err != nil {
		t.Fatalf("validate subject: %v", err)
	}

	report, err := Run(spec, Options{
		CCompiler:      defaultCCompiler,
		SubmissionPath: submissionPath,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if report.Status != "passed" {
		t.Fatalf("unexpected status: got %q, message=%q failures=%v", report.Status, report.Message, report.Failures)
	}
	if len(report.Groups) != 1 || report.Groups[0].Passed != 2 {
		t.Fatalf("unexpected group results: %+v", report.Groups)
	}
}
