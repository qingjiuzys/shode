package commands

import (
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/tester"
	"github.com/spf13/cobra"
)

// NewTestCommand executes Shode script tests.
func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [paths...]",
		Short: "Run Shode script tests (files under tests/ or *_test.shode)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"tests"}
			}

			results, err := tester.Run(args)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No tests discovered.")
				return nil
			}

			var failed int
			for _, res := range results {
				status := "PASS"
				if !res.Passed {
					status = "FAIL"
					failed++
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s (%s)\n", status, res.Name, res.File)
				if !res.Passed {
					for _, failure := range res.Failures {
						fmt.Fprintf(cmd.OutOrStdout(), "    - %s\n", failure)
					}
				}
			}

			if failed > 0 {
				return fmt.Errorf("%d test(s) failed", failed)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok %d test(s) passed\n", len(results))
			return nil
		},
	}
	return cmd
}
