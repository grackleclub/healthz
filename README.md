# healthz

health check tools

![workflow](https://github.com/ddbgio/healthz/actions/workflows/test.yml/badge.svg)

Basic handler and client for returning health and metrics. Responses are currently formatted as JSON. Future work may entail protobuf.

> [!WARNING]
> Only Linux machines are supported. 🐧

## usage
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

Then, check the status of the application with a simple check:
```go
var (
    // override initial wait value for first retry
	healthz.InitialWait = 500 * time.Millisecond
)

url := "https://foo.bar.com/healthz"
maxRetries := 8

health, err := PingWithRetry(url, maxRetries)
if err != nil {
    return healthz.Healthz{}, fmt.Errorf("failed to ping healthz after retries: %w", err)
}
return health, nil
```
