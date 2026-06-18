package fpl

import (
	"testing"
	"time"
)

func TestMapTeams(t *testing.T) {
	raw := []fplTeam{
		{ID: 1, Name: "Arsenal", ShortName: "ARS"},
		{ID: 14, Name: "Liverpool", ShortName: "LIV"},
	}
	teams := mapTeams(raw)
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	if teams[1].ShortName != "LIV" {
		t.Errorf("expected LIV, got %s", teams[1].ShortName)
	}
	if teams[0].GameID != "fpl" {
		t.Errorf("expected game_id=fpl, got %s", teams[0].GameID)
	}
}

func TestMapPlayers(t *testing.T) {
	raw := []fplPlayer{
		{ID: 328, WebName: "Salah", Team: 14, ElementType: 3, NowCost: 129, Form: "8.4", SelectedByPercent: "43.2", Status: "a", News: ""},
		{ID: 1,   WebName: "Raya",  Team: 1,  ElementType: 1, NowCost: 55,  Form: "0",   SelectedByPercent: "5.1",  Status: "d", News: "Knock"},
		{ID: 2,   WebName: "Saka",  Team: 1,  ElementType: 3, NowCost: 99,  Form: "",    SelectedByPercent: "",     Status: "i", News: "Injured"},
	}
	players := mapPlayers(raw)

	if len(players) != 3 {
		t.Fatalf("expected 3, got %d", len(players))
	}

	salah := players[0]
	if salah.Price != 12.9 {
		t.Errorf("Salah price: expected 12.9, got %f", salah.Price)
	}
	if salah.Position != "MID" {
		t.Errorf("Salah position: expected MID, got %s", salah.Position)
	}
	if salah.Form != 8.4 {
		t.Errorf("Salah form: expected 8.4, got %f", salah.Form)
	}
	if salah.GlobalOwnership != 43.2 {
		t.Errorf("Salah global_own: expected 43.2, got %f", salah.GlobalOwnership)
	}
	if salah.Status != "available" {
		t.Errorf("Salah status: expected available, got %s", salah.Status)
	}

	raya := players[1]
	if raya.Position != "GK" {
		t.Errorf("Raya position: expected GK, got %s", raya.Position)
	}
	if raya.Status != "doubt" {
		t.Errorf("Raya status: expected doubt, got %s", raya.Status)
	}
	if raya.Form != 0 {
		t.Errorf("Raya form: expected 0, got %f", raya.Form)
	}

	saka := players[2]
	if saka.Status != "injured" {
		t.Errorf("Saka status: expected injured, got %s", saka.Status)
	}
	if saka.Form != 0 {
		t.Errorf("Saka form: expected 0 for empty string, got %f", saka.Form)
	}
}

func TestMapFixtures_skipsNullEvent(t *testing.T) {
	gw := 30
	raw := []fplFixture{
		{ID: 1, Event: &gw, TeamH: 14, TeamA: 1, TeamHDifficulty: 2, TeamADifficulty: 4, KickoffTime: "2024-03-15T15:00:00Z", Finished: false},
		{ID: 2, Event: nil, TeamH: 2,  TeamA: 3},
	}
	fixtures := mapFixtures(raw)
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture (null event skipped), got %d", len(fixtures))
	}
	if fixtures[0].GW != 30 {
		t.Errorf("expected GW 30, got %d", fixtures[0].GW)
	}
	if fixtures[0].HomeTeamID != "14" {
		t.Errorf("expected HomeTeamID=14, got %s", fixtures[0].HomeTeamID)
	}
}

func TestMapFixtures_badKickoff(t *testing.T) {
	gw := 1
	raw := []fplFixture{
		{ID: 10, Event: &gw, TeamH: 1, TeamA: 2, KickoffTime: ""},
	}
	fixtures := mapFixtures(raw)
	if len(fixtures) != 1 {
		t.Fatalf("expected 1")
	}
	if !fixtures[0].KickoffTime.IsZero() {
		t.Errorf("expected zero kickoff for empty string, got %v", fixtures[0].KickoffTime)
	}
}

func TestMapManagers(t *testing.T) {
	entries := []fplStandingEntry{
		{Entry: 100, EntryName: "Top Team", Rank: 1},
		{Entry: 200, EntryName: "Second Team", Rank: 2},
	}
	managers := mapManagers(entries)
	if len(managers) != 2 {
		t.Fatalf("expected 2, got %d", len(managers))
	}
	if managers[0].ExternalID != 100 {
		t.Errorf("expected ExternalID=100, got %d", managers[0].ExternalID)
	}
	if managers[0].OverallRank != 1 {
		t.Errorf("expected OverallRank=1, got %d", managers[0].OverallRank)
	}
	if managers[0].GameID != "fpl" {
		t.Errorf("expected GameID=fpl")
	}
}

func TestMapPicks(t *testing.T) {
	raw := []fplPick{
		{Element: 328, Position: 1, Multiplier: 2, IsCaptain: true,  IsViceCaptain: false},
		{Element: 1,   Position: 2, Multiplier: 1, IsCaptain: false, IsViceCaptain: true},
	}
	picks := mapPicks(999, 30, raw)
	if len(picks) != 2 {
		t.Fatalf("expected 2, got %d", len(picks))
	}
	if picks[0].ManagerID != "999" {
		t.Errorf("expected ManagerID=999, got %s", picks[0].ManagerID)
	}
	if picks[0].PlayerID != "328" {
		t.Errorf("expected PlayerID=328, got %s", picks[0].PlayerID)
	}
	if !picks[0].IsCaptain {
		t.Errorf("expected IsCaptain=true")
	}
	if picks[1].Multiplier != 1 {
		t.Errorf("expected Multiplier=1, got %d", picks[1].Multiplier)
	}
}

func TestCurrentGW(t *testing.T) {
	events := []fplEvent{
		{ID: 28, IsCurrent: false, IsNext: false},
		{ID: 29, IsCurrent: true,  IsNext: false},
		{ID: 30, IsCurrent: false, IsNext: true},
	}
	gw := currentGW(events)
	if gw != 29 {
		t.Errorf("expected current GW=29, got %d", gw)
	}
}

func TestCurrentGW_fallbackToNext(t *testing.T) {
	events := []fplEvent{
		{ID: 1, IsCurrent: false, IsNext: true},
	}
	gw := currentGW(events)
	if gw != 0 {
		t.Errorf("expected fallback GW=0 (next-1), got %d", gw)
	}
}

func TestParseFloat(t *testing.T) {
	cases := []struct{ in string; want float64 }{
		{"8.4", 8.4},
		{"0", 0},
		{"", 0},
		{"  43.2  ", 43.2},
	}
	for _, c := range cases {
		got := parseFloat(c.in)
		if got != c.want {
			t.Errorf("parseFloat(%q): expected %f, got %f", c.in, c.want, got)
		}
	}
}

// compile-time check that UpdatedAt is set
func TestMapTeams_updatedAt(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	teams := mapTeams([]fplTeam{{ID: 1, Name: "X", ShortName: "X"}})
	if teams[0].UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt not set correctly")
	}
}
