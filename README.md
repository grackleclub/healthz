# healthz

health check tools

![workflow](https://github.com/ddbgio/healthz/actions/workflows/test.yml/badge.svg)

Basic handler and client for returning health and metrics. Responses are currently formatted as JSON. Future work may entail protobuf.

> [!WARNING]
> Only Linux machines are supported. 🐧

## shared object
The `Healthz` struct can be shared between server and client, using json across the wire. An http `404 Not Found` error will be returned if CPU, Memory, or Disk usage retrieval fails.
```go
// Healthz is a shared object between client and server
// to check http status and basic metrics
type Healthz struct {
	Time    int    `json:"time"`    // unix timestamp
	Status  int    `json:"status"`  // http status code
	Uptime  string `json:"uptime"`  // time since last restart
	Version string `json:"version"` // set by the server
	CPU     string `json:"cpu"`     // percent (between 0 and 1)
	Memory  string `json:"memory"`  // percent
	Disk    string `json:"disk"`    // percent
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

url := "https://foo.bar.com/healthz"
maxRetries := 8

health, err := PingWithRetry(url, maxRetries)
if err != nil {
    return healthz.Healthz{}, fmt.Errorf("failed to ping healthz after retries: %w", err)
}
return health, nil
```
