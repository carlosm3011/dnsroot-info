package cmd

import (
	"reflect"
	"testing"
)

func TestParseLetters_empty(t *testing.T) {
	if got := parseLetters(""); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestParseLetters_single(t *testing.T) {
	got := parseLetters("A")
	want := []string{"A"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestParseLetters_multiple(t *testing.T) {
	got := parseLetters("I,K,M")
	want := []string{"I", "K", "M"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestParseLetters_whitespace(t *testing.T) {
	got := parseLetters(" A , B , C ")
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestParseLetters_emptySegments(t *testing.T) {
	got := parseLetters("A,,B")
	want := []string{"A", "B"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}
