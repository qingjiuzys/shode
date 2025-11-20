package commands

import (
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/lint"
	"github.com/spf13/cobra"
)

// NewLintCommand statically analyses scripts for common pitfalls.
func NewLintCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint [paths...]",
		Short: "Run basic static checks against Shode scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"."}
			}

			issues, err := lint.LintPath(args)
			if err != nil {
				return err
			}
			if len(issues) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No lint issues found.")
				return nil
			}

			for _, issue := range issues {
				line := ""
				if issue.Line > 0 {
					line = fmt.Sprintf(":%d", issue.Line)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s [%s] %s\n",
					issue.File, line, issue.Severity, issue.Message)
			}
			return fmt.Errorf("%d lint issue(s) detected", len(issues))
		},
	}
	return cmd
}
