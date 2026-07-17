// Package sysinfo provides system CPU information.
package sysinfo

import (
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/osutil"
)

func collectCPU(info *Info) error {
	cpu := &info.CPU
	cpu.PhysicalCores = osutil.NumCPUs() // fallback to logical cores
	cpu.Cores = osutil.NumCPUs()

	// Read model
	model := osutil.ReadFile("/proc/cpuinfo")
	lines := strings.Split(model, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				cpu.Model = strings.TrimSpace(parts[1])
			}
			break
		}
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				cpu.Model = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	// Read frequency
	freqStr := osutil.ReadFile("/proc/cpuinfo")
	for _, line := range strings.Split(freqStr, "\n") {
		if strings.Contains(line, "cpu MHz") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				freq, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err == nil {
					cpu.Frequency = freq
				}
				break
			}
		}
	}

	// Physical cores: count unique core ids
	physCores := make(map[string]bool)
	for _, line := range strings.Split(model, "\n") {
		if strings.HasPrefix(line, "core id") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				physCores[strings.TrimSpace(parts[1])] = true
			}
		}
	}
	if len(physCores) > 0 {
		cpu.PhysicalCores = len(physCores)
	} else {
		cpu.PhysicalCores = cpu.Cores
	}

	// Load averages
	l1, l5, l15, err := osutil.LoadAverage()
	if err == nil {
		cpu.Load1 = l1
		cpu.Load5 = l5
		cpu.Load15 = l15
	}

	// CPU usage from /proc/stat
	cpu.UsagePercent = readCPUUsage()

	return nil
}

func readCPUUsage() float64 {
	data, err := osutil.ReadFileBytes("/proc/stat")
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return 0
			}
			var user, nice, system, idle, iowait, irq, softirq, steal uint64
			user, _ = strconv.ParseUint(fields[1], 10, 64)
			nice, _ = strconv.ParseUint(fields[2], 10, 64)
			system, _ = strconv.ParseUint(fields[3], 10, 64)
			idle, _ = strconv.ParseUint(fields[4], 10, 64)
			iowait, _ = strconv.ParseUint(fields[5], 10, 64)
			irq, _ = strconv.ParseUint(fields[6], 10, 64)
			softirq, _ = strconv.ParseUint(fields[7], 10, 64)

			total := user + nice + system + idle + iowait + irq + softirq
			if steal >= 10 { // simple steal check
				st, _ := strconv.ParseUint(fields[8], 10, 64)
				total += st
			}
			active := total - idle - iowait
			if total == 0 {
				return 0
			}
			return float64(active) / float64(total) * 100
		}
	}
	return 0
}