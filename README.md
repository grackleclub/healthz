# healthz

health check tools

![workflow](https://github.com/ddbgio/healthz/actions/workflows/test.yml/badge.svg)

Basic handler and client for returning health and metrics. Responses are currently formatted as JSON. Future work may entail protobuf. Linux only. 🐧

## shared object
The `Healthz` struct can be shared between server and client, using json across the wire. The `Errors` object may contain field lookup failures from the server, or information retries added by `Ping()`.
```go
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
```

## examples

### server
First, register the `healthz.Respond` handler with your application:
```go
package main

import "github.com/ddbgio/healthz"

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", healthz.Respond)

    err = http.ListenAndServe(":8888", mux)
    if err != nil {
        panic(err)
    }
}
```

### client
Use built in methods `Ping()` and `PingWithRetry()` to check health:
```go
var (
    // override initial wait value for first retry
    healthz.InitialWait = 500 * time.Millisecond
    // set version or config revision to be returned,
    // else an empty string will be returned
    healthz.Version = "v1.2.3-abcde"
)

url := "https://foo.bar.com:8888/healthz"
maxRetries := 8

health, err := PingWithRetry(url, maxRetries)
if err != nil {
    return healthz.Healthz{}, fmt.Errorf("failed to ping healthz after retries: %w", err)
}
return health, nil
```
