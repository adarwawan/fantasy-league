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

// oddsFixture carries bookmaker odds; a hard FDR (5) proves the odds path, not
// the FDR fallback, is what qualifies (or disqualifies) the fixture.
func oddsFixture(gw int, xg, csPct float64) store.FixtureInfo {
	return store.FixtureInfo{GW: gw, Difficulty: 5, XG: &xg, CSPct: &csPct}
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

func TestCompute_oddsGreenFixture(t *testing.T) {
	// Odds present and above cutoff qualify despite a hard FDR (5); below the
	// cutoff they don't. Defenders judged on CS%, attackers on xG.
	cfg := testConfig() // MinXG 2.0, MinCSPct 50
	form := map[string][]int{"a": {8, 8, 8, 8, 8}}

	cases := []struct {
		name string
		pos  string
		fix  store.FixtureInfo
		want bool
	}{
		{"attacker green xG", "MID", oddsFixture(6, 2.3, 10), true},
		{"attacker low xG", "MID", oddsFixture(6, 1.5, 90), false},
		{"defender green CS", "DEF", oddsFixture(6, 0.5, 55), true},
		{"defender low CS", "DEF", oddsFixture(6, 3.0, 40), false},
		{"gk green CS", "GK", oddsFixture(6, 0.5, 50), true},
	}
	for _, tc := range cases {
		pool := []store.PlayerOwnership{{PlayerID: "a", Position: tc.pos, GlobalOwnership: 50}}
		flags := Compute([]store.PlayerRow{candidate("a", tc.pos, tc.fix)}, pool, form, 5, 6, cfg)
		if flags["a"] != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, flags["a"], tc.want)
		}
	}
}

func TestCompute_missingOddsFallsBackToFDR(t *testing.T) {
	// No odds on the fixture → decision falls back to FDR <= MaxNextFDR.
	pool := []store.PlayerOwnership{{PlayerID: "a", Position: "MID", GlobalOwnership: 50}}
	form := map[string][]int{"a": {8, 8, 8, 8, 8}}

	easy := Compute([]store.PlayerRow{candidate("a", "MID", fixture(6, 2))}, pool, form, 5, 6, testConfig())
	if !easy["a"] {
		t.Errorf("expected FDR-2 fallback fixture to qualify")
	}
	hard := Compute([]store.PlayerRow{candidate("a", "MID", fixture(6, 4))}, pool, form, 5, 6, testConfig())
	if hard["a"] {
		t.Errorf("expected FDR-4 fallback fixture to not qualify")
	}
}

func TestCompute_mixedOddsAndFallbackInDGW(t *testing.T) {
	// A DGW with one odds-priced fixture (below cutoff) and one un-priced but
	// easy-FDR fixture: the FDR fallback on the second still qualifies.
	pool := []store.PlayerOwnership{{PlayerID: "a", Position: "MID", GlobalOwnership: 50}}
	form := map[string][]int{"a": {8, 8, 8, 8, 8}}
	cands := []store.PlayerRow{candidate("a", "MID", oddsFixture(6, 1.0, 10), fixture(6, 2))}

	flags := Compute(cands, pool, form, 5, 6, testConfig())
	if !flags["a"] {
		t.Errorf("expected qualification via easy-FDR leg of the DGW")
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
