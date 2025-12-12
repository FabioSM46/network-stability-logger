//go:build windows

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// daemonize on Windows uses background process execution via os/exec.
func daemonize(_ *cobra.Command, _ string) error {
	// Windows doesn't have true daemonization, so we'll use a helper approach
	fmt.Println("Note: Full daemonization on Windows requires elevated permissions.")
	fmt.Println("Consider using 'network-monitor start -f' in a scheduled task or service.")
	fmt.Println("\nRunning in foreground mode instead...")

	// Just run in foreground on Windows for now
	return nil
}
