package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rin2yh/gh-assert/internal/contract"
	"github.com/rin2yh/gh-assert/internal/diagnostic"
	"github.com/rin2yh/gh-assert/internal/reusable"
	"github.com/rin2yh/gh-assert/internal/runtime"
	"github.com/spf13/cobra"
)

type commandError struct {
	code int
	err  error
}

func (e *commandError) Error() string { return e.err.Error() }
func (e *commandError) Unwrap() error { return e.err }

func main() { os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr)) }

func execute(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] != "validate" && args[0] != "help" && args[0] != "--help" && args[0] != "-h" {
		cmd := &cobra.Command{SilenceErrors: true, SilenceUsage: true}
		cmd.SetErr(stderr)
		if err := runRuntime(cmd, args); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			var ce *commandError
			if errors.As(err, &ce) {
				return ce.code
			}
			return 2
		}
		return 0
	}
	root := newRootCommand(stdout, stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		var ce *commandError
		if errors.As(err, &ce) {
			return ce.code
		}
		return 2
	}
	return 0
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "gh-assert",
		Short:         "Assert GitHub Actions environment variables and workflow inputs",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRuntime(cmd, args)
		},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(newValidateCommand(stdout))
	root.AddCommand(newHelpCommand(stdout))
	return root
}

func newHelpCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Show help for gh-assert",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(stdout)
			return cmd.Root().Help()
		},
	}
}

func newValidateCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [<contract>]",
		Short: "Validate a contract definition",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := ".github"
			if len(args) == 1 {
				path = args[0]
			}
			loaded, err := contract.LoadTargets(path)
			if err != nil {
				return &commandError{code: 2, err: err}
			}
			for _, item := range loaded {
				if err := reusable.Validate(item); err != nil {
					return &commandError{code: 2, err: err}
				}
				if _, err := fmt.Fprintf(stdout, "%s: contract is valid\n", item.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func runRuntime(cmd *cobra.Command, args []string) error {
	path := ".github"
	switch {
	case len(args) == 1:
		path = args[0]
	case len(args) == 2 && args[0] == "--contract":
		path = args[1]
	case len(args) != 0:
		return &commandError{code: 2, err: errors.New("expected a contract path or --contract <path>")}
	}
	loaded, err := contract.LoadTargets(path)
	if err != nil {
		return &commandError{code: 2, err: err}
	}
	failed := false
	for _, item := range loaded {
		if err := reusable.Validate(item); err != nil {
			return &commandError{code: 2, err: err}
		}
		violations, err := runtime.Assert(item)
		if err != nil {
			return &commandError{code: 2, err: err}
		}
		if len(violations) > 0 {
			if err := diagnostic.WriteViolations(cmd.ErrOrStderr(), violations); err != nil {
				return &commandError{code: 2, err: err}
			}
			failed = true
		}
	}
	if failed {
		return &commandError{code: 1, err: errors.New("contract assertion failed")}
	}
	return nil
}
