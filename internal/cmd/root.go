// Package cmd provides the CLI root command.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/alerts"
	"github.com/EdgarOrtegaRamirez/syspulse/internal/history"
	"github.com/EdgarOrtegaRamirez/syspulse/internal/sysinfo"
)

// Execute runs the CLI entry point.
func Execute() error {
	if len(os.Args) < 2 {
		return printHelp()
	}

	switch os.Args[1] {
	case "dashboard":
		return cmdDashboard()
	case "cpu":
		return cmdCpu()
	case "memory":
		return cmdMemory()
	case "disk":
		return cmdDisk()
	case "network":
		return cmdNetwork()
	case "processes":
		return cmdProcesses()
	case "report":
		return cmdReport()
	case "alerts":
		return cmdAlerts()
	case "history":
		return cmdHistory()
	case "help", "--help", "-h":
		return printHelp()
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printHelp() error {
	fmt.Print(`Syspulse — System Resource Pulse

A comprehensive system resource monitoring CLI tool.

Usage:
  syspulse <command> [options]

Commands:
  dashboard   Show a comprehensive system overview
  cpu         Show detailed CPU usage and statistics
  memory      Show memory usage breakdown
  disk        Show disk usage and I/O statistics
  network     Show network interface statistics
  processes   Show top processes by resource usage
  report      Generate a full system report (JSON or text)
  alerts      Check system against configured thresholds
  history     Show historical resource usage trends
  help        Show this help message

Examples:
  syspulse dashboard                # Full system overview
  syspulse report -o json           # Full report as JSON
  syspulse alerts --json cpu=80     # Alert on CPU > 80%
  syspulse history --hours 24       # 24-hour resource trends
  syspulse process 1                # Show details for PID 1
`)
	return nil
}

func cmdDashboard() error {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    SYSULSE — SYSTEM PULSE                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	fmt.Printf("  Host:     %s\n", info.Hostname)
	fmt.Printf("  OS:       %s %s\n", info.Platform, info.Version)
	fmt.Printf("  Kernel:   %s\n", info.Kernel)
	fmt.Printf("  Uptime:   %s\n", info.Uptime)
	fmt.Println()

	// CPU summary
	fmt.Println("  CPU")
	fmt.Printf("    Cores:     %d (%d physical)\n", info.CPU.Cores, info.CPU.PhysicalCores)
	fmt.Printf("    Load:      %.2f, %.2f, %.2f (1/5/15 min)\n", info.CPU.Load1, info.CPU.Load5, info.CPU.Load15)
	fmt.Printf("    Usage:     %6.1f%%\n", info.CPU.UsagePercent)
	fmt.Printf("    Model:     %s\n", info.CPU.Model)
	fmt.Println()

	// Memory summary
	fmt.Println("  Memory")
	fmt.Printf("    Total:     %s\n", formatBytes(info.Memory.Total))
	fmt.Printf("    Used:      %s (%.1f%%)\n", formatBytes(info.Memory.Used), info.Memory.UsagePercent)
	fmt.Printf("    Free:      %s\n", formatBytes(info.Memory.Free))
	fmt.Printf("    Cached:    %s\n", formatBytes(info.Memory.Cached))
	fmt.Println()

	// Disk summary
	fmt.Println("  Disks")
	for _, d := range info.Disks {
		fmt.Printf("    %s %s (%s used of %s, %.1f%%)\n", d.Mount, formatBytes(d.Used), formatBytes(d.Used), formatBytes(d.Total), d.UsagePercent)
	}
	fmt.Println()

	// Network summary
	fmt.Println("  Network")
	for _, n := range info.Networks {
		fmt.Printf("    %s: %s sent, %s received\n", n.Name, formatBytes(n.BytesSent), formatBytes(n.BytesRecv))
	}
	fmt.Println()

	// Processes
	fmt.Printf("  Processes: %d total, %d running, %d sleeping\n", info.Processes.Total, info.Processes.Running, info.Processes.Sleeping)
	fmt.Println()

	// System health
	health := sysinfo.CheckHealth(info, nil)
	if len(health) > 0 {
		fmt.Println("  ⚠  Health Warnings:")
		for _, w := range health {
			fmt.Printf("    - %s\n", w)
		}
	} else {
		fmt.Println("  ✓  System health: OK")
	}

	return nil
}

func cmdCpu() error {
	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	cpu := info.CPU
	fmt.Printf("CPU Usage: %.1f%%\n", cpu.UsagePercent)
	fmt.Printf("Model:     %s\n", cpu.Model)
	fmt.Printf("Frequency: %.0f MHz (current)\n", cpu.Frequency)
	fmt.Printf("Cores:     %d physical, %d logical\n", cpu.PhysicalCores, cpu.Cores)
	fmt.Printf("Load Avg:  %.2f (1m) %.2f (5m) %.2f (15m)\n", cpu.Load1, cpu.Load5, cpu.Load15)

	if len(cpu.PerCoreUsage) > 0 {
		fmt.Println("\nPer-core usage:")
		for i, usage := range cpu.PerCoreUsage {
			bar := ""
			for j := 0; j < int(usage/5); j++ {
				bar += "█"
			}
			fmt.Printf("  Core %d: %6.1f%% %s\n", i, usage, bar)
		}
	}

	fmt.Println("\nPer-process CPU usage (top 10):")
	processes := cpu.TopProcesses
	if len(processes) > 10 {
		processes = processes[:10]
	}
	for i, p := range processes {
		fmt.Printf("  %2d. %-20s CPU: %6.1f%%  MEM: %6.1f%%\n", i+1, p.Name, p.CPU, p.Memory)
	}

	return nil
}

func cmdMemory() error {
	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	mem := info.Memory
	fmt.Printf("Memory Usage: %.1f%%\n", mem.UsagePercent)
	fmt.Printf("Total:     %s\n", formatBytes(mem.Total))
	fmt.Printf("Used:      %s\n", formatBytes(mem.Used))
	fmt.Printf("Free:      %s\n", formatBytes(mem.Free))
	fmt.Printf("Available: %s\n", formatBytes(mem.Available))
	fmt.Printf("Cached:    %s\n", formatBytes(mem.Cached))
	fmt.Printf("Buffers:   %s\n", formatBytes(mem.Buffers))
	fmt.Printf("Swap:      %s / %s (%.1f%%)\n", formatBytes(mem.SwapUsed), formatBytes(mem.SwapTotal), mem.SwapPercent)

	return nil
}

func cmdDisk() error {
	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	for _, d := range info.Disks {
		bar := ""
		for i := 0; i < int(d.UsagePercent/2); i++ {
			bar += "█"
		}
		remaining := 50 - len(bar)
		if remaining > 0 {
			bar += "░"
		}

		fmt.Printf("%s %s %s (%s used of %s)\n", d.Mount, formatBytes(d.Used), formatBytes(d.Free), formatBytes(d.Used), formatBytes(d.Total))
		fmt.Printf("  Usage: %6.1f%% [%s]\n", d.UsagePercent, bar)
		fmt.Printf("  FS: %s  Driver: %s  Type: %s\n", d.Filesystem, d.Driver, d.Type)

		if d.IORead != 0 || d.IOWrite != 0 {
			fmt.Printf("  IO: read %s/s  write %s/s\n", formatBytes(d.IORead), formatBytes(d.IOWrite))
		}
		fmt.Println()
	}

	return nil
}

func cmdNetwork() error {
	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	for _, n := range info.Networks {
		fmt.Printf("Interface: %s\n", n.Name)
		fmt.Printf("  IP:     %s\n", n.IP)
		fmt.Printf("  MAC:    %s\n", n.MAC)
		fmt.Printf("  State:  %s\n", n.State)
		fmt.Printf("  Speed:  %s\n", n.Speed)
		fmt.Printf("  Traffic: sent %s  received %s\n", formatBytes(n.BytesSent), formatBytes(n.BytesRecv))
		fmt.Printf("  Packets: sent %d  received %d\n", n.PacketsSent, n.PacketsRecv)
		fmt.Printf("  Errors:  sent %d  received %d\n", n.ErrorsSent, n.ErrorsRecv)
		fmt.Printf("  Drops:   sent %d  received %d\n", n.DropsSent, n.DropsRecv)
		fmt.Println()
	}

	return nil
}

func cmdProcesses() error {
	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	fmt.Printf("Total: %d  Running: %d  Sleeping: %d\n", info.Processes.Total, info.Processes.Running, info.Processes.Sleeping)
	fmt.Println()

	fmt.Println("Top processes by CPU:")
	for i, p := range info.Processes.TopCPU[:min(10, len(info.Processes.TopCPU))] {
		fmt.Printf("  %2d. PID=%5d  %-20s CPU: %5.1f%%  MEM: %5.1f%%  RSS: %s  State: %s\n",
			i+1, p.PID, p.Name, p.CPU, p.Memory, formatBytes(p.RSS), p.State)
	}

	fmt.Println()
	fmt.Println("Top processes by Memory:")
	for i, p := range info.Processes.TopMemory[:min(10, len(info.Processes.TopMemory))] {
		fmt.Printf("  %2d. PID=%5d  %-20s MEM: %5.1f%%  RSS: %s  CPU: %5.1f%%\n",
			i+1, p.PID, p.Name, p.Memory, formatBytes(p.RSS), p.CPU)
	}

	return nil
}

func cmdReport() error {
	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	health := sysinfo.CheckHealth(info, nil)

	report := map[string]interface{}{
		"timestamp": info.Timestamp,
		"hostname":  info.Hostname,
		"platform":  info.Platform,
		"kernel":    info.Kernel,
		"uptime":    info.Uptime,
		"cpu": map[string]interface{}{
			"usage_percent":  info.CPU.UsagePercent,
			"model":          info.CPU.Model,
			"frequency_mhz":  info.CPU.Frequency,
			"physical_cores": info.CPU.PhysicalCores,
			"logical_cores":  info.CPU.Cores,
			"load_1":         info.CPU.Load1,
			"load_5":         info.CPU.Load5,
			"load_15":        info.CPU.Load15,
		},
		"memory": map[string]interface{}{
			"total":     formatBytes(info.Memory.Total),
			"used":      formatBytes(info.Memory.Used),
			"free":      formatBytes(info.Memory.Free),
			"available": formatBytes(info.Memory.Available),
			"cached":    formatBytes(info.Memory.Cached),
			"usage_pct": info.Memory.UsagePercent,
		},
		"disks":   make([]map[string]interface{}, len(info.Disks)),
		"network": make([]map[string]interface{}, len(info.Networks)),
		"processes": map[string]interface{}{
			"total":    info.Processes.Total,
			"running":  info.Processes.Running,
			"sleeping": info.Processes.Sleeping,
		},
		"health": health,
	}

	for i, d := range info.Disks {
		report["disks"].([]map[string]interface{})[i] = map[string]interface{}{
			"mount":     d.Mount,
			"total":     formatBytes(d.Total),
			"used":      formatBytes(d.Used),
			"free":      formatBytes(d.Free),
			"usage_pct": d.UsagePercent,
			"fs":        d.Filesystem,
		}
	}

	for i, n := range info.Networks {
		report["network"].([]map[string]interface{})[i] = map[string]interface{}{
			"name":       n.Name,
			"ip":         n.IP,
			"bytes_sent": formatBytes(n.BytesSent),
			"bytes_recv": formatBytes(n.BytesRecv),
		}
	}

	fmt.Println(toJSON(report))

	return nil
}

func cmdAlerts() error {
	// Parse alert thresholds from arguments like --json cpu=80 mem=90 disk=85
	thresholds, err := parseThresholds()
	if err != nil {
		return fmt.Errorf("invalid threshold: %w", err)
	}

	info, err := sysinfo.New()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	health := sysinfo.CheckHealth(info, &alerts.Config{
		CPU:           thresholds["cpu"],
		Memory:        thresholds["memory"],
		Disk:          thresholds["disk"],
		Load:          float64(thresholds["load"]),
		ProcessZombie: thresholds["zombie"],
	})

	if len(health) == 0 {
		fmt.Println("✓  All checks passed — system healthy")
		return nil
	}

	fmt.Printf("⚠  %d health check(s) failed:\n", len(health))
	for _, h := range health {
		fmt.Printf("  - %s\n", h)
	}

	return fmt.Errorf("%d health check(s) failed", len(health))
}

func cmdHistory() error {
	h := history.New()
	records, err := h.Load()
	if err != nil {
		return fmt.Errorf("no history available yet")
	}

	if len(records) == 0 {
		fmt.Println("No historical data. Run 'syspulse report' to start collecting.")
		return nil
	}

	// Show recent history
	fmt.Printf("Historical resource usage (%d records):\n", len(records))
	for i, r := range records {
		if i >= 20 {
			fmt.Printf("  ... and %d more\n", len(records)-20)
			break
		}
		fmt.Printf("  %s  CPU: %5.1f%%  MEM: %5.1f%%  DISK: %5.1f%%\n",
			r.Timestamp, r.CPU, r.Memory, r.Disk)
	}

	// Show trends
	if len(records) >= 2 {
		first := records[0]
		last := records[len(records)-1]
		fmt.Println("\nTrends (first → last):")
		fmt.Printf("  CPU:  %.1f%% → %.1f%% (%s)\n", first.CPU, last.CPU, trend(first.CPU, last.CPU))
		fmt.Printf("  MEM:  %.1f%% → %.1f%% (%s)\n", first.Memory, last.Memory, trend(first.Memory, last.Memory))
		fmt.Printf("  DISK: %.1f%% → %.1f%% (%s)\n", first.Disk, last.Disk, trend(first.Disk, last.Disk))
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatBytes(b uint64) string {
	u := uint64(1024)
	if b < u {
		return strconv.FormatUint(b, 10) + " B"
	}
	w := u * u
	if b < w {
		return fmt.Sprintf("%.1f KB", float64(b)/float64(u))
	}
	m := w * u
	if b < m {
		return fmt.Sprintf("%.1f MB", float64(b)/float64(w))
	}
	g := m * u
	if b < g {
		return fmt.Sprintf("%.1f GB", float64(b)/float64(m))
	}
	return fmt.Sprintf("%.1f TB", float64(b)/float64(g))
}

func toJSON(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func parseThresholds() (map[string]int, error) {
	thresholds := map[string]int{
		"cpu": 80, "memory": 80, "disk": 85, "zombie": 10,
	}
	for _, arg := range os.Args[2:] {
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		eqIdx := strings.Index(arg, "=")
		if eqIdx < 0 {
			continue
		}
		key := strings.TrimPrefix(arg[:eqIdx], "--")
		valStr := arg[eqIdx+1:]
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %s", key, valStr)
		}
		thresholds[key] = val
	}
	return thresholds, nil
}

func trend(before, after float64) string {
	diff := after - before
	if diff > 0.1 {
		return "▲"
	}
	if diff < -0.1 {
		return "▼"
	}
	return "→"
}
