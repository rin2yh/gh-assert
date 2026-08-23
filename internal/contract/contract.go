package contract

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"regexp/syntax"
	"slices"
	"strings"

	"github.com/rin2yh/gh-assert/internal/model"
)

func LoadFile(path string) (*model.Contract, error) {
	parsed, err := NewParser(path).Parse()
	if err != nil {
		return nil, err
	}
	if err := Validate(path, parsed); err != nil {
		return nil, err
	}
	Compile(parsed)
	return parsed, nil
}

func LoadTargets(path string) ([]model.ContractFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		parsed, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		return []model.ContractFile{{Path: path, Contract: parsed}}, nil
	}

	var paths []string
	err = filepath.WalkDir(path, func(current string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_assert.yml") {
			paths = append(paths, current)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no contract files matching *_assert.yml found in %s", path)
	}

	loaded := make([]model.ContractFile, 0, len(paths))
	for _, path := range paths {
		parsed, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, model.ContractFile{Path: path, Contract: parsed})
	}
	return loaded, nil
}

// Validate checks a parsed contract without preparing it for runtime use.
func Validate(path string, parsed *model.Contract) error {
	if parsed.Env == nil && parsed.Inputs == nil {
		return fmt.Errorf("%s: env or inputs is required", path)
	}
	if err := validateSection("env", parsed.Env); err != nil {
		return err
	}
	return validateSection("inputs", parsed.Inputs)
}

func validateSection(section string, rules map[string]model.Rule) error {
	for _, name := range slices.Sorted(maps.Keys(rules)) {
		rule := rules[name]
		position := rule.TypePosition
		if position.Path == "" {
			position = rule.Position
		}
		if err := validateRule(section, name, rule, position); err != nil {
			return err
		}
	}
	return nil
}

func validateRule(section, name string, rule model.Rule, position model.Position) error {
	types := 0
	if rule.Type.String != nil {
		types++
		if rule.Type.String.PatternText != "" {
			if _, err := syntax.Parse(rule.Type.String.PatternText, syntax.Perl); err != nil {
				return fmt.Errorf("%s:%d:%d: invalid pattern: %w", position.Path, position.Line, position.Column, err)
			}
		}
	}
	if rule.Type.Integer != nil {
		types++
		integer := rule.Type.Integer
		if integer.Min != nil && integer.Max != nil && *integer.Min > *integer.Max {
			return fmt.Errorf("%s:%d:%d: min must not be greater than max", position.Path, position.Line, position.Column)
		}
	}
	if rule.Type.Boolean != nil {
		types++
	}
	if types != 1 {
		return fmt.Errorf("%s:%d:%d: %s.%s.type must specify exactly one type", position.Path, position.Line, position.Column, section, name)
	}
	return nil
}

// Compile prepares a validated contract for runtime use.
func Compile(parsed *model.Contract) {
	for _, rules := range []map[string]model.Rule{parsed.Env, parsed.Inputs} {
		for name, rule := range rules {
			switch {
			case rule.Type.String != nil:
				rule.Type.Kind = "string"
				if rule.Type.String.PatternText != "" {
					rule.Type.String.Pattern = regexp.MustCompile(rule.Type.String.PatternText)
				}
			case rule.Type.Integer != nil:
				rule.Type.Kind = "integer"
			case rule.Type.Boolean != nil:
				rule.Type.Kind = "boolean"
			}
			rules[name] = rule
		}
	}
}
