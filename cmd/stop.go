package cmd

import (
	"fmt"
	"os"
	"os/exec"
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

func runStop(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pidFile := filepath.Join(homeDir, ".network-monitor", "monitor.pid")

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
		fmt.Printf("Process %d not found\n", pid)
		os.Remove(pidFile)
		return nil
	}

	// Try to kill the process
	fmt.Printf("Stopping network monitor (PID: %d)...\n", pid)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might not exist
		fmt.Printf("Failed to send signal: %v\n", err)
		fmt.Println("Removing stale PID file...")
		os.Remove(pidFile)
		return nil
	}

	// Clean up PID file
	os.Remove(pidFile)
	fmt.Println("Network monitor stopped")

	return nil
}

// Helper function to check if process is running
func isProcessRunning(pid int) bool {
	// Try to send signal 0 (doesn't actually send a signal, just checks if process exists)
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Helper to check using ps command
func isProcessRunningPS(pid int) bool {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid))
	err := cmd.Run()
	return err == nil
}
