# rootinfo

A command-line tool for checking the status of the DNS root servers, inspired by [MTR](https://www.bitwizard.nl/mtr/). For each of the 13 root servers (A–M), `rootinfo` queries the anycast instance name via `CHAOS TXT` DNS queries (`hostname.bind`, with `id.server` as a fallback) and displays the results in a formatted table, including the round-trip time for each query.

```
SRV | IPv4          | IPv4 Result                | IPv4 RTT | IPv6                | IPv6 Result                | IPv6 RTT
----+---------------+----------------------------+----------+---------------------+----------------------------+---------
A   | 198.41.0.4    | "nnn1-lon8"                | 167ms    | 2001:503:ba3e::2:30 | "nnn1-lon8"                | 154ms
B   | 170.247.170.2 | "b2-scl"                   | 44ms     | 2801:1b8:10::b      | "b2-scl"                   | 41ms
C   | 192.33.4.12   | "dca1b.c.root-servers.org" | 146ms    | 2001:500:2::c       | "dca1b.c.root-servers.org" | 143ms
...
```

This tells you which anycast node your machine is reaching for each root server, and how long it took — useful for network diagnostics, anycast reachability checks, and general curiosity about the DNS infrastructure you depend on.

## Building

Requires Go 1.22+.

```sh
make        # produces ./rootinfo for the current platform
make test   # run unit tests
make dist   # cross-compile for all supported platforms (see below)
```

The version number is set via `VERSION` in the Makefile and stamped into the binary at build time. Running with `go run .` directly shows `dev`.

### Distribution builds

`make dist` produces four binaries in `dist/`:

| File | Platform |
|------|----------|
| `rootinfo-darwin-arm64` | macOS — Apple Silicon (M-series) |
| `rootinfo-darwin-amd64` | macOS — Intel |
| `rootinfo-linux-amd64` | Linux — x86-64 |
| `rootinfo-windows-amd64.exe` | Windows — x86-64 (Terminal / PowerShell) |

Each binary is stamped with the version, build date, and its target architecture, which appear in the table footer.

## Usage

**One-shot** — query all servers and print the table, then exit:

```sh
./rootinfo
```

**Continuous / MTR-like** — full-screen TUI that refreshes every 5 seconds:

```sh
./rootinfo -i 5
```

Press `q` or `Ctrl-C` to quit. In TUI mode you can sort the table interactively:

| Key | Sort |
|-----|------|
| `1` | Server letter (default) |
| `2` | IPv4 RTT |
| `3` | IPv6 RTT |

Pressing the same key again toggles ascending/descending order.

**Fixed count** — refresh 10 times then exit:

```sh
./rootinfo -i 5 -c 10
```

**Filter servers** — query only a subset (case-insensitive):

```sh
./rootinfo -s I,K,M
```

**IPv4 or IPv6 only:**

```sh
./rootinfo -4   # IPv4 only
./rootinfo -6   # IPv6 only
```

**JSON output** — newline-delimited JSON, one object per refresh (works in both one-shot and continuous mode):

```sh
./rootinfo --format json
./rootinfo -i 10 --format json | jq .
```

Each JSON object includes `ipv4_rtt_ms` and `ipv6_rtt_ms` fields (omitted on error).

**InfluxDB Line Protocol output** — one batch of measurements per refresh:

```sh
./rootinfo --format influx
./rootinfo -i 10 --format influx --output metrics.lp
```

**Custom DNS server** — route all queries through a specific server instead of querying root server IPs directly:

```sh
./rootinfo --dns-server 9.9.9.9
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --interval <sec>` | `0` | Refresh interval in seconds (`0` = one-shot) |
| `-c, --count <n>` | `0` | Stop after N refreshes (`0` = unlimited) |
| `-t, --timeout <ms>` | `2000` | Per-query DNS timeout in milliseconds |
| `-4` | — | Show IPv4 results only |
| `-6` | — | Show IPv6 results only |
| `-s, --servers <list>` | all 13 | Comma-separated server letters, e.g. `I,K,M` |
| `--format <fmt>` | `table` | Output format: `table`, `json`, or `influx` |
| `--output <file>` | — | Write output to file (`json`/`influx` only; appends in continuous mode) |
| `--dns-server <addr>` | direct | Route queries through this server |

## How it works

For each server, `rootinfo` first tries a `CHAOS TXT hostname.bind` query:

```
dig CHAOS TXT hostname.bind @<root-server-ip>
```

If that query fails (timeout, REFUSED, or no TXT record in the response), it falls back to `id.server`:

```
dig CHAOS TXT id.server @<root-server-ip>
```

Some root servers (G is a known example) do not respond to `hostname.bind` but do answer `id.server`. The RTT reported is from whichever query succeeded.

Because root servers use anycast, the instance name in the response reveals which physical node answered — not just which letter. All 26 IPv4+IPv6 queries for the 13 servers run concurrently.

## Stack

- **[miekg/dns](https://github.com/miekg/dns)** — DNS client library
- **[spf13/cobra](https://github.com/spf13/cobra)** — CLI flag parsing
- **[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)** — full-screen TUI

## Notes

This project was developed with the assistance of [Claude](https://claude.ai) (Anthropic), which was used for design, code generation, and test writing.

---

(c) Carlos Martinez-Cagnazzo, May 2026
