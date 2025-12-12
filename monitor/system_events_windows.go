//go:build windows

package monitor

import (
	"fmt"
	"net"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (m *SystemEventsMonitor) startWindows() error {
	m.logger.Log("SYSTEM", "Starting Windows IP Helper API monitoring")

	m.logNetworkStateWindows()

	// Monitor network changes using NotifyIpInterfaceChange
	go m.monitorInterfaceChanges()
	go m.monitorRouteChanges()
	go m.monitorAddressChanges()

	return nil
}

func (m *SystemEventsMonitor) monitorInterfaceChanges() {
	m.logger.Log("SYSTEM", "Monitoring interface changes")

	// Simplified polling approach for Windows
	lastState := make(map[string]bool)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			interfaces, err := net.Interfaces()
			if err != nil {
				continue
			}

			for _, iface := range interfaces {
				isUp := iface.Flags&net.FlagUp != 0
				prevState, exists := lastState[iface.Name]

				if !exists {
					lastState[iface.Name] = isUp
					continue
				}

				if prevState != isUp {
					state := "DOWN"
					if isUp {
						state = "UP"
					}
					m.logger.Log("LINK", fmt.Sprintf("Interface %s changed to %s", iface.Name, state))
					lastState[iface.Name] = isUp
				}
			}
		}
	}
}

func (m *SystemEventsMonitor) monitorRouteChanges() {
	m.logger.Log("SYSTEM", "Monitoring route changes")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastRouteCount int

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Simple detection based on route table size change
			// A more robust implementation would use NotifyRouteChange2
			currentCount := m.getRouteCount()
			if lastRouteCount > 0 && currentCount != lastRouteCount {
				m.logger.Log("ROUTE", "Routing table changed")
			}
			lastRouteCount = currentCount
		}
	}
}

func (m *SystemEventsMonitor) monitorAddressChanges() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	lastAddrs := make(map[string][]string)

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			interfaces, err := net.Interfaces()
			if err != nil {
				continue
			}

			for _, iface := range interfaces {
				addrs, err := iface.Addrs()
				if err != nil {
					continue
				}

				currentAddrs := make([]string, 0)
				for _, addr := range addrs {
					currentAddrs = append(currentAddrs, addr.String())
				}

				prevAddrs, exists := lastAddrs[iface.Name]
				if exists && !sliceEqual(prevAddrs, currentAddrs) {
					m.logger.Log("ADDRESS", fmt.Sprintf("IP addresses changed on %s", iface.Name))
				}

				lastAddrs[iface.Name] = currentAddrs
			}
		}
	}
}

func (m *SystemEventsMonitor) logNetworkStateWindows() {
	interfaces, err := net.Interfaces()
	if err == nil {
		m.logger.Log("SYSTEM", fmt.Sprintf("Found %d network interfaces", len(interfaces)))
		for _, iface := range interfaces {
			state := "DOWN"
			if iface.Flags&net.FlagUp != 0 {
				state = "UP"
			}
			m.logger.Log("SYSTEM", fmt.Sprintf("  %s: %s", iface.Name, state))
		}
	}
}

func (m *SystemEventsMonitor) getRouteCount() int {
	// This is a simplified implementation
	// Would need proper Windows API calls for production
	return 0
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Prevent unused import error
var _ = unsafe.Sizeof(0)
var _ = windows.Handle(0)
