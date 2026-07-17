# Syspulse 🔍

A comprehensive system resource monitoring CLI tool for Linux. Collects real-time snapshots of CPU, memory, disk, network, and process health — with structured JSON output for CI/CD integration.

## Features

- **System overview** — Single command dashboard with CPU, memory, disk, network, and process stats
- **Per-component detail** — Dedicated commands for CPU, memory, disk, and network with deep breakdowns
- **Health checks** — Configurable thresholds with pass/fail exit codes for CI/CD integration
- **Structured output** — Full JSON report for programmatic use
- **Process analysis** — Top processes by CPU and memory usage
- **Network stats** — Per-interface traffic, packet, and error counters
- **Disk I/O** — Per-device read/write statistics

## Installation

### From source

```bash
go install github.com/EdgarOrtegaRamirez/syspulse/cmd/syspulse@latest
```

### From source (local build)

```bash
git clone https://github.com/EdgarOrtegaRamirez/syspulse
cd syspulse
go build -o syspulse ./cmd/syspulse/
sudo mv syspulse /usr/local/bin/
```

## Quick Start

```bash
# Full system overview
syspulse dashboard

# Detailed CPU report
syspulse cpu

# Memory breakdown
syspulse memory

# Disk usage with I/O stats
syspulse disk

# Network interface statistics
syspulse network

# Top processes
syspulse processes

# Full JSON report (for pipelines)
syspulse report

# Health check with custom thresholds (exits non-zero on failure)
syspulse alerts --cpu=80 --memory=90 --disk=85

# Show historical data
syspulse history
```

## Examples

### CI/CD integration

```bash
#!/bin/bash
# Fail the build if memory usage is above 90%
if syspulse alerts --memory=90 --cpu=90 --disk=90; then
  echo "System healthy"
else
  echo "System health check failed"
  exit 1
fi
```

### JSON report for monitoring

```bash
# Get a structured report
syspulse report > /tmp/syspulse-report.json

# Parse with jq
syspulse report | jq '.cpu.usage_percent'
syspulse report | jq '.disks[] | {mount, usage_pct}'
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `dashboard` | Show a comprehensive system overview |
| `cpu` | Show detailed CPU usage and statistics |
| `memory` | Show memory usage breakdown |
| `disk` | Show disk usage and I/O statistics |
| `network` | Show network interface statistics |
| `processes` | Show top processes by resource usage |
| `report` | Generate a full system report (JSON) |
| `alerts` | Check system against configured thresholds |
| `history` | Show historical resource usage trends |
| `help` | Show this help message |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success / Healthy |
| 1 | Error / Thresholds exceeded |

The `alerts` command exits non-zero when any configured threshold is exceeded, making it ideal for CI/CD pipelines.

## Supported Platforms

- Linux (all distributions using /proc filesystem)
- Requires no root access for reading system stats
- Requires no root access for reading disk I/O stats

## Development

```bash
go build ./...
go test ./...
go vet ./...
```

## License

MIT