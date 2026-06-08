package dns

import (
	"errors"
	"testing"
)

// stubQuerier is a test double for Querier.
type stubQuerier struct {
	responses map[string]string
	errors    map[string]error
}

func (s *stubQuerier) QueryCHAOS(serverAddr string) (string, error) {
	if err, ok := s.errors[serverAddr]; ok {
		return "", err
	}
	if r, ok := s.responses[serverAddr]; ok {
		return r, nil
	}
	return "", errors.New("no response configured for " + serverAddr)
}

// Compile-time check that RealQuerier and stubQuerier implement Querier.
var _ Querier = (*RealQuerier)(nil)
var _ Querier = (*stubQuerier)(nil)

func TestStubQuerier_response(t *testing.T) {
	q := &stubQuerier{
		responses: map[string]string{"1.2.3.4": "a1-iad"},
	}
	got, err := q.QueryCHAOS("1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "a1-iad" {
		t.Errorf("want a1-iad, got %s", got)
	}
}

func TestStubQuerier_error(t *testing.T) {
	sentinel := errors.New("i/o timeout")
	q := &stubQuerier{
		errors: map[string]error{"1.2.3.4": sentinel},
	}
	_, err := q.QueryCHAOS("1.2.3.4")
	if !errors.Is(err, sentinel) {
		t.Errorf("want sentinel error, got %v", err)
	}
}

func TestStubQuerier_missing(t *testing.T) {
	q := &stubQuerier{}
	_, err := q.QueryCHAOS("9.9.9.9")
	if err == nil {
		t.Error("expected error for unconfigured address")
	}
}
