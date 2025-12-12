//go:build !linux

package monitor

import "fmt"

func (m *SystemEventsMonitor) startLinux() error {
	return fmt.Errorf("Linux monitoring not available on this platform")
}

// nolint:unused
func (m *SystemEventsMonitor) handleLinkUpdate(update interface{}) {}

// nolint:unused
func (m *SystemEventsMonitor) handleAddrUpdate(update interface{}) {}

// nolint:unused
func (m *SystemEventsMonitor) handleRouteUpdate(update interface{}) {}

// nolint:unused
func (m *SystemEventsMonitor) monitorDNSChanges() {}

// nolint:unused
func (m *SystemEventsMonitor) logNetworkState() {}

// nolint:unused
func (m *SystemEventsMonitor) logDNSServers() {}
