package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Display network monitor logs",
	Long:  `Display logs from the network stability monitor.`,
	RunE:  runLog,
}

func init() {
	logCmd.Flags().IntP("lines", "n", 50, "Number of lines to display (0 for all)")
	logCmd.Flags().BoolP("follow", "f", false, "Follow log output (like tail -f)")
	logCmd.Flags().StringP("filter", "F", "", "Filter logs by category (SYSTEM, LINK, ADDRESS, ROUTE, DNS, TCP, WATCHDOG, MONITOR)")
}

func runLog(cmd *cobra.Command, args []string) error {
	lines, _ := cmd.Flags().GetInt("lines")
	follow, _ := cmd.Flags().GetBool("follow")
	filter, _ := cmd.Flags().GetString("filter")

	logFile := getLogPath()

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Errorf("log file does not exist: %s", logFile)
	}

	if follow {
		return followLog(logFile, filter)
	}

	return displayLog(logFile, lines, filter)
}

func displayLog(logFile string, numLines int, filter string) error {
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read all lines
	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if filter != "" {
			if !strings.Contains(line, fmt.Sprintf("[%s]", filter)) {
				continue
			}
		}
		allLines = append(allLines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	// Display last N lines
	start := 0
	if numLines > 0 && len(allLines) > numLines {
		start = len(allLines) - numLines
	}

	for i := start; i < len(allLines); i++ {
		fmt.Println(allLines[i])
	}

	return nil
}

func followLog(logFile string, filter string) error {
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Seek to end of file
	_, _ = file.Seek(0, 2)

	fmt.Printf("Following log file: %s\n", logFile)
	if filter != "" {
		fmt.Printf("Filtering by: [%s]\n", filter)
	}
	fmt.Println("Press Ctrl+C to stop...")

	scanner := bufio.NewScanner(file)
	for {
		if scanner.Scan() {
			line := scanner.Text()
			if filter != "" {
				if !strings.Contains(line, fmt.Sprintf("[%s]", filter)) {
					continue
				}
			}
			fmt.Println(line)
		} else {
			// No new data, wait a bit
			time.Sleep(100 * time.Millisecond)

			// Check if file still exists
			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				return fmt.Errorf("log file was removed")
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading log file: %w", err)
		}
	}
}
