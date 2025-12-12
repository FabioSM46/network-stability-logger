//go:build !darwin

package monitor

import "fmt"

func (m *SystemEventsMonitor) startDarwin() error {
	return fmt.Errorf("macOS monitoring not available on this platform")
}

func (m *SystemEventsMonitor) parseRoutingMessage(data []byte) {}
func (m *SystemEventsMonitor) logNetworkStateDarwin()          {}
func (m *SystemEventsMonitor) monitorDNSChangesDarwin()        {}
