package handler

import (
	"testing"

	"fantasy-league/internal/store"
)

func TestDefensiveContributionPoints(t *testing.T) {
	cases := []struct {
		position string
		actions  int
		want     int
	}{
		{"DEF", 9, 0},
		{"DEF", 10, 2},  // defender threshold
		{"DEF", 25, 2},  // capped at 2 per GW
		{"MID", 11, 0},
		{"MID", 12, 2},  // mid/fwd threshold
		{"FWD", 12, 2},
		{"FWD", 11, 0},
		{"GK", 50, 0},   // keepers earn no DC points
		{"", 100, 0},    // unknown position
	}
	for _, c := range cases {
		if got := defensiveContributionPoints(c.position, c.actions); got != c.want {
			t.Errorf("defensiveContributionPoints(%q, %d) = %d, want %d", c.position, c.actions, got, c.want)
		}
	}
}

func TestComponentPoints(t *testing.T) {
	cases := []struct {
		position, component string
		count, want         int
	}{
		{"GK", "goals", 1, 6},
		{"DEF", "goals", 2, 12},
		{"MID", "goals", 2, 10},
		{"FWD", "goals", 3, 12},
		{"DEF", "assists", 2, 6},
		{"FWD", "assists", 1, 3},
		{"GK", "clean_sheet", 3, 12},
		{"DEF", "clean_sheet", 2, 8},
		{"MID", "clean_sheet", 4, 4},
		{"FWD", "clean_sheet", 5, 0}, // forwards earn nothing for clean sheets
	}
	for _, c := range cases {
		if got := componentPoints(c.position, c.component, c.count); got != c.want {
			t.Errorf("componentPoints(%q, %q, %d) = %d, want %d", c.position, c.component, c.count, got, c.want)
		}
	}
}

func TestComputeStatLeaders(t *testing.T) {
	// Two finished GWs per player, mixed positions. Card values are FPL points.
	lines := []store.PlayerStatGW{
		// Defender A: 1 goal (→6 pts), DC threshold met both GWs → 4 DC points.
		{PlayerID: "dA", Position: "DEF", Name: "Alpha", TeamShortName: "ARS", Goals: 1, DefensiveContribution: 12},
		{PlayerID: "dA", Position: "DEF", Name: "Alpha", TeamShortName: "ARS", Goals: 0, DefensiveContribution: 10},
		// Defender B: 2 goals (→12 pts); DC met once (10) missed once (9) → 2 DC points.
		{PlayerID: "dB", Position: "DEF", Name: "Bravo", TeamShortName: "LIV", Goals: 2, DefensiveContribution: 9},
		{PlayerID: "dB", Position: "DEF", Name: "Bravo", TeamShortName: "LIV", Goals: 0, DefensiveContribution: 10},
		// Midfielder C: 3 assists (→9 pts); DC uses the 12 threshold; 11 is below → 0.
		{PlayerID: "mC", Position: "MID", Name: "Charlie", TeamShortName: "MCI", Assists: 3, DefensiveContribution: 11},
	}

	leaders := computeStatLeaders(lines, 5)

	get := func(pos, comp string) []statLeader {
		var out []statLeader
		for _, l := range leaders {
			if l.Position == pos && l.Component == comp {
				out = append(out, l)
			}
		}
		return out
	}

	// DEF goals (points): Bravo 12 ranks above Alpha 6.
	goals := get("DEF", "goals")
	if len(goals) != 2 || goals[0].Name != "Bravo" || goals[0].Value != 12 || goals[0].Rank != 1 {
		t.Fatalf("unexpected DEF goals leaders: %+v", goals)
	}
	if goals[1].Name != "Alpha" || goals[1].Value != 6 || goals[1].Rank != 2 {
		t.Errorf("expected Alpha 6pts rank 2 in DEF goals, got %+v", goals[1])
	}

	// DEF defensive_con: Alpha 4 pts (both GWs) above Bravo 2 pts.
	dc := get("DEF", "defensive_con")
	if len(dc) != 2 || dc[0].Name != "Alpha" || dc[0].Value != 4 || dc[1].Name != "Bravo" || dc[1].Value != 2 {
		t.Fatalf("unexpected DEF defensive_con leaders: %+v", dc)
	}

	// MID defensive_con: Charlie never hit 12 → excluded (no zero-value cards).
	if got := get("MID", "defensive_con"); len(got) != 0 {
		t.Errorf("expected no MID defensive_con leaders, got %+v", got)
	}
	// MID assists: Charlie 3 assists → 9 points.
	if a := get("MID", "assists"); len(a) != 1 || a[0].Value != 9 {
		t.Errorf("unexpected MID assists leaders: %+v", a)
	}
}

