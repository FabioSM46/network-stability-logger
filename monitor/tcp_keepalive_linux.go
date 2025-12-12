//go:build linux

package monitor

import "syscall"

// applyPlatformKeepalive sets aggressive keepalive parameters on Linux.
func applyPlatformKeepalive(fd uintptr) {
	// 10s between probes, 3 failed probes before drop
	_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, 10)
	_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, 3)
}
