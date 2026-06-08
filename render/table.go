package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"rootinfo/query"
)

// Options controls which address families appear in the output.
type Options struct {
	ShowIPv4 bool
	ShowIPv6 bool
}

// Meta carries author/version/build-date/arch for footer and JSON metadata fields.
type Meta struct {
	Author    string
	Version   string
	BuildDate string
	Arch      string
}

func (m Meta) isZero() bool {
	return m.Author == "" && m.Version == "" && m.BuildDate == "" && m.Arch == ""
}

func (m Meta) footerString() string {
	return fmt.Sprintf("(c) %s | rootinfo v%s | Built: %s | %s", m.Author, m.Version, m.BuildDate, m.Arch)
}

// FormatTable renders results as a justified text table with | column separators
// and a separator line of hyphens and plus signs after the header.
func FormatTable(results []query.Result, opts Options, meta Meta) string {
	headers, rows := buildColumns(results, opts)

	// Compute the max width of each column across header and all data rows.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder
	writeRow(&sb, headers, widths, false)
	writeSeparator(&sb, widths)
	for _, row := range rows {
		writeRow(&sb, row, widths, true)
	}
	if !meta.isZero() {
		fmt.Fprintf(&sb, "\n%s\n", meta.footerString())
	}
	return sb.String()
}

// Table writes the formatted table to w.
func Table(w io.Writer, results []query.Result, opts Options, meta Meta) {
	fmt.Fprint(w, FormatTable(results, opts, meta))
}

// buildColumns returns the header slice and one string-slice per data row.
func buildColumns(results []query.Result, opts Options) ([]string, [][]string) {
	headers := []string{"SRV"}
	if opts.ShowIPv4 {
		headers = append(headers, "IPv4", "IPv4 Result", "IPv4 RTT")
	}
	if opts.ShowIPv6 {
		headers = append(headers, "IPv6", "IPv6 Result", "IPv6 RTT")
	}

	rows := make([][]string, len(results))
	for i, r := range results {
		row := []string{r.Server.Letter}
		if opts.ShowIPv4 {
			row = append(row, r.Server.IPv4, fmtResult(r.IPv4Result, r.IPv4Err), fmtRTT(r.IPv4RTT, r.IPv4Err))
		}
		if opts.ShowIPv6 {
			row = append(row, r.Server.IPv6, fmtResult(r.IPv6Result, r.IPv6Err), fmtRTT(r.IPv6RTT, r.IPv6Err))
		}
		rows[i] = row
	}
	return headers, rows
}

func fmtRTT(rtt time.Duration, err error) string {
	if err != nil {
		return "-"
	}
	return fmt.Sprintf("%dms", rtt.Milliseconds())
}

// writeRow emits one row, left-padding all columns except the last.
// When trim is true the last column is not padded (avoids trailing spaces on data rows).
func writeRow(sb *strings.Builder, cells []string, widths []int, trim bool) {
	for i, cell := range cells {
		if i > 0 {
			sb.WriteString(" | ")
		}
		last := i == len(cells)-1
		if last && trim {
			sb.WriteString(cell)
		} else {
			fmt.Fprintf(sb, "%-*s", widths[i], cell)
		}
	}
	sb.WriteByte('\n')
}

// writeSeparator emits a line of hyphens with '+' at each column boundary,
// aligned so that '+' sits directly under the '|' characters in content rows.
func writeSeparator(sb *strings.Builder, widths []int) {
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("-+-")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteByte('\n')
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
