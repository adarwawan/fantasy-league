package handler

import (
	"testing"

	"fantasy-league/internal/store"
)

func TestComputeTeamBonus(t *testing.T) {
	lines := []store.PlayerStatGW{
		{PlayerID: "a", Name: "Alpha", Position: "MID", TeamShortName: "ARS", Bonus: 3},
		{PlayerID: "a", Name: "Alpha", Position: "MID", TeamShortName: "ARS", Bonus: 2},
		{PlayerID: "b", Name: "Bravo", Position: "DEF", TeamShortName: "ARS", Bonus: 5},
		{PlayerID: "c", Name: "Charlie", Position: "GK", TeamShortName: "ARS", Bonus: 1},
		{PlayerID: "d", Name: "Delta", Position: "FWD", TeamShortName: "ARS", Bonus: 1},
		{PlayerID: "e", Name: "Echo", Position: "FWD", TeamShortName: "ARS", Bonus: 0},
		{PlayerID: "f", Name: "Foxtrot", Position: "FWD", TeamShortName: "LIV", Bonus: 4},
		// Club with no finished match in the window must be dropped.
		{PlayerID: "g", Name: "Golf", Position: "MID", TeamShortName: "NEW", Bonus: 2},
	}
	// ARS played a double gameweek: matches counts fixtures, not gameweeks.
	matches := map[string]int{"ARS": 6, "LIV": 5, "NEW": 0}

	got := computeTeamBonus(lines, matches, 3)
	if len(got) != 2 {
		t.Fatalf("got %d teams, want 2 (NEW has no matches): %+v", len(got), got)
	}

	ars := got[0]
	if ars.Team != "ARS" || ars.Matches != 6 || ars.Bonus != 12 {
		t.Errorf("ARS = %+v, want matches 6 bonus 12", ars)
	}
	if want := (teamBonusByPosition{GK: 1, DEF: 5, MID: 5, FWD: 1}); ars.ByPosition != want {
		t.Errorf("ARS by position = %+v, want %+v", ars.ByPosition, want)
	}
	// Alpha (5) ties Bravo (5): name ascending. Zero-bonus Echo is excluded and
	// the list is capped at 3.
	wantTop := []string{"Alpha", "Bravo", "Charlie"}
	if len(ars.Top) != 3 {
		t.Fatalf("ARS top = %+v, want 3 players", ars.Top)
	}
	for i, n := range wantTop {
		if ars.Top[i].Name != n {
			t.Errorf("ARS top[%d] = %s, want %s", i, ars.Top[i].Name, n)
		}
	}
	// Per-position totals must sum to the club total (lossless split).
	p := ars.ByPosition
	if p.GK+p.DEF+p.MID+p.FWD != ars.Bonus {
		t.Errorf("position split %+v does not sum to %d", p, ars.Bonus)
	}

	if got[1].Team != "LIV" || got[1].Bonus != 4 || len(got[1].Top) != 1 {
		t.Errorf("LIV = %+v", got[1])
	}
}

func TestComputeTeamBonusEmptyTopIsNotNil(t *testing.T) {
	got := computeTeamBonus(nil, map[string]int{"ARS": 3}, 3)
	if len(got) != 1 || got[0].Top == nil || got[0].Bonus != 0 {
		t.Errorf("got %+v, want one ARS entry with empty non-nil Top", got)
	}
}
