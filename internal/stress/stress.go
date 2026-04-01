package stress

import (
	"fmt"
	"strings"

	"relay-judge/internal/subject"
)

const stressTimeLimitMs = 5000

func Build(base subject.Subject) (subject.Subject, error) {
	spec := base
	spec.TimeLimitMs = max(base.TimeLimitMs, stressTimeLimitMs)

	tests, err := testsFor(base.ID)
	if err != nil {
		return subject.Subject{}, err
	}

	spec.Tests = tests
	return spec, nil
}

func testsFor(subjectID string) ([]subject.TestCase, error) {
	switch subjectID {
	case "two-sum":
		return []subject.TestCase{
			{
				Name:  "stress_solution_at_tail",
				Group: "perf",
				Args:  []any{buildTwoSumRamp(100000, 4), 4},
			},
			{
				Name:  "stress_negative_pair_at_tail",
				Group: "perf",
				Args:  []any{buildTwoSumRampWithNegativePair(80000), -1},
			},
		}, nil
	case "minimum-size-subarray-sum":
		return []subject.TestCase{
			{
				Name:     "stress_long_ones_window",
				Group:    "perf",
				Args:     []any{100000, buildRepeatedIntSlice(200000, 1)},
				Expected: 100000,
			},
			{
				Name:     "stress_long_prefix_then_spike",
				Group:    "perf",
				Args:     []any{50000, append(buildRepeatedIntSlice(120000, 1), 50000)},
				Expected: 1,
			},
		}, nil
	case "first-unique-character":
		return []subject.TestCase{
			{
				Name:     "stress_unique_last_char",
				Group:    "perf",
				Args:     []any{strings.Repeat("a", 200000) + "b"},
				Expected: 200000,
			},
			{
				Name:     "stress_no_unique_character",
				Group:    "perf",
				Args:     []any{strings.Repeat("ab", 120000)},
				Expected: -1,
			},
		}, nil
	case "longest-substring-without-repeating-characters":
		uniqueRun := uniqueRuneString(8192)
		return []subject.TestCase{
			{
				Name:     "stress_unique_window_then_repeat",
				Group:    "perf",
				Args:     []any{uniqueRun + string([]rune(uniqueRun)[0])},
				Expected: 8192,
			},
			{
				Name:     "stress_repeated_blocks",
				Group:    "perf",
				Args:     []any{strings.Repeat(firstNRunes(uniqueRun, 2048), 32)},
				Expected: 2048,
			},
		}, nil
	case "merge-intervals":
		return []subject.TestCase{
			{
				Name:     "stress_reverse_chain",
				Group:    "perf",
				Args:     []any{buildReverseChainIntervals(50000)},
				Expected: [][]int{{0, 50000}},
			},
			{
				Name:     "stress_nested_then_disjoint",
				Group:    "perf",
				Args:     []any{buildNestedIntervalsWithTail()},
				Expected: [][]int{{0, 100000}, {100002, 100010}},
			},
		}, nil
	case "top-k-frequent-elements":
		nums, expected := buildTopKStressCase(512, 1024, 10)
		return []subject.TestCase{
			{
				Name:     "stress_descending_frequencies",
				Group:    "perf",
				Args:     []any{nums, 10},
				Expected: expected,
			},
		}, nil
	case "valid-parentheses":
		return []subject.TestCase{
			{
				Name:     "stress_deep_balanced_stack",
				Group:    "perf",
				Args:     []any{strings.Repeat("(", 100000) + strings.Repeat(")", 100000)},
				Expected: true,
			},
			{
				Name:     "stress_late_mismatch",
				Group:    "perf",
				Args:     []any{strings.Repeat("(", 90000) + strings.Repeat(")", 89999) + "]"},
				Expected: false,
			},
		}, nil
	default:
		return nil, fmt.Errorf("no stress suite defined for subject %q", subjectID)
	}
}

func buildRepeatedIntSlice(length, value int) []int {
	items := make([]int, length)
	for index := range items {
		items[index] = value
	}
	return items
}

func buildTwoSumRamp(length, step int) []int {
	items := make([]int, 0, length+2)
	for index := 0; index < length; index++ {
		items = append(items, index*step+10)
	}
	items = append(items, 1, 3)
	return items
}

func buildTwoSumRampWithNegativePair(length int) []int {
	items := make([]int, 0, length+2)
	for index := 0; index < length; index++ {
		items = append(items, index*6+12)
	}
	items = append(items, -5, 4)
	return items
}

func uniqueRuneString(length int) string {
	runes := make([]rune, 0, length)
	for codepoint := rune(0x0100); len(runes) < length; codepoint++ {
		if codepoint >= 0xD800 && codepoint <= 0xDFFF {
			codepoint = 0xE000
		}
		runes = append(runes, codepoint)
	}
	return string(runes)
}

func firstNRunes(value string, length int) string {
	runes := []rune(value)
	if len(runes) < length {
		length = len(runes)
	}
	return string(runes[:length])
}

func buildReverseChainIntervals(count int) [][]int {
	intervals := make([][]int, 0, count)
	for index := count - 1; index >= 0; index-- {
		intervals = append(intervals, []int{index, index + 1})
	}
	return intervals
}

func buildNestedIntervalsWithTail() [][]int {
	intervals := make([][]int, 0, 4002)
	for index := 0; index <= 2000; index++ {
		intervals = append(intervals, []int{index, 100000 - index})
	}
	for index := 2000; index >= 0; index-- {
		intervals = append(intervals, []int{index, 100000 - index})
	}
	intervals = append(intervals, []int{100002, 100005}, []int{100004, 100010})
	return intervals
}

func buildTopKStressCase(distinct, maxFrequency, topK int) ([]int, []int) {
	nums := make([]int, 0, (maxFrequency*distinct)-((distinct-1)*distinct)/2)
	expected := make([]int, 0, topK)

	for value := 0; value < distinct; value++ {
		if value < topK {
			expected = append(expected, value)
		}
		for count := maxFrequency - value; count > 0; count-- {
			nums = append(nums, value)
		}
	}

	return nums, expected
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
