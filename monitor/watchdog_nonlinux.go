//go:build !linux

package monitor

// checkDefaultRouteLinux is unused on non-Linux platforms; defined to satisfy builds.
func (m *WatchdogMonitor) checkDefaultRouteLinux() {}
