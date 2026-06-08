# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`rootinfo` is a Go CLI tool that periodically checks the status of all DNS root servers and displays results in a formatted table, inspired by MTR (multitraceroute). It must produce a self-contained binary.

## Build

```sh
make          # build binary
make build    # explicit build
```

When the Makefile and Go module are initialized, the binary should be built with `go build` targeting a single static binary.

## Commands

Once implemented:

```sh
./rootinfo          # run once and display table
```

## CLI Design

### Invocation

```
rootinfo [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --interval <sec>` | `0` | Refresh interval in seconds. `0` = run once and exit |
| `-c, --count <n>` | `0` | Stop after n refreshes (`0` = unlimited) |
| `-t, --timeout <ms>` | `2000` | Per-query DNS timeout in milliseconds |
| `-4` | — | Show IPv4 column only |
| `-6` | — | Show IPv6 column only |
| `--json` | — | Emit newline-delimited JSON (one object per refresh) |
| `--dns-server <addr>` | direct to root server IPs | Route all queries through this server |
| `-s, --servers <list>` | all 13 | Comma-separated root server letters, case-insensitive (e.g. `I,K,M`) |

### Behavior modes

- **One-shot** (`-i 0`, default): query all servers once, print table, exit.
- **Continuous** (`-i <sec>`): full-screen TUI (like MTR), redraws table every interval until `Ctrl-C`.
- **Fixed count** (`-i <sec> -c <n>`): refresh n times then exit.

### Output

One-shot / JSON mode prints to stdout. Continuous mode takes over the terminal (full-screen TUI). Errors shown inline in the result column as `(timeout)` or `(error)`.

```
rootinfo v0.1  2026-06-08 15:04:05 UTC  (refresh #3, interval 5s)

SRV  IPv4              IPv4 Result       IPv6                       IPv6 Result
A    198.41.0.4        "nnn1-lon8"       2001:503:ba3e::2:30        "nnn1-frmrs-3"
B    170.247.170.2     "b4-fra"          2801:168:10::b             "b3-fra"
...
M    202.12.27.33      "m1-nrt"          2001:dc3::35               "m1-nrt"
```

## Architecture

The tool should:

1. **DNS root server list** — hardcode the 13 root server letters (A–M) with their known IPv4 and IPv6 addresses.
2. **CH TXT query** — for each server, issue a DNS `CH TXT hostname.bind` (or `id.server`) query to both the IPv4 and IPv6 addresses to retrieve the anycast instance name.
3. **Output table** — render results as a justified text table with columns: `SRV | IPv4 | IPv4 Result | IPv6 | IPv6 Result`.

### Stack

- **Language**: Go
- **DNS queries**: use a proper DNS client library (e.g., `github.com/miekg/dns`)
- **CLI parsing**: use a CLI library appropriate for flags/options
- **Build**: Makefile

### Output format

```
SRV | IPv4          | IPv4 Result    | IPv6                    | IPv6 Result
A   | 198.41.0.4    | "nnn1-lon8"    | 2001:503:ba3e::2:30     | "nnn1-frmrs-3"
B   | 170.247.170.2 | "b4-fra"       | 2801:168:10::b          | "b3-fra"
```

All columns right-padded for alignment. Instance names shown as returned by the CH query (quoted strings).
