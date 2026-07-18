package osutil

import (
	"testing"
)

func TestNumCPUs(t *testing.T) {
	n := NumCPUs()
	if n < 1 {
		t.Errorf("expected at least 1 CPU, got %d", n)
	}
}

func TestKernelVersion(t *testing.T) {
	kv := KernelVersion()
	if kv == "" {
		t.Error("expected non-empty kernel version")
	}
}

func TestLoadAverage(t *testing.T) {
	l1, l5, l15, err := LoadAverage()
	if err != nil {
		t.Skipf("Cannot read /proc/loadavg: %v", err)
	}
	if l1 < 0 || l5 < 0 || l15 < 0 {
		t.Errorf("expected non-negative load averages, got %f %f %f", l1, l5, l15)
	}
}

func TestReadFile(t *testing.T) {
	// Test with an existing file
	content := ReadFile("/proc/uptime")
	if content == "" {
		t.Error("expected non-empty content for /proc/uptime")
	}

	// Test with a nonexistent file
	content = ReadFile("/nonexistent/file")
	if content != "" {
		t.Error("expected empty content for nonexistent file")
	}
}

func TestUptimeSeconds(t *testing.T) {
	secs, err := UptimeSeconds()
	if err != nil {
		t.Skipf("Cannot read /proc/uptime: %v", err)
	}
	if secs < 0 {
		t.Errorf("expected non-negative uptime, got %f", secs)
	}
}
