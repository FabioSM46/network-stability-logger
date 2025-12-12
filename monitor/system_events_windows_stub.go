//go:build !windows

package monitor

import "fmt"

func (m *SystemEventsMonitor) startWindows() error {
	return fmt.Errorf("Windows monitoring not available on this platform")
}

func (m *SystemEventsMonitor) monitorInterfaceChanges() {}
func (m *SystemEventsMonitor) monitorRouteChanges()     {}
func (m *SystemEventsMonitor) monitorAddressChanges()   {}
func (m *SystemEventsMonitor) logNetworkStateWindows()  {}
func (m *SystemEventsMonitor) getRouteCount() int       { return 0 }
