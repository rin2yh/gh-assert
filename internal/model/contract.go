package model

import "regexp"

type Contract struct {
	Env    map[string]Rule          `yaml:"env"`
	Inputs map[string]Rule          `yaml:"inputs"`
	On     map[string]EventContract `yaml:"on"`
}

type EventContract struct {
	Env    map[string]Rule `yaml:"env"`
	Inputs map[string]Rule `yaml:"inputs"`
}

type ContractFile struct {
	Path        string
	SiblingPath string
	Kind        ContractKind
	Contract    *Contract
}

type ContractKind string

const (
	WorkflowContract        ContractKind = "workflow"
	CompositeActionContract ContractKind = "composite-action"
)

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
