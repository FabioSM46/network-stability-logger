//go:build !linux

package monitor

// applyPlatformKeepalive is a no-op on non-Linux platforms.
func applyPlatformKeepalive(fd uintptr) {}
