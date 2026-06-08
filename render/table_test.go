package render

import (
	"errors"
	"strings"
	"testing"

	"rootinfo/query"
	"rootinfo/rootservers"
)

var bothOpts = Options{ShowIPv4: true, ShowIPv6: true}
var v4Only = Options{ShowIPv4: true, ShowIPv6: false}
var v6Only = Options{ShowIPv4: false, ShowIPv6: true}

func makeResult(letter, ipv4, ipv6, v4res, v6res string, v4err, v6err error) query.Result {
	return query.Result{
		Server:     rootservers.Server{Letter: letter, IPv4: ipv4, IPv6: ipv6},
		IPv4Result: v4res,
		IPv4Err:    v4err,
		IPv6Result: v6res,
		IPv6Err:    v6err,
	}
}

func TestFormatTable_header(t *testing.T) {
	out := FormatTable(nil, bothOpts)
	if !strings.Contains(out, "SRV") {
		t.Error("missing SRV header")
	}
	if !strings.Contains(out, "IPv4") {
		t.Error("missing IPv4 header")
	}
	if !strings.Contains(out, "IPv6") {
		t.Error("missing IPv6 header")
	}
}

func TestFormatTable_success(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-iad", "a1-lax", nil, nil),
	}
	out := FormatTable(results, bothOpts)
	if !strings.Contains(out, `"a1-iad"`) {
		t.Errorf("missing IPv4 result, got:\n%s", out)
	}
	if !strings.Contains(out, `"a1-lax"`) {
		t.Errorf("missing IPv6 result, got:\n%s", out)
	}
}

func TestFormatTable_timeout(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "", "", errors.New("i/o timeout"), errors.New("deadline exceeded")),
	}
	out := FormatTable(results, bothOpts)
	if strings.Count(out, "(timeout)") != 2 {
		t.Errorf("expected 2 timeout markers, got:\n%s", out)
	}
}

func TestFormatTable_genericError(t *testing.T) {
	results := []query.Result{
		makeResult("B", "170.247.170.2", "2801:1b8:10::b", "", "", errors.New("rcode REFUSED"), nil),
	}
	out := FormatTable(results, bothOpts)
	if !strings.Contains(out, "(error)") {
		t.Errorf("expected (error), got:\n%s", out)
	}
}

func TestFormatTable_v4Only(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-iad", "a1-lax", nil, nil),
	}
	out := FormatTable(results, v4Only)
	if strings.Contains(out, "IPv6") {
		t.Errorf("IPv6 column should not appear in v4-only mode, got:\n%s", out)
	}
	if !strings.Contains(out, "198.41.0.4") {
		t.Errorf("IPv4 address missing, got:\n%s", out)
	}
}

func TestFormatTable_v6Only(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-iad", "a1-lax", nil, nil),
	}
	out := FormatTable(results, v6Only)
	if strings.Contains(out, "IPv4") {
		t.Errorf("IPv4 column should not appear in v6-only mode, got:\n%s", out)
	}
	if !strings.Contains(out, "2001:503:ba3e::2:30") {
		t.Errorf("IPv6 address missing, got:\n%s", out)
	}
}

func TestFormatTable_separators(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-iad", "a1-lax", nil, nil),
	}
	out := FormatTable(results, bothOpts)
	if !strings.Contains(out, "|") {
		t.Errorf("expected | column separators, got:\n%s", out)
	}
}

func TestFormatTable_separatorLine(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-iad", "a1-lax", nil, nil),
	}
	out := FormatTable(results, bothOpts)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header, separator, data), got %d", len(lines))
	}
	sep := lines[1]
	if !strings.Contains(sep, "-") {
		t.Errorf("separator line missing hyphens: %q", sep)
	}
	if !strings.Contains(sep, "+") {
		t.Errorf("separator line missing plus signs: %q", sep)
	}
	// Separator must not contain letters or pipe characters.
	for _, c := range sep {
		if c == '|' {
			t.Errorf("separator line should use + not |: %q", sep)
			break
		}
	}
}

func TestFormatTable_separatorAlignsWithPipes(t *testing.T) {
	results := []query.Result{
		makeResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "a1-iad", "a1-lax", nil, nil),
	}
	out := FormatTable(results, bothOpts)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	header := lines[0]
	sep := lines[1]

	// Every '|' in the header must align with a '+' in the separator.
	for i, c := range header {
		if c == '|' {
			if i >= len(sep) || sep[i] != '+' {
				t.Errorf("position %d: header has '|' but separator has %q\nheader: %s\nsep:    %s",
					i, sep[i], header, sep)
			}
		}
	}
}
