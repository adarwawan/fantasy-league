package handler

import "testing"

// These guard the cache-key cardinality bounds: unvalidated query params must
// collapse to a small finite domain so they can't be used to bust the cache
// and hammer Postgres.

func TestCanonicalPos(t *testing.T) {
	cases := map[string]string{
		"GK": "GK", "def": "DEF", " mid ": "MID", "FWD": "FWD",
		"": "", "xyz": "", "'; DROP": "",
	}
	for in, want := range cases {
		if got := canonicalPos(in); got != want {
			t.Errorf("canonicalPos(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCanonicalPlayerSort(t *testing.T) {
	cases := map[string]string{
		"form": "form", "price": "price", "name": "name",
		"top_n_ownership": "top_n_ownership", "global_ownership": "global_ownership",
		"": "global_ownership", "bogus": "global_ownership",
	}
	for in, want := range cases {
		if got := canonicalPlayerSort(in); got != want {
			t.Errorf("canonicalPlayerSort(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClampMaxPrice(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{-5, 0}, {0, 0}, {4.5, 4.5}, {4.567, 4.6}, {1000, maxPlayerPrice},
	}
	for _, c := range cases {
		if got := clampMaxPrice(c.in); got != c.want {
			t.Errorf("clampMaxPrice(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCanonicalTeamSort(t *testing.T) {
	cases := map[string]string{
		"ovr_form": "ovr_form", "xg_sum": "xg_sum", "cs_avg": "cs_avg",
		"": "ovr_form", "bogus": "ovr_form",
	}
	for in, want := range cases {
		if got := canonicalTeamSort(in); got != want {
			t.Errorf("canonicalTeamSort(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClampGW(t *testing.T) {
	cases := []struct {
		v, def, want int
	}{
		{0, 5, 5}, {-3, 5, 5}, {10, 5, 10}, {99, 5, maxGW},
		{0, 0, 0}, {40, 0, maxGW},
	}
	for _, c := range cases {
		if got := clampGW(c.v, c.def); got != c.want {
			t.Errorf("clampGW(%d, %d) = %d, want %d", c.v, c.def, got, c.want)
		}
	}
}
