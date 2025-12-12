//go:build !windows

package monitor

import "fmt"

func (m *SystemEventsMonitor) startWindows() error {
	return fmt.Errorf("windows monitoring not available on this platform")
}

// nolint:unused
func (m *SystemEventsMonitor) monitorInterfaceChanges() {}

// nolint:unused
func (m *SystemEventsMonitor) monitorRouteChanges() {}

// nolint:unused
func (m *SystemEventsMonitor) monitorAddressChanges() {}

// nolint:unused
func (m *SystemEventsMonitor) logNetworkStateWindows() {}

// nolint:unused
func (m *SystemEventsMonitor) getRouteCount() int { return 0 }
