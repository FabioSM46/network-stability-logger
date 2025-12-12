package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	watchdogInterval = 30 * time.Second
	dnsTestDomain    = "www.google.com"
	httpTestURL      = "https://www.google.com"
	httpTimeout      = 10 * time.Second
)

// WatchdogMonitor performs periodic DNS/HTTP/default route checks.
type WatchdogMonitor struct {
	logger     *Logger
	ctx        context.Context
	httpClient *http.Client
}

// NewWatchdogMonitor constructs a watchdog monitor.
func NewWatchdogMonitor(ctx context.Context, logger *Logger) *WatchdogMonitor {
	return &WatchdogMonitor{
		logger: logger,
		ctx:    ctx,
		httpClient: &http.Client{
			Timeout: httpTimeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects (captive portal detection)
			},
		},
	}
}

// Start begins the watchdog periodic checks.
func (m *WatchdogMonitor) Start() error {
	m.logger.Log("WATCHDOG", "Starting watchdog monitor")

	go m.runChecks()

	return nil
}

func (m *WatchdogMonitor) runChecks() {
	m.performChecks()

	ticker := time.NewTicker(watchdogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Log("WATCHDOG", "Stopped watchdog monitor")
			return
		case <-ticker.C:
			m.performChecks()
		}
	}
}

func (m *WatchdogMonitor) performChecks() {
	m.logger.Log("WATCHDOG", "Running periodic checks...")

	m.checkDefaultRoute()
	m.checkDNS()
	m.checkHTTP()
}

func (m *WatchdogMonitor) checkDefaultRoute() {
	switch runtime.GOOS {
	case "linux":
		m.checkDefaultRouteLinux()
	case "darwin", "windows":
		m.checkDefaultRouteGeneric()
	default:
		m.logger.Log("WATCHDOG", "Default route check not supported on this platform")
	}
}

func (m *WatchdogMonitor) checkDefaultRouteGeneric() {
	// Generic check - try to get a UDP connection to check routing
	conn, err := net.DialTimeout("udp", "8.8.8.8:53", 2*time.Second)
	if err != nil {
		m.logger.Log("WATCHDOG", "✗ WARNING: Cannot establish UDP connection (no route?)")
		return
	}
	defer func() { _ = conn.Close() }()

	localAddr := conn.LocalAddr().String()
	m.logger.Log("WATCHDOG", fmt.Sprintf("✓ Default route exists (local addr: %s)", localAddr))
}

func (m *WatchdogMonitor) checkDNS() {
	start := time.Now()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	addrs, err := resolver.LookupHost(ctx, dnsTestDomain)
	duration := time.Since(start)

	if err != nil {
		m.logger.Log("WATCHDOG", fmt.Sprintf("✗ DNS FAILED: %v (took %v)", err, duration))
		return
	}

	if len(addrs) > 0 {
		m.logger.Log("WATCHDOG", fmt.Sprintf("✓ DNS working: %s -> %s (took %v)",
			dnsTestDomain, addrs[0], duration))
	}
}

func (m *WatchdogMonitor) checkHTTP() {
	start := time.Now()

	req, err := http.NewRequestWithContext(m.ctx, "HEAD", httpTestURL, nil)
	if err != nil {
		m.logger.Log("WATCHDOG", fmt.Sprintf("✗ HTTP request creation failed: %v", err))
		return
	}

	resp, err := m.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		m.logger.Log("WATCHDOG", fmt.Sprintf("✗ HTTP FAILED: %v (took %v)", err, duration))
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Detect captive portal
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		m.logger.Log("WATCHDOG", fmt.Sprintf("⚠ CAPTIVE PORTAL detected: redirect to %s", location))
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		m.logger.Log("WATCHDOG", fmt.Sprintf("✓ HTTP working: %d %s (took %v)",
			resp.StatusCode, resp.Status, duration))
	} else {
		m.logger.Log("WATCHDOG", fmt.Sprintf("⚠ HTTP unexpected status: %d %s (took %v)",
			resp.StatusCode, resp.Status, duration))
	}
}
