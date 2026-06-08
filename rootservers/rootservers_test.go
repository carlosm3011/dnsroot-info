package rootservers

import (
	"testing"
)

func TestFilter_empty(t *testing.T) {
	got := Filter(nil)
	if len(got) != len(All) {
		t.Fatalf("expected %d servers, got %d", len(All), len(got))
	}
}

func TestFilter_subset(t *testing.T) {
	got := Filter([]string{"A", "K", "M"})
	if len(got) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(got))
	}
	letters := map[string]bool{"A": true, "K": true, "M": true}
	for _, s := range got {
		if !letters[s.Letter] {
			t.Errorf("unexpected server %s", s.Letter)
		}
	}
}

func TestFilter_caseInsensitive(t *testing.T) {
	lower := Filter([]string{"i", "k", "m"})
	upper := Filter([]string{"I", "K", "M"})
	if len(lower) != len(upper) {
		t.Fatalf("case sensitivity mismatch: %d vs %d", len(lower), len(upper))
	}
	for i := range lower {
		if lower[i].Letter != upper[i].Letter {
			t.Errorf("mismatch at %d: %s vs %s", i, lower[i].Letter, upper[i].Letter)
		}
	}
}

func TestFilter_preservesOrder(t *testing.T) {
	// Filter with reversed input should still return in All order.
	got := Filter([]string{"M", "A", "G"})
	want := []string{"A", "G", "M"}
	if len(got) != len(want) {
		t.Fatalf("expected %d, got %d", len(want), len(got))
	}
	for i, s := range got {
		if s.Letter != want[i] {
			t.Errorf("position %d: want %s, got %s", i, want[i], s.Letter)
		}
	}
}

func TestFilter_unknownLetters(t *testing.T) {
	got := Filter([]string{"X", "Y", "Z"})
	if len(got) != 0 {
		t.Errorf("expected empty result for unknown letters, got %d", len(got))
	}
}
