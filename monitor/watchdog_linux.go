//go:build linux

package monitor

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

func (m *WatchdogMonitor) checkDefaultRouteLinux() {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		m.logger.Log("WATCHDOG", fmt.Sprintf("ERROR: Failed to list routes: %v", err))
		return
	}

	hasDefault := false
	var defaultGw string

	for _, route := range routes {
		if route.Dst == nil {
			hasDefault = true
			if route.Gw != nil {
				defaultGw = route.Gw.String()
			}
			break
		}
	}

	if hasDefault {
		m.logger.Log("WATCHDOG", fmt.Sprintf("✓ Default route exists (via %s)", defaultGw))
	} else {
		m.logger.Log("WATCHDOG", "✗ WARNING: No default route found")
	}
}
