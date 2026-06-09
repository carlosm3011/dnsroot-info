package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	dnsq "rootinfo/dns"
	"rootinfo/query"
	"rootinfo/render"
	"rootinfo/rootservers"
)

var (
	interval   int
	count      int
	timeoutMs  int
	ipv4Only   bool
	ipv6Only   bool
	format     string
	outputFile string
	dnsServer  string
	serverList string
)

var rootCmd = &cobra.Command{
	Use:   "rootinfo",
	Short: fmt.Sprintf("Check DNS root server anycast instance status (v%s)", Version),
	Long: fmt.Sprintf(`rootinfo v%s — DNS root server anycast status checker
(c) Carlos Martinez-Cagnazzo, May 2026

For each of the 13 root servers (A–M), queries the anycast instance name via
CHAOS TXT DNS (hostname.bind, falling back to id.server) and reports the
responding node and round-trip time for both IPv4 and IPv6.

Modes:
  One-shot (default)   Query all servers once, print table, exit.
  Continuous (-i sec)  Full-screen TUI that refreshes every N seconds.
                       Press q or Ctrl-C to quit.
  Fixed count          Add -c N to stop after N refreshes.

Output formats (--format):
  table   Human-readable aligned table with footer (default)
  json    Newline-delimited JSON, one object per refresh
  influx  InfluxDB Line Protocol (rootinfo_ipv4 / rootinfo_ipv6 measurements)

The --output flag writes json or influx output to a file (appended). In
continuous influx-to-file mode the TUI remains visible on screen while data
accumulates in the file — useful for feeding Telegraf or bulk import.`, Version),
	Example: `  rootinfo                          # one-shot table, all servers
  rootinfo -i 5                     # continuous TUI, refresh every 5s
  rootinfo -i 5 -c 10               # refresh 10 times then exit
  rootinfo -s I,K,M                 # query only I, K, and M root servers
  rootinfo -4                       # IPv4 results only
  rootinfo --format json            # one-shot JSON to stdout
  rootinfo --format json -i 10      # streaming JSON, one object per refresh
  rootinfo --format influx -s A,B   # Line Protocol to stdout
  rootinfo --format influx -i 30 --output /tmp/rootinfo.lp
                                    # TUI on screen, influx appended to file
  rootinfo --dns-server 9.9.9.9     # route queries through a specific resolver`,
	RunE: run,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 0, "refresh interval in seconds (0 = one-shot)")
	rootCmd.Flags().IntVarP(&count, "count", "c", 0, "stop after N refreshes (0 = unlimited)")
	rootCmd.Flags().IntVarP(&timeoutMs, "timeout", "t", 2000, "per-query DNS timeout in milliseconds")
	rootCmd.Flags().BoolVarP(&ipv4Only, "ipv4", "4", false, "show IPv4 results only")
	rootCmd.Flags().BoolVarP(&ipv6Only, "ipv6", "6", false, "show IPv6 results only")
	rootCmd.Flags().StringVar(&format, "format", "table", `output format: table, json, or influx`)
	rootCmd.Flags().StringVar(&outputFile, "output", "", "write output to this file (json and influx only; appends in continuous mode)")
	rootCmd.Flags().StringVar(&dnsServer, "dns-server", "", "route all queries through this server (default: direct to root server IPs)")
	rootCmd.Flags().StringVarP(&serverList, "servers", "s", "", `comma-separated root server letters to query, e.g. "I,K,M" (case-insensitive)`)
}

func run(cmd *cobra.Command, args []string) error {
	if ipv4Only && ipv6Only {
		return fmt.Errorf("-4 and -6 are mutually exclusive")
	}
	if format != "table" && format != "json" && format != "influx" {
		return fmt.Errorf("unknown format %q: must be table, json, or influx", format)
	}
	if outputFile != "" && format == "table" {
		return fmt.Errorf("--output requires --format json or --format influx")
	}

	servers := rootservers.Filter(parseLetters(serverList))
	if len(servers) == 0 {
		return fmt.Errorf("no matching root servers for: %s", serverList)
	}

	opts := render.Options{
		ShowIPv4: !ipv6Only,
		ShowIPv6: !ipv4Only,
	}
	meta := render.Meta{
		Author:    "Carlos Martinez-Cagnazzo",
		Version:   Version,
		BuildDate: BuildDate,
		Arch:      runtime.GOARCH,
	}

	runner := &query.Runner{
		Querier:   &dnsq.RealQuerier{Timeout: time.Duration(timeoutMs) * time.Millisecond},
		Servers:   servers,
		DNSServer: dnsServer,
	}

	w, closeW, err := openWriter(outputFile)
	if err != nil {
		return err
	}
	defer closeW()

	if interval == 0 {
		results := runner.Run()
		switch format {
		case "json":
			render.JSON(w, results, 1, meta)
		case "influx":
			render.Influx(w, results)
		default:
			render.Table(os.Stdout, results, opts, meta)
		}
		return nil
	}

	cfg := render.TUIConfig{
		Runner:   runner,
		Interval: time.Duration(interval) * time.Second,
		MaxCount: count,
		Opts:     opts,
		Meta:     meta,
	}

	switch format {
	case "json":
		return runStreamContinuous(runner, cfg, func(results []query.Result, n int) {
			render.JSON(w, results, n, meta)
		})
	case "influx":
		if outputFile != "" {
			// TUI on screen; influx appended to file after each refresh.
			cfg.OnRefresh = func(results []query.Result, n int) {
				render.Influx(w, results)
			}
			return render.RunTUI(cfg)
		}
		return runStreamContinuous(runner, cfg, func(results []query.Result, n int) {
			render.Influx(w, results)
		})
	default:
		return render.RunTUI(cfg)
	}
}

// openWriter returns a writer for the given path (append mode), or stdout if path is empty.
func openWriter(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, func() {}, fmt.Errorf("opening output file: %w", err)
	}
	return f, func() { f.Close() }, nil
}

// runStreamContinuous runs queries in a loop, calling emit after each batch.
// It does not use the TUI — output must remain machine-readable on stdout.
func runStreamContinuous(runner *query.Runner, cfg render.TUIConfig, emit func([]query.Result, int)) error {
	n := 0
	for {
		n++
		results := runner.Run()
		emit(results, n)
		if cfg.MaxCount > 0 && n >= cfg.MaxCount {
			return nil
		}
		time.Sleep(cfg.Interval)
	}
}

// parseLetters splits a comma-separated string of server letters into a slice.
func parseLetters(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
