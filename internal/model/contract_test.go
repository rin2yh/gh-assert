package model

import "testing"

func TestContractEffective(t *testing.T) {
	commonEnv := Rule{Required: true}
	commonInput := Rule{Required: true}
	eventEnv := Rule{Type: Type{Kind: "string"}}
	eventInput := Rule{Type: Type{Kind: "integer"}}
	c := &Contract{
		Env:    map[string]Rule{"COMMON": commonEnv, "MODE": commonEnv},
		Inputs: map[string]Rule{"common": commonInput},
		On: map[string]EventContract{
			"workflow_dispatch": {
				Env:    map[string]Rule{"MODE": eventEnv},
				Inputs: map[string]Rule{"deploy_type": eventInput},
			},
		},
	}

	effective := c.Effective("workflow_dispatch")
	if effective.Env["COMMON"] != commonEnv || effective.Env["MODE"] != eventEnv {
		t.Fatalf("unexpected effective env: %#v", effective.Env)
	}
	if effective.Inputs["common"] != commonInput || effective.Inputs["deploy_type"] != eventInput {
		t.Fatalf("unexpected effective inputs: %#v", effective.Inputs)
	}

	effective.Env["COMMON"] = eventEnv
	if c.Env["COMMON"] != commonEnv {
		t.Fatal("Effective modified the common contract")
	}
}

func TestContractEffectiveWithoutMatchingEvent(t *testing.T) {
	rule := Rule{Required: true}
	c := &Contract{
		Env: map[string]Rule{"COMMON": rule},
		On:  map[string]EventContract{"workflow_dispatch": {Env: map[string]Rule{"DISPATCH": rule}}},
	}

	effective := c.Effective("workflow_run")
	if len(effective.Env) != 1 || effective.Env["COMMON"] != rule {
		t.Fatalf("unexpected effective env: %#v", effective.Env)
	}
}
