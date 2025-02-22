package healthz

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// CPU returns the percentage of CPU used by the system
func CPU() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, fmt.Errorf("unable to read /proc/stat: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return 0, fmt.Errorf("invalid format in /proc/stat, expected >=8, got %d", len(fields))
			}

			user, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse failure reading user field: %w", err)
			}
			nice, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse failure reading nice field: %w", err)
			}
			system, err := strconv.ParseUint(fields[3], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse failure reading system field: %w", err)
			}
			idle, err := strconv.ParseUint(fields[4], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse failure reading idle field: %w", err)
			}

			total := user + nice + system + idle
			usage := float64(user+nice+system) / float64(total)
			return usage, nil
		}
	}
	return 0, fmt.Errorf("could not find CPU usage in /proc/stat")
}

// Load returns the average system load over 1, 5, and 15 minutes
func Load() (float64, float64, float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("unable to read /proc/loadavg: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid format in /proc/loadavg, expected >=3, got %d", len(fields))
	}

	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse failure reading 1-minute load field: %w", err)
	}
	load5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse failure reading 5-minute load field: %w", err)
	}
	load15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse failure reading 15-minute load field: %w", err)
	}
	return load1, load5, load15, nil
}

// MEM returns the percentage of memory used by the system
func MEM() (float64, error) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, fmt.Errorf("unable to read /proc/self/status: %w", err)
	}

	var totalMemory, rssMemory uint64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0, fmt.Errorf("invalid format in /proc/self/status, expected >=2, got %d", len(fields))
			}
			rssMemory, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse failure reading VmRSS field: %w", err)
			}
		}
		if strings.HasPrefix(line, "VmSize:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0, fmt.Errorf("invalid format in /proc/self/status, expected >=2, got %d", len(fields))
			}
			totalMemory, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse failure reading VmSize field: %w", err)
			}
		}
	}

	if totalMemory == 0 {
		return 0, fmt.Errorf("total memory is zero, invalid data")
	}

	// Calculate the percentage of memory used
	percentMemoryUsed := (float64(rssMemory) / float64(totalMemory)) * 100
	return percentMemoryUsed, nil
}

// DISK returns the percentage of disk space in use
func DISK() (float64, error) {
	var stat syscall.Statfs_t

	wd, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("unable to get current working directory: %w", err)
	}

	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return 0, fmt.Errorf("unable to get file system statistics: %w", err)
	}

	// Total blocks * size per block = total size
	total := stat.Blocks * uint64(stat.Bsize)
	// Free blocks * size per block = free size
	free := stat.Bfree * uint64(stat.Bsize)
	// Used size = total size - free size
	used := total - free

	if total == 0 {
		return 0, fmt.Errorf("total disk space is zero, invalid data")
	}

	// Calculate the percentage of disk used
	percentDiskUsed := (float64(used) / float64(total))
	return percentDiskUsed, nil
}
