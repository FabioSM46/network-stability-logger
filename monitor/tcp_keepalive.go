package monitor

import (
	"context"
	"fmt"
	"net"
	"time"
)

const (
	keepaliveTarget = "1.1.1.1:443" // Cloudflare DNS over HTTPS
	reconnectDelay  = 5 * time.Second
)

// TCPKeepaliveMonitor maintains a persistent TCP connection and reports connectivity.
type TCPKeepaliveMonitor struct {
	logger *Logger
	ctx    context.Context
	conn   net.Conn
}

// NewTCPKeepaliveMonitor constructs a TCP keepalive monitor.
func NewTCPKeepaliveMonitor(ctx context.Context, logger *Logger) *TCPKeepaliveMonitor {
	return &TCPKeepaliveMonitor{
		logger: logger,
		ctx:    ctx,
	}
}

// Start launches the TCP keepalive monitoring loop.
func (m *TCPKeepaliveMonitor) Start() error {
	m.logger.Log("TCP", fmt.Sprintf("Starting persistent TCP keepalive monitor to %s", keepaliveTarget))

	go m.maintainConnection()

	return nil
}

func (m *TCPKeepaliveMonitor) maintainConnection() {
	for {
		select {
		case <-m.ctx.Done():
			if m.conn != nil {
				_ = m.conn.Close()
			}
			m.logger.Log("TCP", "Stopped TCP keepalive monitor")
			return
		default:
			if err := m.connect(); err != nil {
				m.logger.Log("TCP", fmt.Sprintf("ERROR: Failed to connect: %v", err))
				time.Sleep(reconnectDelay)
				continue
			}

			// Monitor the connection
			if err := m.monitorConnection(); err != nil {
				m.logger.Log("TCP", fmt.Sprintf("ERROR: Connection failed: %v", err))
				if m.conn != nil {
					_ = m.conn.Close()
					m.conn = nil
				}
				time.Sleep(reconnectDelay)
			}
		}
	}
}

func (m *TCPKeepaliveMonitor) connect() error {
	if m.conn != nil {
		return nil // Already connected
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dialer.DialContext(m.ctx, "tcp", keepaliveTarget)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	// Enable TCP keepalive at the socket level
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetKeepAlive(true); err != nil {
			_ = conn.Close()
			return fmt.Errorf("failed to enable keepalive: %w", err)
		}
		if err := tcpConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
			_ = conn.Close()
			return fmt.Errorf("failed to set keepalive period: %w", err)
		}

		// Apply platform-specific keepalive tuning (no-op where unsupported)
		if rawConn, err := tcpConn.SyscallConn(); err == nil {
			_ = rawConn.Control(func(fd uintptr) {
				applyPlatformKeepalive(fd)
			})
		}
	}

	m.conn = conn
	m.logger.Log("TCP", fmt.Sprintf("SUCCESS: Connected to %s", keepaliveTarget))

	return nil
}

func (m *TCPKeepaliveMonitor) monitorConnection() error {
	// Set read deadline to detect connection failures
	readTimeout := 60 * time.Second
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	buf := make([]byte, 1)

	for {
		select {
		case <-m.ctx.Done():
			return nil
		case <-ticker.C:
			// Try to read with timeout
			_ = m.conn.SetReadDeadline(time.Now().Add(readTimeout))
			n, err := m.conn.Read(buf)

			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout is expected, connection is still alive
					// Reset deadline
					_ = m.conn.SetReadDeadline(time.Time{})
					continue
				}
				// Real error - connection is broken
				return fmt.Errorf("read error: %w", err)
			}

			if n > 0 {
				// Unexpected data, but connection is alive
				_ = m.conn.SetReadDeadline(time.Time{})
			}
		}
	}
}

// IsConnected reports whether the monitor currently has an active connection.
func (m *TCPKeepaliveMonitor) IsConnected() bool {
	return m.conn != nil
}
