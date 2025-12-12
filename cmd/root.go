package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	logPath string
)

var rootCmd = &cobra.Command{
	Use:   "network-monitor",
	Short: "Network Stability Monitor",
	Long: `A comprehensive network stability monitoring tool that tracks:
- Interface up/down events
- IP address changes
- Gateway and routing changes
- DNS server changes
- Internet connectivity via persistent TCP keepalive
- Watchdog checks (DNS, HTTP, captive portal detection)`,
}

// Execute runs the root Cobra command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logPath, "log-path", "l", "",
		"Path to log file (default: $HOME/.network-monitor/network-monitor.log)")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(logCmd)
}

func getLogPath() string {
	if logPath != "" {
		return logPath
	}

	exePath, err := os.Executable()
	if err != nil {
		return "network-monitor.log"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "network-monitor.log")
}
