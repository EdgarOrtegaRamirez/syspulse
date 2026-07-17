# Syspulse — AI Agent Guide

## Project Overview
Syspulse is a Go CLI tool for comprehensive system resource monitoring on Linux. It collects CPU, memory, disk, network, and process information from /proc filesystem and presents it in human-readable or JSON format.

## Build & Test
```bash
go build ./...
go build -o syspulse ./cmd/syspulse/
go test ./...
go vet ./...
```

## Project Structure
```
syspulse/
├── cmd/syspulse/main.go    # CLI entry point
├── internal/
│   ├── cmd/root.go         # CLI commands and formatting
│   ├── sysinfo/             # System info collectors
│   │   ├── types.go         # Data structures
│   │   ├── cpu.go           # CPU collection
│   │   ├── memory.go        # Memory collection
│   │   ├── disk.go          # Disk collection
│   │   ├── network.go       # Network collection
│   │   └── process.go       # Process collection
│   ├── osutil/proc.go      # Low-level /proc reading
│   ├── alerts/config.go    # Alert threshold config
│   └── history/store.go    # Historical data storage
├── README.md
├── LICENSE
└── go.mod
```

## Key Design Decisions
- All data read from /proc filesystem — no external dependencies
- No root access required for reading system stats
- JSON output for CI/CD integration
- Configurable alert thresholds with exit codes
- Minimal dependencies (stdlib only)

## Dependencies
- Go 1.24+
- No external dependencies (stdlib only)