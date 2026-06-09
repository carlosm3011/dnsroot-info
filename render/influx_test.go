package render

import (
	"errors"
	"strings"
	"testing"
	"time"

	"rootinfo/query"
	"rootinfo/rootservers"
)

func makeInfluxResult(letter, ipv4, ipv6, v4res, v6res string, v4rtt, v6rtt time.Duration, v4err, v6err error) query.Result {
	return query.Result{
		Server:     rootservers.Server{Letter: letter, IPv4: ipv4, IPv6: ipv6},
		IPv4Result: v4res,
		IPv4RTT:    v4rtt,
		IPv4Err:    v4err,
		IPv6Result: v6res,
		IPv6RTT:    v6rtt,
		IPv6Err:    v6err,
	}
}

func TestInflux_linesPerServer(t *testing.T) {
	results := []query.Result{
		makeInfluxResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "nnn1-lon8", "nnn1-lon8", 162*time.Millisecond, 154*time.Millisecond, nil, nil),
		makeInfluxResult("B", "170.247.170.2", "2801:1b8:10::b", "b2-scl", "b2-scl", 44*time.Millisecond, 41*time.Millisecond, nil, nil),
	}
	var sb strings.Builder
	Influx(&sb, results)
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (2 per server), got %d:\n%s", len(lines), sb.String())
	}
}

func TestInflux_success(t *testing.T) {
	results := []query.Result{
		makeInfluxResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "nnn1-lon8", "nnn1-lax", 162*time.Millisecond, 154*time.Millisecond, nil, nil),
	}
	var sb strings.Builder
	Influx(&sb, results)
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")

	v4, v6 := lines[0], lines[1]

	if !strings.HasPrefix(v4, "rootinfo_ipv4,") {
		t.Errorf("IPv4 line wrong measurement: %s", v4)
	}
	if !strings.HasPrefix(v6, "rootinfo_ipv6,") {
		t.Errorf("IPv6 line wrong measurement: %s", v6)
	}
	if !strings.Contains(v4, "server=A") {
		t.Errorf("missing server tag: %s", v4)
	}
	if !strings.Contains(v4, "address=198.41.0.4") {
		t.Errorf("missing address tag: %s", v4)
	}
	if !strings.Contains(v4, `instance="nnn1-lon8"`) {
		t.Errorf("missing instance field: %s", v4)
	}
	if !strings.Contains(v4, "rtt_ms=162.000") {
		t.Errorf("missing rtt_ms field: %s", v4)
	}
	if strings.Contains(v4, "error=") {
		t.Errorf("success line should not have error field: %s", v4)
	}
}

func TestInflux_error(t *testing.T) {
	results := []query.Result{
		makeInfluxResult("B", "170.247.170.2", "2801:1b8:10::b", "", "", 0, 0, errors.New("i/o timeout"), errors.New("connection refused")),
	}
	var sb strings.Builder
	Influx(&sb, results)
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	v4, v6 := lines[0], lines[1]

	for _, line := range []string{v4, v6} {
		if strings.Contains(line, "rtt_ms") || strings.Contains(line, "instance") {
			t.Errorf("error line should not contain rtt_ms or instance: %s", line)
		}
		if !strings.Contains(line, `error="`) {
			t.Errorf("error line missing error field: %s", line)
		}
	}
	if !strings.Contains(v4, `error="timeout"`) {
		t.Errorf("expected timeout error on IPv4 line, got: %s", v4)
	}
}

func TestInflux_timestamp(t *testing.T) {
	results := []query.Result{
		makeInfluxResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-lon", "a1-lon", 10*time.Millisecond, 10*time.Millisecond, nil, nil),
	}
	var sb strings.Builder
	Influx(&sb, results)
	line := strings.SplitN(strings.TrimRight(sb.String(), "\n"), "\n", 2)[0]

	// Line Protocol: "measurement,tags fields timestamp" — three space-separated parts.
	parts := strings.Fields(line)
	if len(parts) != 3 {
		t.Fatalf("expected 3 space-separated parts, got %d: %s", len(parts), line)
	}
	for _, c := range parts[2] {
		if c < '0' || c > '9' {
			t.Errorf("timestamp contains non-digit %q: %s", c, parts[2])
		}
	}
}

func TestInflux_tagEscaping(t *testing.T) {
	// Spaces, commas, and = in tag values must be backslash-escaped.
	if got := influxTagEscape("a b"); got != `a\ b` {
		t.Errorf("space: got %q", got)
	}
	if got := influxTagEscape("a,b"); got != `a\,b` {
		t.Errorf("comma: got %q", got)
	}
	if got := influxTagEscape("a=b"); got != `a\=b` {
		t.Errorf("equals: got %q", got)
	}
}
