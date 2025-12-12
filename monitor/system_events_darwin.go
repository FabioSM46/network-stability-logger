//go:build darwin

package monitor

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/sys/unix"
)

func (m *SystemEventsMonitor) startDarwin() error {
	m.logger.Log("SYSTEM", "Starting macOS routing socket monitoring")

	// Create routing socket
	fd, err := unix.Socket(unix.AF_ROUTE, unix.SOCK_RAW, unix.AF_UNSPEC)
	if err != nil {
		return fmt.Errorf("failed to create routing socket: %w", err)
	}

	m.logger.Log("SYSTEM", "Monitoring network changes via routing socket")
	m.logNetworkStateDarwin()

	go func() {
		defer unix.Close(fd)

		buf := make([]byte, 2048)
		for {
			select {
			case <-m.ctx.Done():
				m.logger.Log("SYSTEM", "Stopped macOS routing socket monitoring")
				return
			default:
				n, err := unix.Read(fd, buf)
				if err != nil {
					if m.ctx.Err() != nil {
						return
					}
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if n > 0 {
					m.parseRoutingMessage(buf[:n])
				}
			}
		}
	}()

	// Monitor DNS changes
	go m.monitorDNSChangesDarwin()

	return nil
}

func (m *SystemEventsMonitor) parseRoutingMessage(data []byte) {
	if len(data) < 4 {
		return
	}

	msgType := data[3]

	switch msgType {
	case unix.RTM_NEWADDR:
		m.logger.Log("ADDRESS", "IP address added")
	case unix.RTM_DELADDR:
		m.logger.Log("ADDRESS", "IP address removed")
	case unix.RTM_IFINFO:
		m.logger.Log("LINK", "Interface state changed")
	case unix.RTM_ADD:
		m.logger.Log("ROUTE", "Route added")
	case unix.RTM_DELETE:
		m.logger.Log("ROUTE", "Route deleted")
	case unix.RTM_CHANGE:
		m.logger.Log("ROUTE", "Route modified")
	}
}

func (m *SystemEventsMonitor) logNetworkStateDarwin() {
	interfaces, err := net.Interfaces()
	if err == nil {
		m.logger.Log("SYSTEM", fmt.Sprintf("Found %d network interfaces", len(interfaces)))
		for _, iface := range interfaces {
			if iface.Name != "lo0" {
				state := "DOWN"
				if iface.Flags&net.FlagUp != 0 {
					state = "UP"
				}
				m.logger.Log("SYSTEM", fmt.Sprintf("  %s: %s", iface.Name, state))
			}
		}
	}
}

func (m *SystemEventsMonitor) monitorDNSChangesDarwin() {
	// On macOS, DNS settings can be monitored through system configuration
	// For simplicity, we poll /etc/resolv.conf
	lastModTime := time.Time{}
	resolvPath := "/etc/resolv.conf"

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			var stat unix.Stat_t
			if err := unix.Stat(resolvPath, &stat); err != nil {
				continue
			}

			modTime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
			if !lastModTime.IsZero() && modTime.After(lastModTime) {
				m.logger.Log("DNS", "DNS configuration changed")
			}
			lastModTime = modTime
		}
	}
}
