# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`rootinfo` is a Go CLI tool that periodically checks the status of all DNS root servers and displays results in a formatted table, inspired by MTR (multitraceroute). It produces a self-contained binary.

## Build

```sh
make          # build binary (stamps VERSION into the binary)
make build    # explicit build
make test     # run all unit tests
```

To release a new version, update `VERSION` in the Makefile. The value is injected at build time via `-ldflags="-X rootinfo/cmd.Version=$(VERSION)"` into `cmd.Version` (`cmd/version.go`). Running with `go run .` without make shows `dev`.

## CLI Design

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --interval <sec>` | `0` | Refresh interval in seconds. `0` = run once and exit |
| `-c, --count <n>` | `0` | Stop after n refreshes (`0` = unlimited) |
| `-t, --timeout <ms>` | `2000` | Per-query DNS timeout in milliseconds |
| `-4` | — | Show IPv4 columns only |
| `-6` | — | Show IPv6 columns only |
| `--format <fmt>` | `table` | Output format: `table`, `json`, or `influx` |
| `--output <file>` | — | Write output to file (json/influx only; appends in continuous mode) |
| `--dns-server <addr>` | direct to root server IPs | Route all queries through this server |
| `-s, --servers <list>` | all 13 | Comma-separated root server letters, case-insensitive (e.g. `I,K,M`) |

### Behavior modes

- **One-shot** (`-i 0`, default): query all servers once, print output, exit.
- **Continuous** (`-i <sec>`): full-screen TUI (like MTR), redraws every interval until `Ctrl-C` or `q`.
- **Fixed count** (`-i <sec> -c <n>`): refresh n times then exit.
- **Non-table continuous to stdout**: `--format json|influx -i <sec>` streams one batch per refresh to stdout (no TUI).
- **Influx continuous to file**: `--format influx -i <sec> --output out.lp` runs TUI on screen and appends Line Protocol to the file after each refresh.

`--output` is an error with `--format table`. `--format` with an unrecognised value is an error.

### Output format

```
SRV | IPv4          | IPv4 Result    | IPv4 RTT | IPv6                | IPv6 Result    | IPv6 RTT
----+---------------+----------------+----------+---------------------+----------------+---------
A   | 198.41.0.4    | "nnn1-lon8"    | 12ms     | 2001:503:ba3e::2:30 | "nnn1-lon8"    | 9ms
B   | 170.247.170.2 | "b4-fra"       | 44ms     | 2801:1b8:10::b      | "b3-fra"       | 41ms
```

Columns are dynamically width-fitted to content. RTT shows `-` on error. Instance names are quoted as returned by the CH query.

## Architecture

### Package structure

```
rootinfo/
├── main.go                  # entry point
├── cmd/
│   ├── root.go              # cobra CLI, flag wiring, mode dispatch
│   └── version.go           # var Version/BuildDate = "dev"/"unknown" (overridden by ldflags)
├── rootservers/
│   └── rootservers.go       # hardcoded A-M data; Filter(letters) helper
├── dns/
│   └── querier.go           # Querier interface + RealQuerier (miekg/dns)
├── query/
│   └── runner.go            # Runner: parallel fan-out, Result struct
└── render/
    ├── table.go             # FormatTable(), Table(), Meta struct
    ├── json.go              # JSON() — newline-delimited output
    ├── influx.go            # Influx() — InfluxDB Line Protocol output
    └── tui.go               # bubbletea model; TUIConfig.OnRefresh for side-channel writes
```

### Key types

**`dns.Querier` interface** — the main seam for testing:
```go
QueryCHAOS(serverAddr string) (instance string, rtt time.Duration, err error)
```
`RealQuerier` sends `hostname.bind CH TXT` directly to `serverAddr:53` via miekg/dns, which already measures RTT. Test stubs implement this interface with configurable responses, RTTs, and errors.

**`query.Result`** — one row of data:
```go
type Result struct {
    Server     rootservers.Server
    IPv4Result string; IPv4RTT time.Duration; IPv4Err error
    IPv6Result string; IPv6RTT time.Duration; IPv6Err error
}
```

**`render.Options`** — controls which address families appear:
```go
type Options struct { ShowIPv4, ShowIPv6 bool }
```

### Line Protocol format

Two measurements per server per refresh — `rootinfo_ipv4` and `rootinfo_ipv6`. Tags: `server` (letter A–M), `address` (IP). Fields on success: `instance` (string), `rtt_ms` (float). Fields on error: `error` (string — `"timeout"` or `"error"`). All lines in a batch share the same nanosecond timestamp.

```
rootinfo_ipv4,server=A,address=198.41.0.4 instance="nnn1-lon8",rtt_ms=169.196 1781015379297087000
rootinfo_ipv6,server=A,address=2001:503:ba3e::2:30 instance="nnn1-lon8",rtt_ms=154.000 1781015379297087000
rootinfo_ipv4,server=B,address=170.247.170.2 error="timeout" 1781015379297087000
```

### DNS mocking in tests

Each test package defines a local `stubQuerier` struct with `responses`, `rtts`, and `errors` maps keyed by IP address. This avoids network calls in all unit tests. The interface change propagates: updating `Querier` requires updating all three stubs (in `dns/`, `query/`, `render/`).

### Stack

- **[miekg/dns](https://github.com/miekg/dns)** — DNS queries
- **[spf13/cobra](https://github.com/spf13/cobra)** — CLI parsing
- **[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)** — TUI
