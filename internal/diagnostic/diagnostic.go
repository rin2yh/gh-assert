package diagnostic

import (
	"fmt"
	"io"

	"github.com/rin2yh/gh-assert/internal/assertion"
)

func WriteViolations(w io.Writer, violations []assertion.Violation) error {
	for _, v := range violations {
		if _, err := fmt.Fprintf(w, "%s:%d:%d: env %s: %s\n", v.Position.Path, v.Position.Line, v.Position.Column, v.Env, v.Message); err != nil {
			return err
		}
	}
	return nil
}
