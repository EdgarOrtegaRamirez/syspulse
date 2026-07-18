// Package osutil provides low-level OS info reading.
package osutil

import (
	"os"
	"strconv"
	"strings"
)

// KernelVersion returns the OS kernel version string.
func KernelVersion() string {
	data, _ := os.ReadFile("/proc/version")
	return strings.TrimSpace(string(data))
}

// UptimeSeconds returns the system uptime in seconds.
func UptimeSeconds() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, os.ErrInvalid
	}
	return strconv.ParseFloat(strings.TrimSpace(fields[0]), 64)
}

// LoadAverage returns the 1, 5, 15 minute load averages.
func LoadAverage() (float64, float64, float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, os.ErrInvalid
	}
	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	l15, _ := strconv.ParseFloat(fields[2], 64)
	return l1, l5, l15, nil
}

// NumCPUs returns the number of logical CPUs.
func NumCPUs() int {
	data, _ := os.ReadFile("/proc/cpuinfo")
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}
	if count == 0 {
		count = 1
	}
	return count
}

// ReadFile reads a file and returns its content trimmed.
func ReadFile(path string) string {
	data, _ := os.ReadFile(path)
	return strings.TrimSpace(string(data))
}

// ReadFileBytes reads a file and returns its content.
func ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
