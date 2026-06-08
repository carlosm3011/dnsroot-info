package cmd

import (
	"fmt"
	"os"
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
	jsonOutput bool
	dnsServer  string
	serverList string
)

var rootCmd = &cobra.Command{
	Use:   "rootinfo",
	Short: fmt.Sprintf("Check DNS root server anycast instance status\n(c) Carlos Martinez-Cagnazzo May 2026\nVersion: %s\n", Version),
	RunE:  run,
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
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "emit newline-delimited JSON (one object per refresh)")
	rootCmd.Flags().StringVar(&dnsServer, "dns-server", "", "route all queries through this server (default: direct to root server IPs)")
	rootCmd.Flags().StringVarP(&serverList, "servers", "s", "", `comma-separated root server letters to query, e.g. "I,K,M" (case-insensitive)`)
}

func run(cmd *cobra.Command, args []string) error {
	if ipv4Only && ipv6Only {
		return fmt.Errorf("-4 and -6 are mutually exclusive")
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
	}

	runner := &query.Runner{
		Querier:   &dnsq.RealQuerier{Timeout: time.Duration(timeoutMs) * time.Millisecond},
		Servers:   servers,
		DNSServer: dnsServer,
	}

	if interval == 0 {
		results := runner.Run()
		if jsonOutput {
			render.JSON(os.Stdout, results, 1, meta)
		} else {
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

	if jsonOutput {
		return runJSONContinuous(runner, cfg)
	}
	return render.RunTUI(cfg)
}

// runJSONContinuous runs queries in a loop and emits JSON to stdout.
// It does not use the TUI since JSON output must remain machine-readable.
func runJSONContinuous(runner *query.Runner, cfg render.TUIConfig) error {
	n := 0
	for {
		n++
		results := runner.Run()
		render.JSON(os.Stdout, results, n, cfg.Meta)
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
