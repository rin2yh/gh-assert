package assertion

import (
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/rin2yh/gh-assert/internal/contract"
)

type source interface {
	Lookup(name string) (string, bool)
}

type environment map[string]string

func (e environment) Lookup(name string) (string, bool) { v, ok := e[name]; return v, ok }

type osEnvironment struct{}

func (osEnvironment) Lookup(name string) (string, bool) { return os.LookupEnv(name) }

func ValidateOS(c *contract.Contract) []Violation { return Validate(c, osEnvironment{}) }

type Violation struct {
	Env      string
	Position contract.Position
	Message  string
}

func Validate(c *contract.Contract, source source) []Violation {
	names := make([]string, 0, len(c.Env))
	for name := range c.Env {
		names = append(names, name)
	}
	sort.Strings(names)
	var out []Violation
	for _, name := range names {
		rule := c.Env[name]
		value, present := source.Lookup(name)
		if !present {
			if rule.Required {
				out = append(out, Violation{name, rule.Position, "required environment variable is not set"})
			}
			continue
		}
		if rule.Required && value == "" {
			out = append(out, Violation{name, rule.Position, "required environment variable is empty"})
			continue
		}
		switch rule.Type.Kind {
		case "string":
			if rule.Type.String != nil {
				if len(rule.Type.String.Enum) > 0 && !slices.Contains(rule.Type.String.Enum, value) {
					out = append(out, Violation{name, rule.Position, "value is not in the allowed enum"})
				}
				if rule.Type.String.Pattern != nil && !rule.Type.String.Pattern.MatchString(value) {
					out = append(out, Violation{name, rule.Position, "value does not match the pattern"})
				}
			}
		case "integer":
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				out = append(out, Violation{name, rule.Position, "value is not an integer"})
				continue
			}
			if rule.Type.Integer.Min != nil && v < *rule.Type.Integer.Min {
				out = append(out, Violation{name, rule.Position, "value is below the minimum"})
			}
			if rule.Type.Integer.Max != nil && v > *rule.Type.Integer.Max {
				out = append(out, Violation{name, rule.Position, "value is above the maximum"})
			}
		case "boolean":
			if !strings.EqualFold(value, "true") && !strings.EqualFold(value, "false") {
				out = append(out, Violation{name, rule.Position, "value is not a boolean"})
			}
		}
	}
	return out
}
