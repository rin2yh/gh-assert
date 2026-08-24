package target

import (
	"path/filepath"
	"testing"
)

func TestWorkflowPath(t *testing.T) {
	tests := []struct {
		name, contract, want string
		ok                   bool
	}{
		{name: "workflow", contract: filepath.Join("repo", ".github", "workflows", "deploy_assert.yml"), want: filepath.Join("repo", ".github", "workflows", "deploy.yml"), ok: true},
		{name: "workflow named action", contract: filepath.Join("repo", ".github", "workflows", "action_assert.yml"), want: filepath.Join("repo", ".github", "workflows", "action.yml"), ok: true},
		{name: "action", contract: filepath.Join("repo", ".github", "actions", "deploy", "action_assert.yml")},
		{name: "other directory", contract: filepath.Join("repo", "deploy_assert.yml")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := WorkflowPath(tt.contract)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("WorkflowPath() = %q, %t, want %q, %t", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestActionPath(t *testing.T) {
	tests := []struct {
		name, contract, want string
		ok                   bool
	}{
		{name: "action", contract: filepath.Join("repo", ".github", "actions", "deploy", "action_assert.yml"), want: filepath.Join("repo", ".github", "actions", "deploy", "action.yml"), ok: true},
		{name: "nested action", contract: filepath.Join("repo", ".github", "actions", "group", "deploy", "action_assert.yml"), want: filepath.Join("repo", ".github", "actions", "group", "deploy", "action.yml"), ok: true},
		{name: "workflow named action", contract: filepath.Join("repo", ".github", "workflows", "action_assert.yml")},
		{name: "filename only", contract: filepath.Join("repo", "action_assert.yml")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ActionPath(tt.contract)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("ActionPath() = %q, %t, want %q, %t", got, ok, tt.want, tt.ok)
			}
		})
	}
}
