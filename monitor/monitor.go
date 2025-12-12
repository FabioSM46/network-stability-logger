package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type NetworkMonitor struct {
	logger     *Logger
	ctx        context.Context
	cancel     context.CancelFunc
	sysEvents  *SystemEventsMonitor
	tcpMonitor *TCPKeepaliveMonitor
	watchdog   *WatchdogMonitor
}

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
		sysEvents:  NewSystemEventsMonitor(logger, ctx),
		tcpMonitor: NewTCPKeepaliveMonitor(logger, ctx),
		watchdog:   NewWatchdogMonitor(logger, ctx),
	}, nil
}

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

func (nm *NetworkMonitor) Stop() error {
	nm.logger.Log("MONITOR", "=== Network Stability Monitor Stopping ===")

	nm.cancel()

	if err := nm.logger.Close(); err != nil {
		return fmt.Errorf("failed to close logger: %w", err)
	}

	return nil
}

func (nm *NetworkMonitor) Wait() {
	<-nm.ctx.Done()
}

func GetDefaultLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".network-monitor", "network-monitor.log")
}
