package render

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"rootinfo/query"
)

// Options controls which address families appear in the output.
type Options struct {
	ShowIPv4 bool
	ShowIPv6 bool
}

// FormatTable renders results as a justified text table with | column separators.
func FormatTable(results []query.Result, opts Options) string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fmt.Fprint(tw, "SRV")
	if opts.ShowIPv4 {
		fmt.Fprint(tw, "\t| IPv4\t| IPv4 Result")
	}
	if opts.ShowIPv6 {
		fmt.Fprint(tw, "\t| IPv6\t| IPv6 Result")
	}
	fmt.Fprintln(tw)

	for _, r := range results {
		fmt.Fprint(tw, r.Server.Letter)
		if opts.ShowIPv4 {
			fmt.Fprintf(tw, "\t| %s\t| %s", r.Server.IPv4, fmtResult(r.IPv4Result, r.IPv4Err))
		}
		if opts.ShowIPv6 {
			fmt.Fprintf(tw, "\t| %s\t| %s", r.Server.IPv6, fmtResult(r.IPv6Result, r.IPv6Err))
		}
		fmt.Fprintln(tw)
	}

	tw.Flush()
	return sb.String()
}

// Table writes the formatted table to w.
func Table(w io.Writer, results []query.Result, opts Options) {
	fmt.Fprint(w, FormatTable(results, opts))
}

func fmtResult(result string, err error) string {
	if err != nil {
		return "(" + summarizeErr(err) + ")"
	}
	return `"` + result + `"`
}

func summarizeErr(err error) string {
	s := err.Error()
	if strings.Contains(s, "timeout") ||
		strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "timed out") {
		return "timeout"
	}
	return "error"
}
