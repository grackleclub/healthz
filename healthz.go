package healthz

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Healthz struct {
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Disk    string `json:"disk"`
}

var uptime time.Time

func init() {
	uptime = time.Now()
	slog.Warn("Healthz init starting", "started", uptime)
}

// Respond is an http.HandlerFunc that returns a JSON response with
// health information and basic metrics
func Respond(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(uptime)
	cpu, err := CPU()
	var missing bool
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "cpu",
			"error", err,
		)
		missing = true
	}
	memory, err := MEM()
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "memory",
			"error", err,
		)
		missing = true
	}
	disk, err := DISK()
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "disk",
			"error", err,
		)
		missing = true
	}

	h := Healthz{
		Uptime:  uptime.String(),
		Version: "TODO",
		CPU:     fmt.Sprintf("%.2f", cpu),
		Memory:  fmt.Sprintf("%.2f", memory),
		Disk:    fmt.Sprintf("%.2f", disk),
	}

	w.Header().Set("Content-Type", "application/json")
	slog.Info("responding to healthz",
		"uptime", h.Uptime,
		"version", h.Version,
		"cpu", h.CPU,
		"memory", h.Memory,
		"disk", h.Disk,
	)
	w.Header().Set("Content-Type", "application/json")
	if missing {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(h)
}

// DISK returns the percentage of disk used by the system
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
	percentDiskUsed := (float64(used) / float64(total)) * 100
	return percentDiskUsed, nil
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
			usage := float64(user+nice+system) / float64(total) * 100
			return usage, nil
		}
	}
	return 0, fmt.Errorf("could not find CPU usage in /proc/stat")
}
