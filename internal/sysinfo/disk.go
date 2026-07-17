// Package sysinfo provides system disk information.
package sysinfo

import (
	"os"
	"bufio"
	"strconv"
	"strings"
	"syscall"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/osutil"
)

func collectDisks(info *Info) error {
	seen := make(map[string]bool)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		mount := fields[1]
		fstype := fields[2]

		// Skip virtual filesystems
		if isVirtualFS(fstype) {
			continue
		}

		if seen[mount] {
			continue
		}
		seen[mount] = true

		var disk DiskInfo
		disk.Mount = mount
		disk.Filesystem = fstype

		// Get device
		if len(fields) > 0 {
			dev := fields[0]
			dev = strings.TrimPrefix(dev, "/dev/")
			dev = stripPartition(dev)
			disk.Device = fields[0]
			disk.Driver = dev
		}

		// Get size via statfs
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err == nil {
			disk.Total = stat.Blocks * uint64(stat.Bsize)
			disk.Free = stat.Bavail * uint64(stat.Bsize)
			disk.Used = disk.Total - disk.Free
			if disk.Total > 0 {
				disk.UsagePercent = float64(disk.Used) / float64(disk.Total) * 100
			}
		}

		// Get IO stats
		disk.IORead, disk.IOWrite = getDiskIOStats(disk.Device)

		info.Disks = append(info.Disks, disk)
	}

	return scanner.Err()
}

func isVirtualFS(fstype string) bool {
	virtual := []string{
		"proc", "sysfs", "devpts", "tmpfs", "devtmpfs", "cgroup",
		"securityfs", "pstore", "debugfs", "hugetlbfs", "mqueue",
		"configfs", "binfmt_misc", "autofs", "rpc_pipefs", "nsfs",
		"efivarfs", "tracefs", "fusectl", "bpf", "fuse.lxcfs",
	}
	for _, v := range virtual {
		if v == fstype {
			return true
		}
	}
	return false
}

func getDiskIOStats(dev string) (uint64, uint64) {
	name := stripPartition(dev)
	name = strings.TrimPrefix(name, "/dev/")

	data, err := osutil.ReadFileBytes("/proc/diskstats")
	if err != nil {
		return 0, 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 14 && fields[2] == name {
			reads, _ := strconv.ParseUint(fields[5], 10, 64)
			writes, _ := strconv.ParseUint(fields[9], 10, 64)
			return reads * 512, writes * 512
		}
	}
	return 0, 0
}

func stripPartition(dev string) string {
	for len(dev) > 0 {
		last := dev[len(dev)-1]
		if last >= '0' && last <= '9' {
			dev = dev[:len(dev)-1]
		} else {
			break
		}
	}
	return dev
}