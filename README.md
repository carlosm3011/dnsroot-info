# rootinfo

A command-line tool for checking the status of the DNS root servers, inspired by [MTR](https://www.bitwizard.nl/mtr/). For each of the 13 root servers (A–M), `rootinfo` queries the anycast instance name via a `CHAOS TXT hostname.bind` DNS query and displays the results in a formatted table.

```
SRV | IPv4          | IPv4 Result | IPv6                | IPv6 Result
----+---------------+-------------+---------------------+------------
A   | 198.41.0.4    | "nnn1-lon8" | 2001:503:ba3e::2:30 | "nnn1-lon8"
B   | 170.247.170.2 | "b2-scl"    | 2801:1b8:10::b      | "b2-scl"
C   | 192.33.4.12   | "dca1b.c.root-servers.org" | 2001:500:2::c | "dca1b.c.root-servers.org"
...
```

This tells you which anycast node your machine is reaching for each root server — useful for network diagnostics, anycast reachability checks, and general curiosity about the DNS infrastructure you depend on.

## Building

Requires Go 1.22+.

```sh
make        # produces ./rootinfo
make test   # run unit tests
```

## Usage

**One-shot** — query all servers and print the table, then exit:

```sh
./rootinfo
```

**Continuous / MTR-like** — full-screen TUI that refreshes every 5 seconds:

```sh
./rootinfo -i 5
```

Press `q` or `Ctrl-C` to quit.

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
./rootinfo --json
./rootinfo -i 10 --json | jq .
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
| `--json` | — | Emit newline-delimited JSON |
| `--dns-server <addr>` | direct | Route queries through this server |

## How it works

Each query is a `CHAOS TXT hostname.bind` DNS request sent directly to the known IPv4 and IPv6 address of each root server. Because root servers use anycast, the response reveals which physical node answered — not just which letter. All 26 IPv4+IPv6 queries for the 13 servers run concurrently.

## Stack

- **[miekg/dns](https://github.com/miekg/dns)** — DNS client library
- **[spf13/cobra](https://github.com/spf13/cobra)** — CLI flag parsing
- **[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)** — full-screen TUI

## Notes

This project was developed with the assistance of [Claude](https://claude.ai) (Anthropic), which was used for design, code generation, and test writing.

---

(c) Carlos Martinez-Cagnazzo, May 2026
