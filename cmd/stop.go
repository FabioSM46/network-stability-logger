// Package cmd contains CLI commands for the network monitor.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the network stability monitor",
	Long:  `Stops a running network stability monitor process.`,
	RunE:  runStop,
}

func runStop(_ *cobra.Command, _ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pidFile := filepath.Clean(filepath.Join(homeDir, ".network-monitor", "monitor.pid"))

	// Read PID file
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Network monitor is not running (no PID file found)")
			return nil
		}
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(pidFile)
		return fmt.Errorf("process %d not found: %w", pid, err)
	}

	// Try to kill the process
	fmt.Printf("Stopping network monitor (PID: %d)...\n", pid)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might not exist
		_ = os.Remove(pidFile)
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Clean up PID file
	_ = os.Remove(pidFile)
	fmt.Println("Network monitor stopped")

	return nil
}
