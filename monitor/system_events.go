package monitor

import (
	"context"
	"fmt"
	"runtime"
)

// SystemEventsMonitor handles platform-specific network event monitoring
type SystemEventsMonitor struct {
	logger *Logger
	ctx    context.Context
}

// NewSystemEventsMonitor creates a system events monitor for the given context.
func NewSystemEventsMonitor(ctx context.Context, logger *Logger) *SystemEventsMonitor {
	return &SystemEventsMonitor{
		logger: logger,
		ctx:    ctx,
	}
}

// Start begins platform-specific system events monitoring.
func (m *SystemEventsMonitor) Start() error {
	m.logger.Log("SYSTEM", "Starting system events monitor")

	switch runtime.GOOS {
	case "linux":
		return m.startLinux()
	case "darwin":
		return m.startDarwin()
	case "windows":
		return m.startWindows()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
