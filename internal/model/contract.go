package model

import (
	"maps"
	"regexp"
)

type Contract struct {
	Env    map[string]Rule          `yaml:"env"`
	Inputs map[string]Rule          `yaml:"inputs"`
	On     map[string]EventContract `yaml:"on"`
}

type EventContract struct {
	Env    map[string]Rule `yaml:"env"`
	Inputs map[string]Rule `yaml:"inputs"`
}

// Effective returns the common rules combined with the rules for each event.
// Event-specific rules replace common rules with the same name.
func (c *Contract) Effective(events ...string) *Contract {
	effective := &Contract{
		Env:    maps.Clone(c.Env),
		Inputs: maps.Clone(c.Inputs),
	}
	for _, event := range events {
		scoped, ok := c.On[event]
		if !ok {
			continue
		}
		effective.Env = mergeRules(effective.Env, scoped.Env)
		effective.Inputs = mergeRules(effective.Inputs, scoped.Inputs)
	}
	return effective
}

func mergeRules(common, scoped map[string]Rule) map[string]Rule {
	if scoped == nil {
		return common
	}
	if common == nil {
		common = make(map[string]Rule, len(scoped))
	}
	maps.Copy(common, scoped)
	return common
}

type ContractFile struct {
	Path     string
	Contract *Contract
}

type Rule struct {
	Required     bool     `yaml:"required,omitempty"`
	Type         Type     `yaml:"type"`
	Position     Position `yaml:"-"`
	TypePosition Position `yaml:"-"`
}

type Type struct {
	String  *StringType  `yaml:"string,omitempty"`
	Integer *IntegerType `yaml:"integer,omitempty"`
	Boolean *struct{}    `yaml:"boolean,omitempty"`
	Kind    string       `yaml:"-"`
}

type StringType struct {
	Enum        []string       `yaml:"enum,omitempty"`
	PatternText string         `yaml:"pattern,omitempty"`
	Pattern     *regexp.Regexp `yaml:"-"`
}

type IntegerType struct {
	Min *int64 `yaml:"min,omitempty"`
	Max *int64 `yaml:"max,omitempty"`
}
