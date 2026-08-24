package contract

import (
	"maps"

	"github.com/rin2yh/gh-assert/internal/model"
)

func Resolve(parsed *model.Contract, event string) *model.Contract {
	resolved := &model.Contract{
		Env:    maps.Clone(parsed.Env),
		Inputs: maps.Clone(parsed.Inputs),
	}
	scoped, ok := parsed.On[event]
	if !ok {
		return resolved
	}
	resolved.Env = mergeRules(resolved.Env, scoped.Env)
	resolved.Inputs = mergeRules(resolved.Inputs, scoped.Inputs)
	return resolved
}

func mergeRules(common, scoped map[string]model.Rule) map[string]model.Rule {
	if scoped == nil {
		return common
	}
	if common == nil {
		common = make(map[string]model.Rule, len(scoped))
	}
	maps.Copy(common, scoped)
	return common
}
