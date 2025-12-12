//go:build !linux

package monitor

import "fmt"

func (m *SystemEventsMonitor) startLinux() error {
	return fmt.Errorf("Linux monitoring not available on this platform")
}

func (m *SystemEventsMonitor) handleLinkUpdate(update interface{})  {}
func (m *SystemEventsMonitor) handleAddrUpdate(update interface{})  {}
func (m *SystemEventsMonitor) handleRouteUpdate(update interface{}) {}
func (m *SystemEventsMonitor) monitorDNSChanges()                   {}
func (m *SystemEventsMonitor) logNetworkState()                     {}
func (m *SystemEventsMonitor) logDNSServers()                       {}
