package subject

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var normalizedFileNamePattern = regexp.MustCompile(`[^a-z0-9]+`)

type Subject struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Language     string     `json:"language,omitempty"`
	Prototype    string     `json:"prototype,omitempty"`
	Description  string     `json:"description,omitempty"`
	FileName     string     `json:"file_name"`
	FunctionName string     `json:"function_name"`
	ResultSource string     `json:"result_source,omitempty"`
	Checker      string     `json:"checker"`
	TimeLimitMs  int        `json:"time_limit_ms"`
	Tests        []TestCase `json:"tests"`
}

type TestCase struct {
	Name     string `json:"name"`
	Group    string `json:"group"`
	Args     []any  `json:"args"`
	Expected any    `json:"expected,omitempty"`
}

type Summary struct {
	ID       string
	Title    string
	FileName string
	Path     string
}

func Load(path string) (Subject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Subject{}, err
	}

	var spec Subject
	if err := json.Unmarshal(data, &spec); err != nil {
		return Subject{}, fmt.Errorf("parse %s: %w", path, err)
	}

	if err := spec.Validate(); err != nil {
		return Subject{}, fmt.Errorf("invalid subject %s: %w", path, err)
	}

	return spec, nil
}

func (s Subject) Validate() error {
	switch {
	case strings.TrimSpace(s.ID) == "":
		return fmt.Errorf("missing id")
	case strings.TrimSpace(s.Title) == "":
		return fmt.Errorf("missing title")
	case strings.TrimSpace(s.FileName) == "":
		return fmt.Errorf("missing file_name")
	case strings.TrimSpace(s.FunctionName) == "":
		return fmt.Errorf("missing function_name")
	case strings.TrimSpace(s.Checker) == "":
		return fmt.Errorf("missing checker")
	case s.TimeLimitMs <= 0:
		return fmt.Errorf("time_limit_ms must be > 0")
	case len(s.Tests) == 0:
		return fmt.Errorf("at least one test is required")
	}

	if !isValidLanguage(s.NormalizedLanguage()) {
		return fmt.Errorf("unsupported language %q", s.NormalizedLanguage())
	}

	if s.NormalizedLanguage() == "c" && strings.TrimSpace(s.Prototype) == "" {
		return fmt.Errorf("prototype is required for c subjects")
	}

	if !isValidResultSource(s.NormalizedResultSource()) {
		return fmt.Errorf("unsupported result_source %q", strings.TrimSpace(s.ResultSource))
	}

	for index, test := range s.Tests {
		if strings.TrimSpace(test.Name) == "" {
			return fmt.Errorf("test %d: missing name", index)
		}
		if strings.TrimSpace(test.Group) == "" {
			return fmt.Errorf("test %s: missing group", test.Name)
		}
		if !isValidGroup(test.Group) {
			return fmt.Errorf("test %s: invalid group %q", test.Name, test.Group)
		}
	}

	return nil
}

func (s Subject) NormalizedLanguage() string {
	language := strings.TrimSpace(strings.ToLower(s.Language))
	if language != "" {
		return language
	}

	switch strings.ToLower(filepath.Ext(strings.TrimSpace(s.FileName))) {
	case ".c":
		return "c"
	case ".py":
		return "python"
	default:
		return "python"
	}
}

func (s Subject) NormalizedResultSource() string {
	source := strings.TrimSpace(strings.ToLower(s.ResultSource))
	if source != "" {
		return source
	}

	if s.NormalizedLanguage() == "c" && strings.HasPrefix(strings.TrimSpace(s.Prototype), "void ") {
		switch s.Checker {
		case "exact_array", "exact_int_list", "set_of_ints":
			return "stdout_ints"
		}
	}

	return "return"
}

func isValidGroup(group string) bool {
	switch group {
	case "core", "edge", "anti-hardcode", "perf":
		return true
	default:
		return false
	}
}

func isValidLanguage(language string) bool {
	switch language {
	case "python", "c":
		return true
	default:
		return false
	}
}

func isValidResultSource(source string) bool {
	switch source {
	case "return", "stdout_ints":
		return true
	default:
		return false
	}
}

func Discover(dir string) ([]Summary, error) {
	var subjects []Summary

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || entry.Name() != "subject.json" {
			return nil
		}

		spec, err := Load(path)
		if err != nil {
			return err
		}

		subjects = append(subjects, Summary{
			ID:       spec.ID,
			Title:    spec.Title,
			FileName: spec.FileName,
			Path:     path,
		})

		return nil
	})

	return subjects, err
}

func ResolvePath(subjectsDir, subjectArg string) (string, error) {
	if strings.HasSuffix(subjectArg, ".json") {
		return subjectArg, nil
	}

	candidate := filepath.Join(subjectsDir, subjectArg, "subject.json")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf("subject %q not found in %s", subjectArg, subjectsDir)
}

func ResolveByFileName(subjectsDir, fileName string) (Summary, error) {
	items, err := Discover(subjectsDir)
	if err != nil {
		return Summary{}, err
	}

	normalizedFileName := strings.TrimSpace(filepath.Base(fileName))
	if normalizedFileName == "" {
		return Summary{}, fmt.Errorf("empty submission filename")
	}

	normalizedKey := normalizeFileName(normalizedFileName)
	var matches []Summary
	for _, item := range items {
		if item.FileName == normalizedFileName || normalizeFileName(item.FileName) == normalizedKey {
			matches = append(matches, item)
		}
	}

	switch len(matches) {
	case 0:
		return Summary{}, fmt.Errorf("no subject matches filename %q", normalizedFileName)
	case 1:
		return matches[0], nil
	default:
		ids := make([]string, 0, len(matches))
		for _, match := range matches {
			ids = append(ids, match.ID)
		}
		sort.Strings(ids)
		return Summary{}, fmt.Errorf("filename %q matches multiple subjects: %s", normalizedFileName, strings.Join(ids, ", "))
	}
}

func normalizeFileName(fileName string) string {
	base := strings.TrimSpace(filepath.Base(fileName))
	stem := strings.TrimSuffix(strings.ToLower(base), filepath.Ext(base))
	return normalizedFileNamePattern.ReplaceAllString(stem, "")
}
