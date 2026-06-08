package query

import (
	"errors"
	"testing"
	"time"

	"rootinfo/rootservers"
)

// stubQuerier implements dns.Querier for testing.
type stubQuerier struct {
	responses map[string]string
	rtts      map[string]time.Duration
	errors    map[string]error
}

func (s *stubQuerier) QueryCHAOS(addr string) (string, time.Duration, error) {
	if err, ok := s.errors[addr]; ok {
		return "", 0, err
	}
	if r, ok := s.responses[addr]; ok {
		return r, s.rtts[addr], nil
	}
	return "", 0, errors.New("unconfigured: " + addr)
}

var testServers = []rootservers.Server{
	{Letter: "A", IPv4: "198.41.0.4", IPv6: "2001:503:ba3e::2:30"},
	{Letter: "B", IPv4: "170.247.170.2", IPv6: "2801:1b8:10::b"},
}

func TestRunner_allSuccess(t *testing.T) {
	q := &stubQuerier{
		responses: map[string]string{
			"198.41.0.4":          "a1-iad",
			"2001:503:ba3e::2:30": "a1-lax",
			"170.247.170.2":       "b4-fra",
			"2801:1b8:10::b":      "b3-fra",
		},
		rtts: map[string]time.Duration{
			"198.41.0.4":          10 * time.Millisecond,
			"2001:503:ba3e::2:30": 15 * time.Millisecond,
			"170.247.170.2":       20 * time.Millisecond,
			"2801:1b8:10::b":      25 * time.Millisecond,
		},
	}
	r := &Runner{Querier: q, Servers: testServers}
	results := r.Run()

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Server.Letter != "A" {
		t.Errorf("expected A first, got %s", results[0].Server.Letter)
	}
	if results[0].IPv4Result != "a1-iad" {
		t.Errorf("A IPv4: want a1-iad, got %s", results[0].IPv4Result)
	}
	if results[0].IPv4RTT != 10*time.Millisecond {
		t.Errorf("A IPv4 RTT: want 10ms, got %v", results[0].IPv4RTT)
	}
	if results[0].IPv6Result != "a1-lax" {
		t.Errorf("A IPv6: want a1-lax, got %s", results[0].IPv6Result)
	}
	if results[0].IPv6RTT != 15*time.Millisecond {
		t.Errorf("A IPv6 RTT: want 15ms, got %v", results[0].IPv6RTT)
	}
	if results[1].IPv4Result != "b4-fra" {
		t.Errorf("B IPv4: want b4-fra, got %s", results[1].IPv4Result)
	}
}

func TestRunner_rttZeroOnError(t *testing.T) {
	q := &stubQuerier{
		errors: map[string]error{
			"198.41.0.4":          errors.New("i/o timeout"),
			"2001:503:ba3e::2:30": errors.New("i/o timeout"),
		},
	}
	r := &Runner{Querier: q, Servers: testServers[:1]}
	results := r.Run()

	if results[0].IPv4RTT != 0 {
		t.Errorf("expected zero RTT on error, got %v", results[0].IPv4RTT)
	}
	if results[0].IPv6RTT != 0 {
		t.Errorf("expected zero RTT on error, got %v", results[0].IPv6RTT)
	}
}

func TestRunner_partialFailure(t *testing.T) {
	timeout := errors.New("i/o timeout")
	q := &stubQuerier{
		responses: map[string]string{
			"198.41.0.4": "a1-iad",
		},
		errors: map[string]error{
			"2001:503:ba3e::2:30": timeout,
		},
	}
	r := &Runner{Querier: q, Servers: testServers[:1]}
	results := r.Run()

	if results[0].IPv4Err != nil {
		t.Errorf("expected IPv4 success, got %v", results[0].IPv4Err)
	}
	if results[0].IPv6Err == nil {
		t.Error("expected IPv6 error, got nil")
	}
}

func TestRunner_preservesOrder(t *testing.T) {
	// Use many servers to expose ordering issues from concurrent execution.
	many := rootservers.All
	responses := make(map[string]string)
	for _, s := range many {
		responses[s.IPv4] = s.Letter + "-v4"
		responses[s.IPv6] = s.Letter + "-v6"
	}
	q := &stubQuerier{responses: responses}
	r := &Runner{Querier: q, Servers: many}
	results := r.Run()

	for i, res := range results {
		if res.Server.Letter != many[i].Letter {
			t.Errorf("position %d: want %s, got %s", i, many[i].Letter, res.Server.Letter)
		}
	}
}

func TestRunner_dnsServerOverride(t *testing.T) {
	// When DNSServer is set, all queries go to it, not to the root server IPs.
	q := &stubQuerier{
		responses: map[string]string{
			"9.9.9.9": "override-instance",
		},
	}
	r := &Runner{Querier: q, Servers: testServers[:1], DNSServer: "9.9.9.9"}
	results := r.Run()

	if results[0].IPv4Result != "override-instance" {
		t.Errorf("want override-instance, got %s", results[0].IPv4Result)
	}
	if results[0].IPv6Result != "override-instance" {
		t.Errorf("want override-instance (v6), got %s", results[0].IPv6Result)
	}
}
