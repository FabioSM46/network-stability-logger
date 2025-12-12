//go:build !darwin

package monitor

import "fmt"

func (m *SystemEventsMonitor) startDarwin() error {
	return fmt.Errorf("macOS monitoring not available on this platform")
}

// nolint:unused
func (m *SystemEventsMonitor) parseRoutingMessage(_ []byte) {}

//nolint:revive // Darwin stub; parameter required for signature parity

// nolint:unused
func (m *SystemEventsMonitor) logNetworkStateDarwin() {}

// nolint:unused
func (m *SystemEventsMonitor) monitorDNSChangesDarwin() {}
