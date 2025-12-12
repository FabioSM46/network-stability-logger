// Package monitor provides logging and monitoring primitives.
package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// NetworkMonitor coordinates system, TCP, and watchdog monitors.
type NetworkMonitor struct {
	logger     *Logger
	ctx        context.Context
	cancel     context.CancelFunc
	sysEvents  *SystemEventsMonitor
	tcpMonitor *TCPKeepaliveMonitor
	watchdog   *WatchdogMonitor
}

// NewNetworkMonitor constructs a monitor with the given log path.
func NewNetworkMonitor(logPath string) (*NetworkMonitor, error) {
	logger, err := NewLogger(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &NetworkMonitor{
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		sysEvents:  NewSystemEventsMonitor(ctx, logger),
		tcpMonitor: NewTCPKeepaliveMonitor(ctx, logger),
		watchdog:   NewWatchdogMonitor(ctx, logger),
	}, nil
}

// Start begins all monitoring routines.
func (nm *NetworkMonitor) Start() error {
	nm.logger.Log("MONITOR", "=== Network Stability Monitor Starting ===")

	// Start all three goroutines
	if err := nm.sysEvents.Start(); err != nil {
		return fmt.Errorf("failed to start system events monitor: %w", err)
	}

	if err := nm.tcpMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start TCP keepalive monitor: %w", err)
	}

	if err := nm.watchdog.Start(); err != nil {
		return fmt.Errorf("failed to start watchdog monitor: %w", err)
	}

	nm.logger.Log("MONITOR", "All monitors started successfully")

	return nil
}

// Stop signals monitors to terminate and closes the logger.
func (nm *NetworkMonitor) Stop() error {
	nm.logger.Log("MONITOR", "=== Network Stability Monitor Stopping ===")

	nm.cancel()

	if err := nm.logger.Close(); err != nil {
		return fmt.Errorf("failed to close logger: %w", err)
	}

	return nil
}

// Wait blocks until the monitor context is canceled.
func (nm *NetworkMonitor) Wait() {
	<-nm.ctx.Done()
}

// GetDefaultLogPath returns a default log file path near the executable.
func GetDefaultLogPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "network-monitor.log"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "network-monitor.log")
}
