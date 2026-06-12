# Changelog

## [0.5.0] - 2026-06-12

### Added
- TUI interactive sorting: press `1` to sort by server letter, `2` by IPv4 RTT, `3` by IPv6 RTT. Pressing the active key again toggles ascending/descending. Errors/timeouts sort last.
- Status line at the bottom of the TUI showing sort keys and current sort direction.
- macOS Intel (`darwin/amd64`) distribution target in `make dist` and GitHub Actions release workflow.

## [0.4.0] - 2026-05-xx

### Added
- InfluxDB Line Protocol output format (`--format influx`).
- `--output <file>` flag to write `json`/`influx` output to a file (appends in continuous mode).
- `--format` flag replacing the old `--json` flag; accepted values: `table`, `json`, `influx`.
- `--dns-server` flag to route all queries through a specific resolver.
- `-s / --servers` flag to query a subset of root servers.
- GitHub Actions automated release workflow.
- Windows (`windows/amd64`) and Apple Silicon (`darwin/arm64`) distribution targets.

## [0.3.0] and earlier

Initial development: one-shot and continuous TUI modes, IPv4/IPv6 columns, `-4`/`-6` flags, per-query timeout, fixed-count mode.
