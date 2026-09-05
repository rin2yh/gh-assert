package runtime

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/rin2yh/gh-assert/internal/model"
)

const (
	scopeEnv   = "env"
	scopeInput = "input"
)

type section struct {
	scope   string
	subject string
}

var (
	envSection   = section{scope: scopeEnv, subject: "environment variable"}
	inputSection = section{scope: scopeInput, subject: "input"}
)

type source interface {
	lookup(name string) (string, bool)
}

type values map[string]string

func (v values) lookup(name string) (string, bool) { value, ok := v[name]; return value, ok }

type environment struct{}

func (environment) lookup(name string) (string, bool) { return os.LookupEnv(name) }

type Violation struct {
	Scope    string
	Name     string
	Position model.Position
	Message  string
}

func Assert(item model.ContractFile) ([]Violation, error) {
	event := os.Getenv("GITHUB_EVENT_NAME")
	forwarded := item.Kind == model.CompositeActionContract
	reusable := false
	if item.Kind == model.WorkflowContract {
		var err error
		reusable, err = isReusableWorkflow(item.SiblingPath)
		if err != nil {
			return nil, err
		}
		forwarded = reusable
	}
	envRules := mergeRules(item.Contract.Env, item.Contract.On[event].Env)
	inputEvent := event
	if reusable {
		inputEvent = "workflow_call"
	}
	inputRules := mergeRules(item.Contract.Inputs, item.Contract.On[inputEvent].Inputs)
	if len(inputRules) == 0 {
		return assert(envRules, inputRules, environment{}, values(nil)), nil
	}

	if forwarded {
		if os.Getenv(inputsJSON) == "" {
			return nil, fmt.Errorf("%s: inputs must be passed with inputs: ${{ toJSON(inputs) }}", item.Path)
		}
		inputs, err := loadForwardedInputs()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", item.Path, err)
		}
		return assert(envRules, inputRules, environment{}, values(inputs)), nil
	}

	if event == workflowDispatch {
		inputs, err := loadEventInputs()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", item.Path, err)
		}
		return assert(envRules, inputRules, environment{}, values(inputs)), nil
	}

	if event == "" {
		return nil, fmt.Errorf("%s: an inputs contract requires a %s run: GITHUB_EVENT_NAME is not set", item.Path, workflowDispatch)
	}
	return nil, fmt.Errorf("%s: an inputs contract requires a %s run, but the current event is %s", item.Path, workflowDispatch, event)
}

func mergeRules(common, scoped map[string]model.Rule) map[string]model.Rule {
	if len(scoped) == 0 {
		return common
	}
	merged := maps.Clone(common)
	if merged == nil {
		merged = make(map[string]model.Rule, len(scoped))
	}
	maps.Copy(merged, scoped)
	return merged
}

func assert(envRules, inputRules map[string]model.Rule, env, inputs source) []Violation {
	out := assertSection(envSection, envRules, env)
	return append(out, assertSection(inputSection, inputRules, inputs)...)
}

func assertSection(section section, rules map[string]model.Rule, source source) []Violation {
	var out []Violation
	for _, name := range slices.Sorted(maps.Keys(rules)) {
		rule := rules[name]
		value, present := source.lookup(name)
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
		out = append(out, assertValue(section.scope, name, rule, value)...)
	}
	return out
}

func assertValue(scope, name string, rule model.Rule, value string) []Violation {
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
