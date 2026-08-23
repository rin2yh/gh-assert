package assertion

import (
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/rin2yh/gh-assert/internal/parser"
)

const (
	ScopeEnv   = "env"
	ScopeInput = "input"
)

type section struct {
	scope   string
	subject string
}

var (
	envSection   = section{scope: ScopeEnv, subject: "environment variable"}
	inputSection = section{scope: ScopeInput, subject: "input"}
)

type source interface {
	Lookup(name string) (string, bool)
}

type values map[string]string

func (v values) Lookup(name string) (string, bool) { value, ok := v[name]; return value, ok }

type osEnvironment struct{}

func (osEnvironment) Lookup(name string) (string, bool) { return os.LookupEnv(name) }

type Violation struct {
	Scope    string
	Name     string
	Position parser.Position
	Message  string
}

func ValidateRuntime(c *parser.Contract, inputs map[string]string) []Violation {
	return Validate(c, osEnvironment{}, values(inputs))
}

func Validate(c *parser.Contract, env, inputs source) []Violation {
	out := validateSection(envSection, c.Env, env)
	return append(out, validateSection(inputSection, c.Inputs, inputs)...)
}

func validateSection(section section, rules map[string]parser.Rule, source source) []Violation {
	var out []Violation
	for _, name := range slices.Sorted(maps.Keys(rules)) {
		rule := rules[name]
		value, present := source.Lookup(name)
		if !present {
			if rule.Required {
				out = append(out, Violation{section.scope, name, rule.Position, "required " + section.subject + " is not set"})
			}
			continue
		}
		if rule.Required && value == "" {
			out = append(out, Violation{section.scope, name, rule.Position, "required " + section.subject + " is empty"})
			continue
		}
		out = append(out, validateValue(section.scope, name, rule, value)...)
	}
	return out
}

func validateValue(scope, name string, rule parser.Rule, value string) []Violation {
	var out []Violation
	switch rule.Type.Kind {
	case "string":
		if rule.Type.String != nil {
			if len(rule.Type.String.Enum) > 0 && !slices.Contains(rule.Type.String.Enum, value) {
				out = append(out, Violation{scope, name, rule.Position, "value is not in the allowed enum"})
			}
			if rule.Type.String.Pattern != nil && !rule.Type.String.Pattern.MatchString(value) {
				out = append(out, Violation{scope, name, rule.Position, "value does not match the pattern"})
			}
		}
	case "integer":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return append(out, Violation{scope, name, rule.Position, "value is not an integer"})
		}
		if rule.Type.Integer.Min != nil && v < *rule.Type.Integer.Min {
			out = append(out, Violation{scope, name, rule.Position, "value is below the minimum"})
		}
		if rule.Type.Integer.Max != nil && v > *rule.Type.Integer.Max {
			out = append(out, Violation{scope, name, rule.Position, "value is above the maximum"})
		}
	case "boolean":
		if !strings.EqualFold(value, "true") && !strings.EqualFold(value, "false") {
			out = append(out, Violation{scope, name, rule.Position, "value is not a boolean"})
		}
	}
	return out
}
