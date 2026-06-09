package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"rootinfo/query"
)

// Influx writes one batch of query results as InfluxDB Line Protocol to w.
// Each server produces two lines: one for IPv4 (rootinfo_ipv4) and one for
// IPv6 (rootinfo_ipv6). On error, the line carries an error string field
// instead of instance and rtt_ms.
func Influx(w io.Writer, results []query.Result) {
	ts := time.Now().UnixNano()
	for _, r := range results {
		fmt.Fprintln(w, influxLine("rootinfo_ipv4", r.Server.Letter, r.Server.IPv4, r.IPv4Result, r.IPv4RTT, r.IPv4Err, ts))
		fmt.Fprintln(w, influxLine("rootinfo_ipv6", r.Server.Letter, r.Server.IPv6, r.IPv6Result, r.IPv6RTT, r.IPv6Err, ts))
	}
}

func influxLine(measurement, server, address, instance string, rtt time.Duration, qErr error, ts int64) string {
	tagSet := fmt.Sprintf("%s,server=%s,address=%s",
		measurement,
		influxTagEscape(server),
		influxTagEscape(address),
	)
	var fields string
	if qErr != nil {
		fields = fmt.Sprintf(`error="%s"`, summarizeErr(qErr))
	} else {
		fields = fmt.Sprintf(`instance="%s",rtt_ms=%.3f`, influxStringEscape(instance), float64(rtt.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%s %s %d", tagSet, fields, ts)
}

func influxTagEscape(s string) string {
	s = strings.ReplaceAll(s, `,`, `\,`)
	s = strings.ReplaceAll(s, `=`, `\=`)
	s = strings.ReplaceAll(s, ` `, `\ `)
	return s
}

func influxStringEscape(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
