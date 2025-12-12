//go:build linux || darwin

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

// daemonize forks the process to run as a background daemon (Unix-like systems).
func daemonize(cmd *cobra.Command, logFile string) error {
	// Re-exec self as daemon child
	argv := []string{os.Args[0], "start"}

	// Preserve foreground flag (will be false by default)
	foreground, _ := cmd.Flags().GetBool("foreground")
	if foreground {
		argv = append(argv, "-f")
	}

	// Preserve log path if custom
	logPath, _ := cmd.Flags().GetString("log-path")
	if logPath != "" {
		argv = append(argv, "-l", logPath)
	}

	attr := &syscall.ProcAttr{
		Dir:   ".",
		Env:   append(os.Environ(), "_NETWORK_MONITOR_DAEMON=1"),
		Files: []uintptr{0, 1, 2}, // stdin, stdout, stderr
		Sys: &syscall.SysProcAttr{
			Setsid: true, // Create new session
		},
	}

	pid, err := syscall.ForkExec(argv[0], argv, attr)
	if err != nil {
		return fmt.Errorf("failed to fork: %w", err)
	}

	// Write PID file
	homeDir, _ := os.UserHomeDir()
	pidDir := filepath.Join(homeDir, ".network-monitor")
	_ = os.MkdirAll(pidDir, 0755)

	pidFile := filepath.Join(pidDir, "monitor.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	fmt.Printf("Started network monitor in background (PID: %d)\n", pid)
	fmt.Printf("Log file: %s\n", logFile)
	fmt.Printf("\nUse 'network-monitor log' to view logs\n")
	fmt.Printf("Use 'network-monitor stop' to stop the monitor\n")

	return nil
}
