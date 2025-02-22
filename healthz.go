package healthz

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// Healthz is a shared object between client and server
// to check http status and basic metrics
type Healthz struct {
	Time    int     `json:"time"`    // unix timestamp
	Status  int     `json:"status"`  // http status code
	Version string  `json:"version"` // version of the service
	Uptime  string  `json:"uptime"`  // minutes since last init
	CPU     string  `json:"cpu"`     // percent (between 0 and 1)
	Memory  string  `json:"memory"`  // percent
	Disk    string  `json:"disk"`    // percent
	Load1   string  `json:"load1"`   // 1 minute load average
	Load5   string  `json:"load5"`   // 5 minute load average
	Load15  string  `json:"load15"`  // 15 minute load average
	Errors  []error `json:"errors"`  // list of errors encountered
}

var (
	initTime  time.Time                // time since the server started
	Version   string                   // user-defined version which can be set by service
	retryWait = 500 * time.Millisecond // length of first wait (2x for all subsequent)
)

func init() {
	initTime = time.Now()
	if testing.Testing() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	slog.Debug("Healthz init starting",
		"initTime", initTime,
		"retryWait", retryWait,
	)
}

// Respond is an http.HandlerFunc that returns a JSON response unmarshalable
// into a Healthz object. A failure of any field will return a partial response
// and a 404 status code for the request. Service status is reflected in Healthz.Status
func Respond(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(initTime)
	cpu, err := CPU()
	var errors []error
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "cpu",
			"error", err,
		)
		errors = append(errors, fmt.Errorf("cpu fetch failed"))
	}
	memory, err := MEM()
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "memory",
			"error", err,
		)
		errors = append(errors, fmt.Errorf("memory fetch failed"))
	}
	disk, err := DISK()
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "disk",
			"error", err,
		)
		errors = append(errors, fmt.Errorf("disk fetch failed"))
	}
	load1, load5, load15, err := Load()
	if err != nil {
		slog.Error("healthz metrics check failed",
			"target", "load",
			"error", err,
		)
		errors = append(errors, fmt.Errorf("load fetch failed"))
	}

	h := Healthz{
		Time:    int(time.Now().Unix()),
		Uptime:  fmt.Sprintf("%.2f", uptime.Minutes()),
		Version: Version,
		CPU:     fmt.Sprintf("%.2f", cpu),
		Memory:  fmt.Sprintf("%.2f", memory),
		Disk:    fmt.Sprintf("%.2f", disk),
		Load1:   fmt.Sprintf("%.2f", load1),
		Load5:   fmt.Sprintf("%.2f", load5),
		Load15:  fmt.Sprintf("%.2f", load15),
		Errors:  errors,
	}

	w.Header().Set("Content-Type", "application/json")
	slog.Debug("responding to healthz",
		"time", h.Time,
		"uptime", h.Uptime,
		"cpu", h.CPU,
		"memory", h.Memory,
		"disk", h.Disk,
		"load1", h.Load1,
		"load5", h.Load5,
		"load15", h.Load15,
		"errors", h.Errors,
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h)
}

// Ping sends a GET request to the provided healthz URL,
// returning a healthz object
func Ping(url string) (Healthz, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Healthz{}, fmt.Errorf("unable to create new healthz request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Healthz{}, fmt.Errorf("unable to perform healthz request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Healthz{}, fmt.Errorf("unable to read healthz response: %w", err)
	}

	h := Healthz{}
	err = json.Unmarshal(body, &h)
	if err != nil {
		return Healthz{}, fmt.Errorf("unable to unmarshal healthz response: %w", err)
	}
	slog.Warn("healthz ping response", "status", resp.StatusCode)
	h.Status = resp.StatusCode
	return h, nil
}

// PingWithRetry sends a GET request to the provided healthz URL,
// retrying up to maxRetries times with exponential backoff
func PingWithRetry(url string, maxRetries int) (Healthz, error) {
	wait := retryWait
	for i := 0; i < maxRetries; i++ {
		wait *= 2
		h, err := Ping(url)
		if err == nil && h.Status == http.StatusOK {
			return h, nil
		}
		if err != nil {
			slog.Info("healthz ping failed",
				"url", url,
				"attempt", i+1,
				"attemptMax", maxRetries,
				"error", err,
			)
		}
		if h.Status != http.StatusOK {
			slog.Info("healthz ping returned non-200 status",
				"url", url,
				"attempt", i+1,
				"attemptMax", maxRetries,
				"status", h.Status,
			)
		}
		time.Sleep(wait)
	}
	return Healthz{}, fmt.Errorf(
		"unable to ping healthz after %d retries", maxRetries,
	)
}
