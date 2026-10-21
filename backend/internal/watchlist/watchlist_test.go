package watchlist

import (
	"testing"

	"fantasy-league/internal/musthave"
	"fantasy-league/internal/store"
)

func player(id string, fixtures ...store.FixtureInfo) store.PlayerRow {
	return store.PlayerRow{ID: id, Position: "MID", Status: "available", Fixtures: fixtures}
}

func inputs(own float64, p store.PlayerRow, points []int) musthave.Inputs {
	return musthave.Inputs{
		Players:    []store.PlayerRow{p},
		Pool:       []store.PlayerOwnership{{PlayerID: p.ID, Position: p.Position, GlobalOwnership: own}},
		Points:     map[string][]int{p.ID: points},
		GWsCounted: 5,
		NextGW:     6,
	}
}

func TestCompute_ownershipBand(t *testing.T) {
	good := player("a", store.FixtureInfo{GW: 6, Difficulty: 2})
	pts := []int{8, 7, 6, 0, 0}
	cases := []struct {
		name string
		own  float64
		want bool
	}{
		{"below floor", 1.9, false},
		{"at floor", 2, true},
		{"inside", 5, true},
		{"just below ceiling", 9.9, true},
		{"at ceiling", 10, false},
		{"above ceiling", 40, false},
	}
	for _, tc := range cases {
		got := Compute(inputs(tc.own, good, pts), musthave.DefaultConfig(), DefaultConfig(), nil)
		if got["a"] != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got["a"], tc.want)
		}
	}
}

func TestCompute_sharesMustHaveRules(t *testing.T) {
	pts := []int{8, 7, 6, 0, 0}
	cases := []struct {
		name string
		p    store.PlayerRow
		pts  []int
	}{
		{"unavailable", store.PlayerRow{ID: "a", Position: "MID", Status: "injured", Fixtures: []store.FixtureInfo{{GW: 6, Difficulty: 2}}}, pts},
		{"hard fixture", player("a", store.FixtureInfo{GW: 6, Difficulty: 4}), pts},
		{"no fixture", player("a"), pts},
		{"bad form", player("a", store.FixtureInfo{GW: 6, Difficulty: 2}), []int{8, 5, 5, 0, 0}},
	}
	for _, tc := range cases {
		got := Compute(inputs(5, tc.p, tc.pts), musthave.DefaultConfig(), DefaultConfig(), nil)
		if got["a"] {
			t.Errorf("%s: should not be on watchlist", tc.name)
		}
	}
}

func TestCompute_mustHaveWinsOverlap(t *testing.T) {
	p := player("a", store.FixtureInfo{GW: 6, Difficulty: 2})
	got := Compute(inputs(5, p, []int{8, 7, 6, 0, 0}), musthave.DefaultConfig(), DefaultConfig(), map[string]bool{"a": true})
	if got["a"] {
		t.Error("must-have player should be excluded from watchlist")
	}
}

func TestCompute_noGWReturnsEmpty(t *testing.T) {
	p := player("a", store.FixtureInfo{GW: 6, Difficulty: 2})
	in := inputs(5, p, []int{8, 7, 6, 0, 0})
	in.NextGW = 0
	if got := Compute(in, musthave.DefaultConfig(), DefaultConfig(), nil); got == nil || len(got) != 0 {
		t.Errorf("want empty non-nil map, got %v", got)
	}
}
