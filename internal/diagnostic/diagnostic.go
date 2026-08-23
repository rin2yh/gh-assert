package diagnostic

import (
	"fmt"
	"io"

	"github.com/rin2yh/gh-assert/internal/runtime"
)

func WriteViolations(w io.Writer, violations []runtime.Violation) error {
	for _, v := range violations {
		if _, err := fmt.Fprintf(w, "%s:%d:%d: %s %s: %s\n", v.Position.Path, v.Position.Line, v.Position.Column, v.Scope, v.Name, v.Message); err != nil {
			return err
		}
	}
	return nil
}
