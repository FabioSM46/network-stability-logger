# Network Stability Logger

A comprehensive network stability monitoring tool that tracks system-level network events and maintains persistent TCP connections to detect Internet failures.

## Features

### 1. System Events Monitor (Fast & Accurate)
Monitors low-level network changes using platform-specific APIs:

**Linux** (via `netlink`):
- Ethernet/WiFi link up/down
- IP address changes (add/remove)
- Gateway changes
- Routing table changes
### Prerequisites
- Go 1.21 or higher
- Linux, macOS, or Windows

### Build from source

```bash
# Clone the repository
git clone https://github.com/FabioSM46/network-stability-logger.git
cd network-stability-logger

# Install dependencies
go mod download

# Build
go build -o network-monitor

# Optional: Install to $GOPATH/bin
go install
```

**macOS** (via routing socket):
- Interface state changes

### Start Monitoring
- Address changes
- Route modifications
# Start in foreground mode
./network-monitor start -f

# Start with custom log path
./network-monitor start -f --log-path /var/log/network-monitor.log
**Windows** (via IP Helper API):
- Interface status changes
- IP configuration changes
- Route table modifications

### View Logs

```bash
# Show last 50 lines (default)
./network-monitor log

# Show last 100 lines
./network-monitor log -n 100

# Show all logs
./network-monitor log -n 0

# Follow logs in real-time (like tail -f)
./network-monitor log -f

# Filter by category
./network-monitor log -F TCP        # Show only TCP keepalive events
./network-monitor log -F LINK       # Show only link events
./network-monitor log -F WATCHDOG   # Show only watchdog checks
```

**Available filters**:
- `SYSTEM` - System startup/shutdown messages
- `LINK` - Interface up/down events
- `ADDRESS` - IP address changes
- `ROUTE` - Routing table changes
- `DNS` - DNS configuration changes
- `TCP` - TCP keepalive connection status
- `WATCHDOG` - Periodic check results
- `MONITOR` - Monitor control messages

### Stop Monitoring

```bash
./network-monitor stop
```

## Example Output

```
[2025-12-12 10:15:32.123] [MONITOR] === Network Stability Monitor Starting ===
[2025-12-12 10:15:32.145] [SYSTEM] Starting Linux netlink monitoring
[2025-12-12 10:15:32.147] [SYSTEM] Found 3 network interfaces
[2025-12-12 10:15:32.147] [SYSTEM]   eth0: UP
[2025-12-12 10:15:32.147] [SYSTEM]   wlan0: DOWN
[2025-12-12 10:15:32.148] [SYSTEM]   Default route via 192.168.1.1
[2025-12-12 10:15:32.149] [TCP] Starting persistent TCP keepalive monitor to 1.1.1.1:443
[2025-12-12 10:15:32.150] [WATCHDOG] Starting watchdog monitor
[2025-12-12 10:15:32.275] [TCP] SUCCESS: Connected to 1.1.1.1:443
[2025-12-12 10:15:32.276] [WATCHDOG] Running periodic checks...
[2025-12-12 10:15:32.277] [WATCHDOG] ✓ Default route exists (via 192.168.1.1)
[2025-12-12 10:15:32.389] [WATCHDOG] ✓ DNS working: www.google.com -> 142.250.185.36 (took 112ms)
[2025-12-12 10:15:32.512] [WATCHDOG] ✓ HTTP working: 200 OK (took 123ms)
[2025-12-12 10:16:05.234] [LINK] Interface wlan0 [device]: UP (flags: up|broadcast|multicast)
[2025-12-12 10:16:06.123] [ADDRESS] IP address ADDED on wlan0: 192.168.1.45/24
[2025-12-12 10:16:06.234] [ROUTE] Route ADDED: 192.168.1.0/24 via direct dev wlan0
```

## Development

### Project Structure

### 2. Persistent TCP Keepalive Connection
- Maintains a single persistent TCP connection to detect Internet failures
- Detects scenarios where link is "up" but:
	- Router has no Internet
	- DNS server is dead
	- Packets drop silently
	- Gateway is reachable but upstream Internet is down

### 3. Watchdog Checks
Runs periodic checks every 30 seconds:
- **Default route verification** - Ensures routing table has default gateway
- **DNS resolution test** - Tests DNS by resolving `www.google.com`
- **HTTP connectivity check** - Performs HEAD request to detect:
	- Internet connectivity
	- Captive portals (redirect detection)
	- Network restrictions

## Architecture

The application runs three concurrent goroutines:

```
┌─────────────────────────────────────────────────────────────┐
│                    Network Monitor                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────┐  ┌──────────────────────┐       │
│  │ System Events        │  │ TCP Keepalive        │       │
│  │ Monitor              │  │ Monitor              │       │
│  │                      │  │                      │       │
│  │ • netlink (Linux)    │  │ • Single persistent  │       │
│  │ • routing socket     │  │   TCP connection     │       │
│  │   (macOS)            │  │ • 30s keepalive      │       │
│  │ • IP Helper          │  │ • Auto-reconnect     │       │
│  │   (Windows)          │  │                      │       │
│  └──────────────────────┘  └──────────────────────┘       │
│                                                              │
│  ┌────────────────────────────────────────────────┐        │
│  │ Watchdog Monitor (every 30s)                   │        │
│  │                                                 │        │
│  │ • Default route check                          │        │
│  │ • DNS resolution test                          │        │
│  │ • HTTP HEAD request                            │        │
│  │ • Captive portal detection                     │        │
│  └────────────────────────────────────────────────┘        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Installation

## Getting Started

### Prerequisites
- Go 1.21 or higher

### Installation

```bash
go mod download
```

### Running

```bash
go run main.go
```

### Building

```bash
go build -o network-stability-logger
```

## Project Structure

- `main.go` - Entry point
- `go.mod` - Module definition

## License

MIT
