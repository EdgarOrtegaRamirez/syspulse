// Package sysinfo provides system process information.
package sysinfo

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sort"

	"github.com/EdgarOrtegaRamirez/syspulse/internal/osutil"
)

func collectProcesses(info *Info) error {
	procs := &info.Processes

	directories, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	running, sleeping := 0, 0
	totalPIDs := osutil.NumCPUs() // At least one
	type pidInfo struct {
		pid   int
		name  string
		cpu   float64
		rss   uint64
		state string
	}
	var allProcs []pidInfo

	for _, d := range directories {
		if !d.IsDir() || !isDigit(d.Name()) {
			continue
		}
		pid, err := strconv.Atoi(d.Name())
		if err != nil {
			continue
		}

		var pi pidInfo
		pi.pid = pid

		// Read stat
		statData, err := osutil.ReadFileBytes("/proc/" + d.Name() + "/stat")
		if err == nil {
			statStr := string(statData)
			start := strings.LastIndex(statStr, "(")
			end := strings.LastIndex(statStr, ")")
			if start > 0 && end > start {
				pi.name = statStr[start+1 : end]
				rest := statStr[end+2:]
				fields := strings.Fields(rest)
				if len(fields) >= 24 {
					pi.state = string(fields[0])
					if pi.state == "R" {
						running++
					} else {
						sleeping++
					}
					utime, _ := strconv.ParseUint(fields[11], 10, 64)
					stime, _ := strconv.ParseUint(fields[12], 10, 64)
					pi.cpu = float64(utime+stime) * 100 / 100
					rss, _ := strconv.ParseUint(fields[23], 10, 64)
					pi.rss = rss * 4096
				}
			}
		}

		// Read status for memory
		statusData, err := osutil.ReadFileBytes("/proc/" + d.Name() + "/status")
		if err == nil {
			for _, line := range strings.Split(string(statusData), "\n") {
				if strings.HasPrefix(line, "VmRSS:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						val, _ := strconv.ParseUint(fields[1], 10, 64)
						pi.rss = val * 1024
					}
				}
			}
		}

		// Read cmdline for name
		if pi.name == "" {
			cmdline, err := osutil.ReadFileBytes("/proc/" + d.Name() + "/cmdline")
			if err == nil {
				cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
				cmdStr = strings.TrimSpace(cmdStr)
				if cmdStr != "" {
					pi.name = filepath.Base(strings.Split(cmdStr, " ")[0])
				}
			}
		}

		if pi.pid > 0 {
			totalPIDs++
			allProcs = append(allProcs, pi)
		}
	}

	procs.Total = totalPIDs
	procs.Running = running
	procs.Sleeping = sleeping

	// Top 10 by CPU
	sort.Slice(allProcs, func(i, j int) bool {
		return allProcs[i].cpu > allProcs[j].cpu
	})
	for i := 0; i < 10 && i < len(allProcs); i++ {
		p := allProcs[i]
		procs.TopCPU = append(procs.TopCPU, CpuProc{
			PID:   p.pid,
			Name:  p.name,
			CPU:   p.cpu,
			RSS:   p.rss,
			State: p.state,
		})
	}

	// Top 10 by memory
	sort.Slice(allProcs, func(i, j int) bool {
		return allProcs[i].rss > allProcs[j].rss
	})
	for i := 0; i < 10 && i < len(allProcs); i++ {
		p := allProcs[i]
		procs.TopMemory = append(procs.TopMemory, CpuProc{
			PID:   p.pid,
			Name:  p.name,
			CPU:   p.cpu,
			RSS:   p.rss,
			State: p.state,
		})
	}

	return nil
}

func isDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}