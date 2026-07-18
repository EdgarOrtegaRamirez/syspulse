package sysinfo

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/alerts"
)

func TestCheckHealth_NoThresholds(t *testing.T) {
	info := &Info{
		CPU:    CPUInfo{UsagePercent: 50},
		Memory: MemoryInfo{UsagePercent: 50},
		Disks: []DiskInfo{
			{Mount: "/", UsagePercent: 50},
		},
	}
	health := CheckHealth(info, nil)
	if len(health) != 0 {
		t.Errorf("expected no warnings with nil config, got %d", len(health))
	}
}

func TestCheckHealth_CPUExceeded(t *testing.T) {
	info := &Info{
		CPU: CPUInfo{UsagePercent: 90},
	}
	health := CheckHealth(info, &alerts.Config{CPU: 80})
	if len(health) != 1 {
		t.Errorf("expected 1 warning, got %d", len(health))
	}
}

func TestCheckHealth_MemoryExceeded(t *testing.T) {
	info := &Info{
		Memory: MemoryInfo{UsagePercent: 95},
	}
	health := CheckHealth(info, &alerts.Config{Memory: 90})
	if len(health) != 1 {
		t.Errorf("expected 1 warning, got %d", len(health))
	}
}

func TestCheckHealth_DiskExceeded(t *testing.T) {
	info := &Info{
		Disks: []DiskInfo{
			{Mount: "/data", UsagePercent: 95},
		},
	}
	health := CheckHealth(info, &alerts.Config{Disk: 90})
	if len(health) != 1 {
		t.Errorf("expected 1 warning, got %d", len(health))
	}
}

func TestCheckHealth_MultipleWarnings(t *testing.T) {
	info := &Info{
		CPU:    CPUInfo{UsagePercent: 95},
		Memory: MemoryInfo{UsagePercent: 95},
		Disks:  []DiskInfo{{Mount: "/", UsagePercent: 95}},
	}
	health := CheckHealth(info, &alerts.Config{CPU: 80, Memory: 80, Disk: 80})
	if len(health) != 3 {
		t.Errorf("expected 3 warnings, got %d", len(health))
	}
}
