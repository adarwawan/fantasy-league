package musthave

import (
	"testing"

	"fantasy-league/internal/store"
)

func testConfig() Config {
	return DefaultConfig() // window 5, min 6 pts, ratio 0.5, FDR <= 3, top 4/8/8/5
}

func candidate(id, pos string, fixtures ...store.FixtureInfo) store.PlayerRow {
	return store.PlayerRow{ID: id, Position: pos, Status: "available", Fixtures: fixtures}
}

func fixture(gw, difficulty int) store.FixtureInfo {
	return store.FixtureInfo{GW: gw, Difficulty: difficulty}
}

func TestCompute_allConditionsMet(t *testing.T) {
	cands := []store.PlayerRow{candidate("a", "MID", fixture(6, 2))}
	pool := []store.PlayerOwnership{{PlayerID: "a", Position: "MID", GlobalOwnership: 50}}
	points := map[string][]int{"a": {8, 2, 7, 6, 3}} // 3 hits of 5

	flags := Compute(cands, pool, points, 5, 6, testConfig())
	if !flags["a"] {
		t.Errorf("expected a to be must-have")
	}
}

func TestCompute_failsEachCondition(t *testing.T) {
	pool := []store.PlayerOwnership{
		{PlayerID: "a", Position: "MID", GlobalOwnership: 50},
	}
	goodPoints := map[string][]int{"a": {8, 7, 6, 0, 0}}
	cfg := testConfig()

	cases := []struct {
		name   string
		cand   store.PlayerRow
		points map[string][]int
		nextGW int
	}{
		{"unavailable", store.PlayerRow{ID: "a", Position: "MID", Status: "injured", Fixtures: []store.FixtureInfo{fixture(6, 2)}}, goodPoints, 6},
		{"no next fixture", candidate("a", "MID", fixture(8, 2)), goodPoints, 6},
		{"hard fixture", candidate("a", "MID", fixture(6, 4)), goodPoints, 6},
		{"bad form", candidate("a", "MID", fixture(6, 2)), map[string][]int{"a": {8, 5, 5, 0, 0}}, 6},
		{"no points at all", candidate("a", "MID", fixture(6, 2)), map[string][]int{}, 6},
		{"no next gw", candidate("a", "MID", fixture(6, 2)), goodPoints, 0},
	}
	for _, tc := range cases {
		flags := Compute([]store.PlayerRow{tc.cand}, pool, tc.points, 5, tc.nextGW, cfg)
		if flags["a"] {
			t.Errorf("%s: expected a to not be must-have", tc.name)
		}
	}
}

func TestCompute_ownershipRankCutoff(t *testing.T) {
	// 5 FWDs; cutoff is top 5, so a 6th-ranked FWD misses even with good form.
	var pool []store.PlayerOwnership
	var cands []store.PlayerRow
	points := map[string][]int{}
	ids := []string{"f1", "f2", "f3", "f4", "f5", "f6"}
	for i, id := range ids {
		pool = append(pool, store.PlayerOwnership{PlayerID: id, Position: "FWD", GlobalOwnership: float64(60 - i)})
		cands = append(cands, candidate(id, "FWD", fixture(6, 2)))
		points[id] = []int{8, 8, 8, 8, 8}
	}

	flags := Compute(cands, pool, points, 5, 6, testConfig())
	for _, id := range ids[:5] {
		if !flags[id] {
			t.Errorf("expected %s (top 5) to be must-have", id)
		}
	}
	if flags["f6"] {
		t.Errorf("expected f6 (rank 6) to not be must-have")
	}
}

func TestCompute_rankIsPerPosition(t *testing.T) {
	// A GK with lower ownership than many MIDs still ranks 1st among GKs.
	pool := []store.PlayerOwnership{
		{PlayerID: "gk", Position: "GK", GlobalOwnership: 10},
		{PlayerID: "m1", Position: "MID", GlobalOwnership: 50},
		{PlayerID: "m2", Position: "MID", GlobalOwnership: 40},
	}
	cands := []store.PlayerRow{candidate("gk", "GK", fixture(6, 2))}
	points := map[string][]int{"gk": {8, 8, 8}}

	flags := Compute(cands, pool, points, 5, 6, testConfig())
	if !flags["gk"] {
		t.Errorf("expected gk to be must-have")
	}
}

func TestCompute_adaptiveThreshold(t *testing.T) {
	// Only 2 finished GWs → need ceil(0.5*2)=1 hit.
	cands := []store.PlayerRow{candidate("a", "MID", fixture(3, 2))}
	pool := []store.PlayerOwnership{{PlayerID: "a", Position: "MID", GlobalOwnership: 50}}

	flags := Compute(cands, pool, map[string][]int{"a": {7, 0}}, 2, 3, testConfig())
	if !flags["a"] {
		t.Errorf("expected a to be must-have with 1 hit in 2 GWs")
	}
	// Zero GWs counted → threshold floors at 1, zero hits can never qualify.
	flags = Compute(cands, pool, map[string][]int{}, 0, 3, testConfig())
	if flags["a"] {
		t.Errorf("expected a to not be must-have with no GWs played")
	}
}

func TestCompute_doubleGameweek(t *testing.T) {
	// Two fixtures in the next GW: one easy is enough.
	cands := []store.PlayerRow{candidate("a", "MID", fixture(6, 5), fixture(6, 2))}
	pool := []store.PlayerOwnership{{PlayerID: "a", Position: "MID", GlobalOwnership: 50}}
	points := map[string][]int{"a": {8, 8, 8, 8, 8}}

	flags := Compute(cands, pool, points, 5, 6, testConfig())
	if !flags["a"] {
		t.Errorf("expected a to be must-have via easier DGW fixture")
	}
}

func TestCompute_nextGWAlreadyPlayed(t *testing.T) {
	// Player already played their nextGW (6) match, so Fixtures starts at 7
	// (nextGW+1). A good fixture there still qualifies them.
	cands := []store.PlayerRow{candidate("a", "MID", fixture(7, 2))}
	pool := []store.PlayerOwnership{{PlayerID: "a", Position: "MID", GlobalOwnership: 50}}
	points := map[string][]int{"a": {8, 8, 8, 8, 8}}

	flags := Compute(cands, pool, points, 5, 6, testConfig())
	if !flags["a"] {
		t.Errorf("expected a to be must-have via nextGW+1 fixture")
	}
}

func TestOwnershipRanks_tieBreakByID(t *testing.T) {
	pool := []store.PlayerOwnership{
		{PlayerID: "b", Position: "MID", GlobalOwnership: 30},
		{PlayerID: "a", Position: "MID", GlobalOwnership: 30},
	}
	ranks := ownershipRanks(pool)
	if ranks["a"] != 1 || ranks["b"] != 2 {
		t.Errorf("expected deterministic tie-break a=1 b=2, got %v", ranks)
	}
}
