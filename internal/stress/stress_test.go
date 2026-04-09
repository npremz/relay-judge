package stress

import (
	"testing"

	"relay-judge/internal/subject"
)

func TestBuildSortTheStackStressSuite(t *testing.T) {
	t.Parallel()

	spec, err := Build(subject.Subject{
		ID:           "sort-the-stack",
		Title:        "Sort the stack",
		Language:     "c",
		Prototype:    "void sort_the_stack(int *stack1, int len_stack1, int *stack2, int len_stack2):",
		FileName:     "sort_the_stack.c",
		FunctionName: "sort_the_stack",
		ResultSource: "stdout_ints",
		Checker:      "exact_array",
		TimeLimitMs:  1200,
		Tests: []subject.TestCase{
			{Name: "placeholder", Group: "core", Args: []any{[]int{1}, 1, []int{2}, 1}, Expected: []int{1, 2}},
		},
	})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if spec.TimeLimitMs != stressTimeLimitMs {
		t.Fatalf("unexpected stress time limit: got %d want %d", spec.TimeLimitMs, stressTimeLimitMs)
	}
	if len(spec.Tests) != 2 {
		t.Fatalf("unexpected stress test count: %d", len(spec.Tests))
	}

	for _, test := range spec.Tests {
		if test.Group != "perf" {
			t.Fatalf("expected perf group, got %q", test.Group)
		}
		if len(test.Args) != 4 {
			t.Fatalf("expected 4 args, got %d", len(test.Args))
		}
		expected, ok := test.Expected.([]int)
		if !ok {
			t.Fatalf("expected []int payload, got %T", test.Expected)
		}
		if len(expected) == 0 {
			t.Fatalf("expected non-empty expected array")
		}
	}
}
