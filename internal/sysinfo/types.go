// Package sysinfo collects and provides system resource information.
package sysinfo

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/alerts"
	"github.com/EdgarOrtegaRamirez/syspulse/internal/osutil"
)

// Info contains a snapshot of system resource information.
type Info struct {
	Timestamp time.Time   `json:"timestamp"`
	Hostname  string      `json:"hostname"`
	Platform  string      `json:"platform"`
	Version   string      `json:"version"`
	Kernel    string      `json:"kernel"`
	Uptime    string      `json:"uptime"`
	CPU       CPUInfo     `json:"cpu"`
	Memory    MemoryInfo  `json:"memory"`
	Disks     []DiskInfo  `json:"disks"`
	Networks  []NetInfo   `json:"networks"`
	Processes ProcessInfo `json:"processes"`
}

// CPUInfo holds CPU-related information.
type CPUInfo struct {
	Model         string    `json:"model"`
	Frequency     float64   `json:"frequency_mhz"`
	PhysicalCores int       `json:"physical_cores"`
	Cores         int       `json:"logical_cores"`
	UsagePercent  float64   `json:"usage_percent"`
	Load1         float64   `json:"load_1"`
	Load5         float64   `json:"load_5"`
	Load15        float64   `json:"load_15"`
	PerCoreUsage  []float64 `json:"per_core_usage,omitempty"`
	TopProcesses  []CpuProc `json:"top_processes,omitempty"`
}

// CpuProc is a process CPU usage snapshot.
type CpuProc struct {
	PID    int     `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu_percent"`
	Memory float64 `json:"memory_percent"`
	RSS    uint64  `json:"rss_bytes"`
	State  string  `json:"state"`
}

// MemoryInfo holds memory-related information.
type MemoryInfo struct {
	Total        uint64     `json:"total_bytes"`
	Used         uint64     `json:"used_bytes"`
	Free         uint64     `json:"free_bytes"`
	Available    uint64     `json:"available_bytes"`
	Cached       uint64     `json:"cached_bytes"`
	Buffers      uint64     `json:"buffers_bytes"`
	UsagePercent float64    `json:"usage_percent"`
	SwapTotal    uint64     `json:"swap_total_bytes"`
	SwapUsed     uint64     `json:"swap_used_bytes"`
	SwapPercent  float64    `json:"swap_percent"`
	TopProcesses []SwapProc `json:"top_swap_processes,omitempty"`
}

// SwapProc is a process swap usage snapshot.
type SwapProc struct {
	Name string `json:"name"`
	Swap uint64 `json:"swap_bytes"`
}

// DiskInfo holds disk partition information.
type DiskInfo struct {
	Mount        string  `json:"mount"`
	Device       string  `json:"device"`
	Filesystem   string  `json:"filesystem"`
	Type         string  `json:"type"`
	Driver       string  `json:"driver"`
	Total        uint64  `json:"total_bytes"`
	Used         uint64  `json:"used_bytes"`
	Free         uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	IORead       uint64  `json:"io_read_bytes_per_sec"`
	IOWrite      uint64  `json:"io_write_bytes_per_sec"`
}

// NetInfo holds network interface information.
type NetInfo struct {
	Name        string `json:"name"`
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
	State       string `json:"state"`
	Speed       string `json:"speed"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsSent  uint64 `json:"errors_sent"`
	ErrorsRecv  uint64 `json:"errors_recv"`
	DropsSent   uint64 `json:"drops_sent"`
	DropsRecv   uint64 `json:"drops_recv"`
}

// ProcessInfo holds process statistics.
type ProcessInfo struct {
	Total     int       `json:"total"`
	Running   int       `json:"running"`
	Sleeping  int       `json:"sleeping"`
	TopCPU    []CpuProc `json:"top_cpu"`
	TopMemory []CpuProc `json:"top_memory"`
}

// New creates a new Info snapshot of the current system state.
func New() (*Info, error) {
	info := &Info{
		Timestamp: time.Now(),
	}

	hostname, _ := os.Hostname()
	info.Hostname = hostname

	info.Platform = runtime.GOOS
	info.Version = runtime.GOARCH

	// Collect all subsystems
	var errs []string
	if err := collectCPU(info); err != nil {
		errs = append(errs, err.Error())
	}
	if err := collectMemory(info); err != nil {
		errs = append(errs, err.Error())
	}
	if err := collectDisks(info); err != nil {
		errs = append(errs, err.Error())
	}
	if err := collectNetwork(info); err != nil {
		errs = append(errs, err.Error())
	}
	if err := collectProcesses(info); err != nil {
		errs = append(errs, err.Error())
	}

	if err := collectSystemInfo(info); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 && len(info.Disks) == 0 && len(info.Networks) == 0 {
		return nil, fmt.Errorf("failed to collect system info: %s", strings.Join(errs, "; "))
	}

	return info, nil
}

// CheckHealth evaluates the system info against thresholds and returns warnings.
func CheckHealth(info *Info, config *alerts.Config) []string {
	var health []string
	if config == nil {
		config = &alerts.Config{}
	}

	if config.CPU > 0 && info.CPU.UsagePercent > float64(config.CPU) {
		health = append(health, fmt.Sprintf("CPU usage at %.1f%% (threshold: %d%%)", info.CPU.UsagePercent, config.CPU))
	}

	if config.Memory > 0 && info.Memory.UsagePercent > float64(config.Memory) {
		health = append(health, fmt.Sprintf("Memory usage at %.1f%% (threshold: %d%%)", info.Memory.UsagePercent, config.Memory))
	}

	if config.Disk > 0 {
		for _, d := range info.Disks {
			if d.UsagePercent > float64(config.Disk) {
				health = append(health, fmt.Sprintf("Disk %s at %.1f%% (threshold: %d%%)", d.Mount, d.UsagePercent, config.Disk))
			}
		}
	}

	if config.Load > 0 && info.CPU.Cores > 0 {
		loadRatio := info.CPU.Load1 / float64(info.CPU.Cores)
		if loadRatio > config.Load {
			health = append(health, fmt.Sprintf("Load average %.2f is %.1fx the number of cores (threshold: %.1fx)", info.CPU.Load1, loadRatio, config.Load))
		}
	}

	if config.ProcessZombie > 0 && info.Processes.Running > config.ProcessZombie {
		health = append(health, fmt.Sprintf("Running processes at %d (threshold: %d)", info.Processes.Running, config.ProcessZombie))
	}

	return health
}

func collectSystemInfo(info *Info) error {
	info.Kernel = osutil.KernelVersion()
	if up, err := osutil.UptimeSeconds(); err == nil {
		info.Uptime = uptimeString(up)
	}
	return nil
}

func uptimeString(seconds float64) string {
	d := int(seconds) / 86400
	h := (int(seconds) % 86400) / 3600
	m := (int(seconds) % 3600) / 60
	if d > 0 {
		return "up " + fmt.Sprintf("%dd %dh %dm", d, h, m)
	}
	return fmt.Sprintf("up %dh %dm", h, m)
}
