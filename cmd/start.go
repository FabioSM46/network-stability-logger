package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FabioSM46/network-stability-logger/monitor"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the network stability monitor",
	Long: `Starts monitoring network stability with three concurrent monitors:
1. System Events: Interface changes, routing changes, DNS changes
2. TCP Keepalive: Persistent connection to detect Internet failures
3. Watchdog: Periodic DNS, HTTP, and routing checks`,
	RunE: runStart,
}

func init() {
	startCmd.Flags().BoolP("foreground", "f", false, "Run in foreground (don't daemonize)")
}

func runStart(cmd *cobra.Command, args []string) error {
	foreground, _ := cmd.Flags().GetBool("foreground")

	logFile := getLogPath()

	// If not in foreground and not already a daemon child, re-exec as daemon
	if !foreground && os.Getenv("_NETWORK_MONITOR_DAEMON") == "" {
		return daemonize(cmd, logFile)
	}

	if !foreground {
		// Already daemonized, suppress console output
		devNull, _ := os.Open(os.DevNull)
		os.Stdout = devNull
		os.Stderr = devNull
		os.Stdin, _ = os.Open(os.DevNull)
	}

	// Create and start the monitor
	nm, err := monitor.NewNetworkMonitor(logFile)
	if err != nil {
		return fmt.Errorf("failed to create network monitor: %w", err)
	}

	if err := nm.Start(); err != nil {
		return fmt.Errorf("failed to start network monitor: %w", err)
	}

	if foreground {
		fmt.Printf("Network monitor started successfully\n")
		fmt.Printf("Logging to: %s\n", logFile)
		fmt.Printf("Press Ctrl+C to stop...\n\n")
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\nStopping network monitor...")
	if err := nm.Stop(); err != nil {
		return fmt.Errorf("failed to stop network monitor: %w", err)
	}

	fmt.Println("Network monitor stopped")
	return nil
}
