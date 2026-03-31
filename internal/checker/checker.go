package checker

import (
	"fmt"
	"reflect"
	"sort"

	"relay-judge/internal/subject"
)

type Result struct {
	Passed  bool
	Message string
}

type Checker func(test subject.TestCase, actual any) Result

var registry = map[string]Checker{
	"exact_bool":      exactBool,
	"exact_int":       exactInt,
	"exact_int_list":  exactIntList,
	"two_sum_pair":    twoSumPair,
	"intervals_exact": intervalsExact,
	"set_of_ints":     setOfInts,
}

func Evaluate(name string, test subject.TestCase, actual any) Result {
	fn, ok := registry[name]
	if !ok {
		return Result{
			Passed:  false,
			Message: fmt.Sprintf("unknown checker %q", name),
		}
	}

	return fn(test, actual)
}

func exactBool(test subject.TestCase, actual any) Result {
	expected, ok := test.Expected.(bool)
	if !ok {
		return Result{Passed: false, Message: "expected value must be a bool"}
	}

	value, ok := actual.(bool)
	if !ok {
		return Result{Passed: false, Message: fmt.Sprintf("expected bool return, got %T", actual)}
	}

	if value != expected {
		return Result{Passed: false, Message: fmt.Sprintf("expected %v, got %v", expected, value)}
	}

	return Result{Passed: true}
}

func exactInt(test subject.TestCase, actual any) Result {
	expected, ok := toInt(test.Expected)
	if !ok {
		return Result{Passed: false, Message: "expected value must be an int"}
	}

	value, ok := toInt(actual)
	if !ok {
		return Result{Passed: false, Message: fmt.Sprintf("expected int return, got %T", actual)}
	}

	if value != expected {
		return Result{Passed: false, Message: fmt.Sprintf("expected %d, got %d", expected, value)}
	}

	return Result{Passed: true}
}

func exactIntList(test subject.TestCase, actual any) Result {
	expected, ok := toIntSlice(test.Expected)
	if !ok {
		return Result{Passed: false, Message: "expected value must be a list of ints"}
	}

	value, ok := toIntSlice(actual)
	if !ok {
		return Result{Passed: false, Message: fmt.Sprintf("expected list[int] return, got %T", actual)}
	}

	if !reflect.DeepEqual(value, expected) {
		return Result{Passed: false, Message: fmt.Sprintf("expected %v, got %v", expected, value)}
	}

	return Result{Passed: true}
}

func twoSumPair(test subject.TestCase, actual any) Result {
	if len(test.Args) < 2 {
		return Result{Passed: false, Message: "two_sum_pair needs nums and target in args"}
	}

	nums, ok := toIntSlice(test.Args[0])
	if !ok {
		return Result{Passed: false, Message: "first arg must be a list of ints"}
	}

	target, ok := toInt(test.Args[1])
	if !ok {
		return Result{Passed: false, Message: "second arg must be an int target"}
	}

	indexes, ok := toIntSlice(actual)
	if !ok || len(indexes) != 2 {
		return Result{Passed: false, Message: "expected a list of exactly two indexes"}
	}

	i, j := indexes[0], indexes[1]
	if i == j {
		return Result{Passed: false, Message: "indexes must be distinct"}
	}
	if i < 0 || i >= len(nums) || j < 0 || j >= len(nums) {
		return Result{Passed: false, Message: "index out of bounds"}
	}
	if nums[i]+nums[j] != target {
		return Result{Passed: false, Message: fmt.Sprintf("nums[%d] + nums[%d] != %d", i, j, target)}
	}

	return Result{Passed: true}
}

func intervalsExact(test subject.TestCase, actual any) Result {
	expected, ok := toIntervals(test.Expected)
	if !ok {
		return Result{Passed: false, Message: "expected value must be a list of [start, end] pairs"}
	}

	value, ok := toIntervals(actual)
	if !ok {
		return Result{Passed: false, Message: fmt.Sprintf("expected interval list return, got %T", actual)}
	}

	if !reflect.DeepEqual(value, expected) {
		return Result{Passed: false, Message: fmt.Sprintf("expected %v, got %v", expected, value)}
	}

	return Result{Passed: true}
}

func setOfInts(test subject.TestCase, actual any) Result {
	expected, ok := toIntSlice(test.Expected)
	if !ok {
		return Result{Passed: false, Message: "expected value must be a list of ints"}
	}

	value, ok := toIntSlice(actual)
	if !ok {
		return Result{Passed: false, Message: fmt.Sprintf("expected list[int] return, got %T", actual)}
	}

	sort.Ints(expected)
	sort.Ints(value)

	if !reflect.DeepEqual(value, expected) {
		return Result{Passed: false, Message: fmt.Sprintf("expected set %v, got %v", expected, value)}
	}

	return Result{Passed: true}
}

func toInt(value any) (int, bool) {
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

func toIntSlice(value any) ([]int, bool) {
	items, ok := value.([]any)
	if ok {
		result := make([]int, 0, len(items))
		for _, item := range items {
			next, ok := toInt(item)
			if !ok {
				return nil, false
			}
			result = append(result, next)
		}
		return result, true
	}

	switch typed := value.(type) {
	case []int:
		return append([]int(nil), typed...), true
	default:
		return nil, false
	}
}

func toIntervals(value any) ([][]int, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}

	result := make([][]int, 0, len(items))
	for _, item := range items {
		pair, ok := toIntSlice(item)
		if !ok || len(pair) != 2 {
			return nil, false
		}
		result = append(result, pair)
	}

	return result, true
}
