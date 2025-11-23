package commands

import (
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/formatter"
	"github.com/spf13/cobra"
)

// NewFmtCommand formats Shode scripts in-place.
func NewFmtCommand() *cobra.Command {
	var check bool

	cmd := &cobra.Command{
		Use:   "fmt [paths...]",
		Short: "Format Shode scripts",
		Long: `Formats .sh/.sho/.shode scripts with a simple indentation style.
By default files are rewritten in place. Use --check to verify formatting without writing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"."}
			}

			changed, err := formatter.FormatPath(args, !check)
			if err != nil {
				return err
			}

			if check && len(changed) > 0 {
				for _, file := range changed {
					fmt.Fprintf(cmd.ErrOrStderr(), "needs formatting: %s\n", file)
				}
				return fmt.Errorf("formatting required for %d file(s)", len(changed))
			}

			if !check {
				for _, file := range changed {
					fmt.Fprintf(cmd.OutOrStdout(), "formatted %s\n", file)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "check formatting without modifying files")
	return cmd
}
