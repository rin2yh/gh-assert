package github

import (
	"path/filepath"
	"testing"
)

func TestActionParser(t *testing.T) {
	path := filepath.Join("testdata", "action.yml")
	action, err := NewActionParser(path).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if action.Runs.Using != "composite" {
		t.Fatalf("runs.using = %q, want composite", action.Runs.Using)
	}
	if !action.Inputs["environment"].Required {
		t.Fatal("environment input is not required")
	}
}

func TestActionParserReturnsReadError(t *testing.T) {
	_, err := NewActionParser(filepath.Join(t.TempDir(), "missing.yml")).Parse()
	if err == nil {
		t.Fatal("expected an error")
	}
}