func TestComputeStatLeadersRespectsLimit(t *testing.T) {
	var lines []store.PlayerStatGW
	for i := 0; i < 8; i++ {
		lines = append(lines, store.PlayerStatGW{
			PlayerID: string(rune('a' + i)), Position: "FWD",
			Name: string(rune('a' + i)), TeamShortName: "XXX", Goals: i + 1,
		})
	}
	leaders := computeStatLeaders(lines, 5)
	n := 0
	for _, l := range leaders {
		if l.Position == "FWD" && l.Component == "goals" {
			n++
		}
	}
	if n != 5 {
		t.Errorf("expected limit of 5 FWD goals leaders, got %d", n)
	}
}

func TestBuildStatsResponse(t *testing.T) {
	leaders := []statLeader{
		{Position: "GK", Component: "clean_sheet", Rank: 1, PlayerID: "g1", Name: "Keeper", Team: "ARS", Value: 4},
		{Position: "DEF", Component: "goals", Rank: 1, PlayerID: "d1", Name: "Backman", Team: "LIV", Value: 2},
		{Position: "DEF", Component: "defensive_con", Rank: 1, PlayerID: "d2", Name: "Tackler", Team: "EVE", Value: 4},
		{Position: "FWD", Component: "goals", Rank: 1, PlayerID: "f1", Name: "Striker", Team: "MCI", Value: 6},
	}

	resp := buildStatsResponse("fpl", 12, leaders)

	if got := len(resp.Sections); got != 4 {
		t.Fatalf("expected 4 sections, got %d", got)
	}
	wantOrder := []string{"GK", "DEF", "MID", "FWD"}
	for i, w := range wantOrder {
		if resp.Sections[i].Position != w {
			t.Errorf("section %d: expected %s, got %s", i, w, resp.Sections[i].Position)
		}
	}

	// GK section: Clean Sheet then Bonus.
	gk := resp.Sections[0]
	if len(gk.Cards) != 2 || gk.Cards[0].Component != "clean_sheet" || gk.Cards[0].Points != "+4" {
		t.Fatalf("unexpected GK cards: %+v", gk.Cards)
	}
	if len(gk.Cards[0].Leaders) != 1 || gk.Cards[0].Leaders[0].Name != "Keeper" {
		t.Errorf("expected Keeper as clean-sheet leader, got %+v", gk.Cards[0].Leaders)
	}
	// Bonus card has no data → present but empty (stable layout, non-null JSON).
	if gk.Cards[1].Component != "bonus" || gk.Cards[1].Leaders == nil || len(gk.Cards[1].Leaders) != 0 {
		t.Errorf("expected empty non-nil bonus leaders, got %+v", gk.Cards[1].Leaders)
	}

	// DEF section emits all five spec cards.
	if got := len(resp.Sections[1].Cards); got != 5 {
		t.Errorf("expected 5 DEF cards, got %d", got)
	}

	if resp.Meta.Window != statsWindow || resp.Meta.GW != 12 || resp.Meta.GameID != "fpl" {
		t.Errorf("unexpected meta: %+v", resp.Meta)
	}
}
