// Package sysinfo provides system memory information.
package sysinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func collectMemory(info *Info) error {
	mem := &info.Memory
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return err
	}
	defer file.Close()

	maps := make(map[string]uint64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			maps[fields[0]] = val * 1024 // Convert kB to bytes
		}
	}

	mem.Total = maps["MemTotal:"]
	mem.Free = maps["MemFree:"]
	mem.Available = maps["MemAvailable:"]
	mem.Cached = maps["Buffers:"]
	mem.Buffers = maps["Buffers:"]
	// Also add Active(file) and Inactive(file) for cached
	if active, ok := maps["Active(file):"]; ok {
		mem.Cached += active
	}
	if inactive, ok := maps["Inactive(file):"]; ok {
		mem.Cached += inactive
	}
	if sActive, ok := maps["Active(anon):"]; ok {
		mem.Buffers += sActive
	}

	if mem.Available == 0 {
		mem.Available = mem.Free + mem.Cached
	}

	mem.Used = mem.Total - mem.Free - mem.Cached - mem.Buffers
	if mem.Used < 0 {
		mem.Used = mem.Total - mem.Available
	}
	if mem.Total > 0 {
		mem.UsagePercent = float64(mem.Used) / float64(mem.Total) * 100
	}

	// Swap
	mem.SwapTotal = maps["SwapTotal:"]
	mem.SwapUsed = maps["SwapTotal:"] - maps["SwapFree:"]
	if mem.SwapTotal > 0 {
		mem.SwapPercent = float64(mem.SwapUsed) / float64(mem.SwapTotal) * 100
	}

	return nil
}
