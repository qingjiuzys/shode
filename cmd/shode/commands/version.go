package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the 'version' command
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print Shode version information",
		Long:  `Version displays the current version of the Shode runtime platform.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Shode - Secure Shell Script Runtime Platform")
			fmt.Println("Version: 0.4.0")
			fmt.Println("Build: production")
			fmt.Println("Release Date: 2025-01-21")
			fmt.Println("License: MIT")
		},
	}

	return cmd
}
