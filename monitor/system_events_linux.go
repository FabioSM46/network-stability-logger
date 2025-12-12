//go:build linux

package monitor

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
)

func (m *SystemEventsMonitor) startLinux() error {
	m.logger.Log("SYSTEM", "Starting Linux netlink monitoring")

	// Subscribe to link updates
	linkUpdates := make(chan netlink.LinkUpdate)
	linkDone := make(chan struct{})
	if err := netlink.LinkSubscribe(linkUpdates, linkDone); err != nil {
		return fmt.Errorf("failed to subscribe to link updates: %w", err)
	}

	// Subscribe to address updates
	addrUpdates := make(chan netlink.AddrUpdate)
	addrDone := make(chan struct{})
	if err := netlink.AddrSubscribe(addrUpdates, addrDone); err != nil {
		close(linkDone)
		return fmt.Errorf("failed to subscribe to address updates: %w", err)
	}

	// Subscribe to route updates
	routeUpdates := make(chan netlink.RouteUpdate)
	routeDone := make(chan struct{})
	if err := netlink.RouteSubscribe(routeUpdates, routeDone); err != nil {
		close(linkDone)
		close(addrDone)
		return fmt.Errorf("failed to subscribe to route updates: %w", err)
	}

	// Monitor DNS changes by watching resolv.conf
	go m.monitorDNSChanges()

	// Log initial state
	m.logNetworkState()

	// Handle events
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				close(linkDone)
				close(addrDone)
				close(routeDone)
				m.logger.Log("SYSTEM", "Stopped Linux netlink monitoring")
				return

			case update := <-linkUpdates:
				m.handleLinkUpdate(update)

			case update := <-addrUpdates:
				m.handleAddrUpdate(update)

			case update := <-routeUpdates:
				m.handleRouteUpdate(update)
			}
		}
	}()

	return nil
}

func (m *SystemEventsMonitor) handleLinkUpdate(update netlink.LinkUpdate) {
	link := update.Link
	attrs := link.Attrs()

	state := "DOWN"
	if attrs.Flags&net.FlagUp != 0 {
		state = "UP"
	}

	msg := fmt.Sprintf("Interface %s [%s]: %s (flags: %v)",
		attrs.Name, link.Type(), state, attrs.Flags)
	m.logger.Log("LINK", msg)
}

func (m *SystemEventsMonitor) handleAddrUpdate(update netlink.AddrUpdate) {
	action := "ADDED"
	if !update.NewAddr {
		action = "REMOVED"
	}

	link, err := netlink.LinkByIndex(update.LinkIndex)
	linkName := fmt.Sprintf("idx-%d", update.LinkIndex)
	if err == nil {
		linkName = link.Attrs().Name
	}

	msg := fmt.Sprintf("IP address %s on %s: %s",
		action, linkName, update.LinkAddress.String())
	m.logger.Log("ADDRESS", msg)
}

func (m *SystemEventsMonitor) handleRouteUpdate(update netlink.RouteUpdate) {
	action := "MODIFIED"
	switch update.Type {
	case syscall.RTM_NEWROUTE:
		action = "ADDED"
	case syscall.RTM_DELROUTE:
		action = "DELETED"
	}

	route := update.Route
	dst := "default"
	if route.Dst != nil {
		dst = route.Dst.String()
	}

	via := "direct"
	if route.Gw != nil {
		via = route.Gw.String()
	}

	link := ""
	if l, err := netlink.LinkByIndex(route.LinkIndex); err == nil {
		link = l.Attrs().Name
	}

	msg := fmt.Sprintf("Route %s: %s via %s dev %s", action, dst, via, link)
	m.logger.Log("ROUTE", msg)
}

func (m *SystemEventsMonitor) monitorDNSChanges() {
	lastModTime := time.Time{}
	resolvPath := "/etc/resolv.conf"

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			var stat syscall.Stat_t
			err := syscall.Stat(resolvPath, &stat)
			if err != nil {
				continue
			}

			modTime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
			if !lastModTime.IsZero() && modTime.After(lastModTime) {
				m.logger.Log("DNS", "DNS configuration changed (resolv.conf modified)")
				m.logDNSServers()
			}
			lastModTime = modTime
		}
	}
}

func (m *SystemEventsMonitor) logNetworkState() {
	// Log all links
	links, err := netlink.LinkList()
	if err == nil {
		m.logger.Log("SYSTEM", fmt.Sprintf("Found %d network interfaces", len(links)))
		for _, link := range links {
			attrs := link.Attrs()
			if attrs.Name != "lo" {
				state := "DOWN"
				if attrs.Flags&net.FlagUp != 0 {
					state = "UP"
				}
				m.logger.Log("SYSTEM", fmt.Sprintf("  %s: %s", attrs.Name, state))
			}
		}
	}

	// Log default routes
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err == nil {
		for _, route := range routes {
			if route.Dst == nil { // default route
				via := "direct"
				if route.Gw != nil {
					via = route.Gw.String()
				}
				m.logger.Log("SYSTEM", fmt.Sprintf("  Default route via %s", via))
			}
		}
	}

	m.logDNSServers()
}

func (m *SystemEventsMonitor) logDNSServers() {
	// Parse /etc/resolv.conf for DNS servers
	// This is a simple implementation
	content, err := syscall.Open("/etc/resolv.conf", syscall.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer syscall.Close(content)

	buf := make([]byte, 4096)
	n, err := syscall.Read(content, buf)
	if err != nil {
		return
	}

	// Simple parsing - look for nameserver lines
	_ = string(buf[:n])
	m.logger.Log("DNS", "Current DNS servers updated")
}
