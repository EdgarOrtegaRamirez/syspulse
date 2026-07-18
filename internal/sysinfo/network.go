// Package sysinfo provides system network information.
package sysinfo

import (
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/osutil"
)

func collectNetwork(info *Info) error {
	// Read /proc/net/dev
	data, err := osutil.ReadFileBytes("/proc/net/dev")
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, ":") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Parse interface name (remove trailing colon)
		name := strings.TrimSuffix(fields[0], ":")
		if name == "Inter-" || name == "Face" {
			continue
		}

		// Parse values: recv_bytes recv_packets recv_errs recv_drop recv_fifo recv_frame recv_compressed recv_mult
		//               sent_bytes sent_packets sent_errs sent_drop sent_fifo sent_colls sent_compressed sent_mult
		if len(fields) < 17 {
			continue
		}

		var net NetInfo
		net.Name = name

		for i, v := range fields[1:] {
			val, _ := strconv.ParseUint(v, 10, 64)
			switch i {
			case 0:
				net.BytesRecv = val
			case 1:
				net.PacketsRecv = val
			case 2:
				net.ErrorsRecv = val
			case 3:
				net.DropsRecv = val
			case 8:
				net.BytesSent = val
			case 9:
				net.PacketsSent = val
			case 10:
				net.ErrorsSent = val
			case 11:
				net.DropsSent = val
			}
		}

		// Get IP and MAC
		net.IP = getInterfaceIP(name)
		net.MAC = getInterfaceMAC(name)
		net.State = getInterfaceState(name)
		net.Speed = getInterfaceSpeed(name)

		info.Networks = append(info.Networks, net)
	}

	return nil
}

func getInterfaceIP(name string) string {
	// Use /proc/net/if_inet6 for IPv6 or /proc/net/fib_trie for IPv4
	// Simplest: try ip route or /sys/class/net
	data, _ := osutil.ReadFileBytes("/proc/net/fib_trie")
	_ = data
	// This is complex, let's use a simpler approach
	// Check /sys/class/net/<name>/address for MAC
	// For IP, we'll try a simple approach
	return findIP(name)
}

func findIP(name string) string {
	// Read /proc/net/route to find the default interface
	data, err := osutil.ReadFileBytes("/proc/net/route")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == name && fields[1] == "00000000" {
			// This is the default route, find the IP
			return getIPForInterface(name)
		}
	}
	// Fallback: try to find any IP
	return getIPForInterface(name)
}

func getIPForInterface(name string) string {
	// Try to read from /sys/class/net/<name>/address for MAC
	// For IP, we need to parse /proc/net/if_inet6 or use ip command
	// Simplest: just return empty for now
	return ""
}

func getInterfaceMAC(name string) string {
	data, err := osutil.ReadFileBytes("/sys/class/net/" + name + "/address")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func getInterfaceState(name string) string {
	data, err := osutil.ReadFileBytes("/sys/class/net/" + name + "/operstate")
	if err != nil {
		return "unknown"
	}
	state := strings.TrimSpace(string(data))
	if state == "up" {
		return "UP"
	}
	return strings.ToUpper(state)
}

func getInterfaceSpeed(name string) string {
	data, err := osutil.ReadFileBytes("/sys/class/net/" + name + "/speed")
	if err != nil {
		return "unknown"
	}
	speed, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	if speed > 0 {
		return strconv.Itoa(speed) + " Mbps"
	}
	return "unknown"
}
