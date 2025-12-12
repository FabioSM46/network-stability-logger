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

func NewSystemEventsMonitor(logger *Logger, ctx context.Context) *SystemEventsMonitor {
	return &SystemEventsMonitor{
		logger: logger,
		ctx:    ctx,
	}
}

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
