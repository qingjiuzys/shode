package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitee.com/com_818cloud/shode/pkg/debugger"
	"github.com/spf13/cobra"
)

// NewDebugAdapterCommand launches the DAP server over stdio.
func NewDebugAdapterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "debug-adapter",
		Short: "Start the Shode Debug Adapter (DAP) server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// handle ctrl+c
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
			}()

			server := debugger.NewDAPServer(os.Stdin, os.Stdout)
			if err := server.Run(ctx); err != nil && err != context.Canceled {
				return fmt.Errorf("debug adapter error: %w", err)
			}
			return nil
		},
	}
}
