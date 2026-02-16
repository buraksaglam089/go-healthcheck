# go-healthcheck

A small Go application that checks a list of URLs at fixed intervals, prints results to the terminal, and appends logs as line-by-line JSON.

You can think of it as a minimal uptime and microservice checker.

## What it does

- Monitors multiple target URLs concurrently
- Checks each target on its own `interval`
- Produces HTTP status code, latency, timestamp, and error info
- Writes output both to the terminal (colored) and to a JSON log file
- Shuts down gracefully on `Ctrl+C` (`SIGTERM` / `SIGINT`)

## Project structure

- `main.go`: App entrypoint, config loading, signal handling
- `monitor/checker.go`: HTTP check logic (`GET` request + timeout)
- `monitor/engine.go`: Worker goroutines and result flow
- `storage/logger.go`: File logging + terminal output
- `targets.json`: Example target configuration
- `storage/target_json.json`: Generated logs (JSON Lines)

## Requirements

- Go 1.22+

## Setup

```bash
git clone https://github.com/buraksaglam089/go-healthcheck.git
cd go-healthcheck
go mod tidy
```

## Run

By default, it reads `targets.json`:

```bash
go run .
```

To use a different config file:

```bash
go run . --config my-targets.json
```

## Config format

Each target uses the following fields:

- `id`: Target label
- `url`: URL to check
- `interval`: Check frequency in seconds
- `timeout`: Request timeout in seconds

Example:

```json
[
  {
    "id": "google",
    "url": "https://google.com",
    "interval": 5,
    "timeout": 3
  },
  {
    "id": "github",
    "url": "https://github.com",
    "interval": 10,
    "timeout": 5
  }
]
```

## Log output

Results are appended to `storage/target_json.json`, one JSON object per line (JSON Lines format).

Example line:

```json
{"TargetID":"google","TargetURL":"https://google.com","StatusCode":200,"Latency":123456789,"Timestamp":"2026-02-16T18:40:00.000000+03:00","Err":null}
```
