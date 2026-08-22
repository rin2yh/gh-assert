package assertion

import (
	"testing"

	"github.com/rin2yh/gh-assert/internal/contract"
)

func loadRuntimeContract(t *testing.T) *contract.Contract {
	t.Helper()
	path := "testdata/runtime.yml"
	c, err := contract.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func assertViolationCount(t *testing.T, c *contract.Contract, env environment, want int) {
	t.Helper()
	if got := len(Validate(c, env)); got != want {
		t.Fatalf("got %d violations, want %d", got, want)
	}
}

func TestValidateRequiredEnvironment(t *testing.T) {
	c := loadRuntimeContract(t)
	tests := []struct {
		name string
		env  environment
		want int
	}{
		{"required missing", environment{"COUNT": "3", "FLAG": "false"}, 1},
		{"required empty", environment{"NAME": "", "COUNT": "3", "FLAG": "false"}, 1},
		{"optional missing", environment{"NAME": "staging", "FLAG": "true"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertViolationCount(t, c, tt.env, tt.want) })
	}
}

func TestValidateStringConstraints(t *testing.T) {
	c := loadRuntimeContract(t)
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"valid", "staging", 0},
		{"enum and pattern", "production1", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := environment{"NAME": tt.value, "COUNT": "3", "FLAG": "false"}
			assertViolationCount(t, c, env, tt.want)
		})
	}
}

func TestValidateIntegerConstraints(t *testing.T) {
	c := loadRuntimeContract(t)
	tests := []struct{ name, value string }{
		{"above maximum", "6"},
		{"invalid type", "three"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertViolationCount(t, c, environment{"NAME": "staging", "COUNT": tt.value, "FLAG": "false"}, 1)
		})
	}
}

func TestValidateBooleanType(t *testing.T) {
	c := loadRuntimeContract(t)
	assertViolationCount(t, c, environment{"NAME": "staging", "COUNT": "3", "FLAG": "yes"}, 1)
}
