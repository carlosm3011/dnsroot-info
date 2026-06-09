package render

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"rootinfo/query"
	"rootinfo/rootservers"
)

var testMeta = Meta{Author: "Test", Version: "1.0", BuildDate: "2026-01-01", Arch: "amd64"}

func makeJSONResult(letter, ipv4, ipv6, v4res, v6res string, v4rtt, v6rtt time.Duration, v4err, v6err error) query.Result {
	return query.Result{
		Server:     rootservers.Server{Letter: letter, IPv4: ipv4, IPv6: ipv6},
		IPv4Result: v4res, IPv4RTT: v4rtt, IPv4Err: v4err,
		IPv6Result: v6res, IPv6RTT: v6rtt, IPv6Err: v6err,
	}
}

func parseJSON(t *testing.T, w *strings.Builder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(w.String(), "\n")), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, w.String())
	}
	return out
}

func TestJSON_validJSON(t *testing.T) {
	var sb strings.Builder
	JSON(&sb, nil, 1, testMeta)
	if !json.Valid([]byte(strings.TrimRight(sb.String(), "\n"))) {
		t.Errorf("output is not valid JSON:\n%s", sb.String())
	}
}

func TestJSON_newlineDelimited(t *testing.T) {
	var sb strings.Builder
	JSON(&sb, nil, 1, testMeta)
	if !strings.HasSuffix(sb.String(), "\n") {
		t.Errorf("output must end with newline, got: %q", sb.String())
	}
	// Must be exactly one line (one JSON object).
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestJSON_refreshField(t *testing.T) {
	var sb strings.Builder
	JSON(&sb, nil, 7, testMeta)
	out := parseJSON(t, &sb)
	if got := out["refresh"].(float64); got != 7 {
		t.Errorf("refresh: want 7, got %v", got)
	}
}

func TestJSON_timestampPresent(t *testing.T) {
	var sb strings.Builder
	JSON(&sb, nil, 1, testMeta)
	out := parseJSON(t, &sb)
	ts, ok := out["timestamp"].(string)
	if !ok || ts == "" {
		t.Errorf("timestamp missing or not a string: %v", out["timestamp"])
	}
	if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
		t.Errorf("timestamp not RFC3339: %v", err)
	}
}

func TestJSON_metaFields(t *testing.T) {
	var sb strings.Builder
	JSON(&sb, nil, 1, testMeta)
	out := parseJSON(t, &sb)
	meta, ok := out["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta field missing or wrong type")
	}
	if meta["author"] != "Test" {
		t.Errorf("meta.author: got %v", meta["author"])
	}
	if meta["version"] != "1.0" {
		t.Errorf("meta.version: got %v", meta["version"])
	}
	if meta["build_date"] != "2026-01-01" {
		t.Errorf("meta.build_date: got %v", meta["build_date"])
	}
	if meta["arch"] != "amd64" {
		t.Errorf("meta.arch: got %v", meta["arch"])
	}
}

func TestJSON_success(t *testing.T) {
	results := []query.Result{
		makeJSONResult("A", "198.41.0.4", "2001:503:ba3e::2:30", "nnn1-lon8", "nnn1-lax", 162*time.Millisecond, 154*time.Millisecond, nil, nil),
	}
	var sb strings.Builder
	JSON(&sb, results, 1, testMeta)
	out := parseJSON(t, &sb)

	servers := out["servers"].([]any)
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	s := servers[0].(map[string]any)
	if s["letter"] != "A" {
		t.Errorf("letter: got %v", s["letter"])
	}
	if s["ipv4"] != "198.41.0.4" {
		t.Errorf("ipv4: got %v", s["ipv4"])
	}
	if s["ipv4_result"] != "nnn1-lon8" {
		t.Errorf("ipv4_result: got %v", s["ipv4_result"])
	}
	if s["ipv4_rtt_ms"].(float64) != 162 {
		t.Errorf("ipv4_rtt_ms: got %v", s["ipv4_rtt_ms"])
	}
	if _, hasErr := s["ipv4_error"]; hasErr {
		t.Errorf("ipv4_error should be absent on success")
	}
}

func TestJSON_error(t *testing.T) {
	results := []query.Result{
		makeJSONResult("B", "170.247.170.2", "2801:1b8:10::b", "", "", 0, 0, errors.New("i/o timeout"), errors.New("connection refused")),
	}
	var sb strings.Builder
	JSON(&sb, results, 1, testMeta)
	out := parseJSON(t, &sb)

	s := out["servers"].([]any)[0].(map[string]any)
	if _, ok := s["ipv4_result"]; ok {
		t.Errorf("ipv4_result should be absent on error")
	}
	if _, ok := s["ipv4_rtt_ms"]; ok {
		t.Errorf("ipv4_rtt_ms should be absent on error")
	}
	if s["ipv4_error"] == nil {
		t.Errorf("ipv4_error should be present on error")
	}
	if s["ipv6_error"] == nil {
		t.Errorf("ipv6_error should be present on error")
	}
}

func TestJSON_multipleRefreshes(t *testing.T) {
	var sb strings.Builder
	JSON(&sb, nil, 1, testMeta)
	JSON(&sb, nil, 2, testMeta)
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines for 2 refreshes, got %d", len(lines))
	}
	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Errorf("line %d is not valid JSON: %s", i+1, line)
		}
	}
}
