package wcf

import (
	"testing"
	"time"

	"fantasy-league/internal/fantasy"
)

func TestMapTeams(t *testing.T) {
	raw := []wcfSquad{
		{ID: 1, Name: "Brazil", Abbr: "BRA"},
		{ID: 2, Name: "France", Abbr: "FRA"},
	}
	teams := mapTeams(raw)
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	if teams[0].GameID != "wcf" {
		t.Errorf("expected game_id=wcf, got %s", teams[0].GameID)
	}
	if teams[1].ShortName != "FRA" {
		t.Errorf("expected FRA, got %s", teams[1].ShortName)
	}
}

func TestMapPlayers_basic(t *testing.T) {
	knownName := "Mbappe"
	raw := []wcfPlayer{
		{
			ID: 10, FirstName: "Kylian", LastName: "Mbappé",
			KnownName: &knownName, SquadID: 2, Position: "FWD",
			Price: 13.0, PercentSelected: 55.0, Status: "playing",
			Stats: wcfPlayerStats{Form: 9.1},
		},
		{
			ID: 20, FirstName: "Alisson", LastName: "Becker",
			KnownName: nil, SquadID: 1, Position: "GK",
			Price: 5.5, PercentSelected: 20.0, Status: "suspended",
			Stats: wcfPlayerStats{Form: 6.0},
		},
	}
	players := mapPlayers(raw)

	if len(players) != 2 {
		t.Fatalf("expected 2, got %d", len(players))
	}

	mbappe := players[0]
	if mbappe.Name != "Mbappe" {
		t.Errorf("expected knownName Mbappe, got %s", mbappe.Name)
	}
	if mbappe.Position != "FWD" {
		t.Errorf("expected FWD, got %s", mbappe.Position)
	}
	if mbappe.Price != 13.0 {
		t.Errorf("expected price 13.0, got %f", mbappe.Price)
	}
	if mbappe.Form != 9.1 {
		t.Errorf("expected form 9.1, got %f", mbappe.Form)
	}
	if mbappe.GlobalOwnership != 55.0 {
		t.Errorf("expected global_own 55.0, got %f", mbappe.GlobalOwnership)
	}
	if mbappe.Status != "available" {
		t.Errorf("expected available, got %s", mbappe.Status)
	}
	if mbappe.GameID != "wcf" {
		t.Errorf("expected game_id=wcf")
	}

	alisson := players[1]
	// No knownName → firstName + lastName
	if alisson.Name != "Alisson Becker" {
		t.Errorf("expected 'Alisson Becker', got %s", alisson.Name)
	}
	if alisson.Position != "GK" {
		t.Errorf("expected GK, got %s", alisson.Position)
	}
	if alisson.Status != "injured" {
		t.Errorf("suspended → injured, got %s", alisson.Status)
	}
}

func TestMapPlayers_statusFallbacks(t *testing.T) {
	raw := []wcfPlayer{
		{ID: 1, FirstName: "X", LastName: "X", Position: "MID", Status: "transferred"},
		{ID: 2, FirstName: "Y", LastName: "Y", Position: "DEF", Status: ""},
		{ID: 3, FirstName: "Z", LastName: "Z", Position: "UNKNOWN", Status: "playing"},
	}
	players := mapPlayers(raw)

	if players[0].Status != "injured" {
		t.Errorf("transferred → injured, got %s", players[0].Status)
	}
	if players[1].Status != "available" {
		t.Errorf("empty status → available, got %s", players[1].Status)
	}
	if players[2].Position != "MID" {
		t.Errorf("unknown position → MID fallback, got %s", players[2].Position)
	}
}

func TestMapFixtures(t *testing.T) {
	rounds := []wcfRound{
		{
			ID: 1,
			Tournaments: []wcfTournament{
				{ID: 1, HomeSquadID: 28, AwaySquadID: 40, Date: "2026-06-11T20:00:00+01:00", Status: "complete"},
				{ID: 2, HomeSquadID: 5, AwaySquadID: 10, Date: "", Status: "scheduled"},
			},
		},
		{
			ID: 2,
			Tournaments: []wcfTournament{
				{ID: 3, HomeSquadID: 1, AwaySquadID: 2, Date: "2026-06-15T17:00:00Z", Status: "scheduled"},
			},
		},
	}
	fixtures := mapFixtures(rounds)

	if len(fixtures) != 3 {
		t.Fatalf("expected 3 fixtures, got %d", len(fixtures))
	}

	f0 := fixtures[0]
	if f0.GW != 1 {
		t.Errorf("expected GW=1, got %d", f0.GW)
	}
	if f0.HomeTeamID != "28" {
		t.Errorf("expected HomeTeamID=28, got %s", f0.HomeTeamID)
	}
	if f0.HomeDifficulty != 3 || f0.AwayDifficulty != 3 {
		t.Errorf("expected default difficulty=3")
	}
	if !f0.Finished {
		t.Errorf("status=complete → finished=true")
	}
	if f0.KickoffTime.IsZero() {
		t.Errorf("expected non-zero kickoff")
	}

	f1 := fixtures[1]
	if !f1.KickoffTime.IsZero() {
		t.Errorf("empty date → zero kickoff")
	}
	if f1.Finished {
		t.Errorf("status=scheduled → finished=false")
	}

	if fixtures[2].GW != 2 {
		t.Errorf("expected GW=2, got %d", fixtures[2].GW)
	}
	if fixtures[0].GameID != "wcf" {
		t.Errorf("expected game_id=wcf")
	}
}

func TestMapManagers(t *testing.T) {
	entries := []wcfRankEntry{
		{UserID: 101, UserName: "Top WCF Team", OverallRank: 1},
		{UserID: 202, UserName: "Second WCF Team", OverallRank: 2},
	}
	managers := mapManagers(entries)
	if len(managers) != 2 {
		t.Fatalf("expected 2, got %d", len(managers))
	}
	if managers[0].ExternalID != 101 {
		t.Errorf("expected ExternalID=101, got %d", managers[0].ExternalID)
	}
	if managers[1].OverallRank != 2 {
		t.Errorf("expected OverallRank=2, got %d", managers[1].OverallRank)
	}
	if managers[0].GameID != "wcf" {
		t.Errorf("expected GameID=wcf")
	}
}

func TestExtractRanks_sortsAscending(t *testing.T) {
	entries := []wcfRankEntry{
		{UserID: 3, OverallRank: 3},
		{UserID: 1, OverallRank: 1},
		{UserID: 2, OverallRank: 2},
	}
	sorted := extractRanks(entries)
	for i, e := range sorted {
		if e.OverallRank != i+1 {
			t.Errorf("position %d: expected rank %d, got %d", i, i+1, e.OverallRank)
		}
	}
}

func TestMapPicks_normalCaptain(t *testing.T) {
	captainID := 10
	viceID := 20
	entry := &wcfPickEntry{
		Captain: &captainID,
		Vice:    &viceID,
		Lineup: map[string][]int{
			"GK":  {5},
			"DEF": {20, 21, 22, 23},
			"MID": {10, 11, 12},
			"FWD": {30, 31},
		},
		BenchOrder: []int{40, 41, 42, 43},
	}
	picks := mapPicks(999, 1, entry)

	// 1 GK + 4 DEF + 3 MID + 2 FWD = 10 lineup + 4 bench = 14
	if len(picks) != 14 {
		t.Fatalf("expected 14 picks, got %d", len(picks))
	}

	// First pick is GK
	if picks[0].PlayerID != "5" {
		t.Errorf("first pick should be GK (5), got %s", picks[0].PlayerID)
	}

	// Captain (id=10) is in MID slot, multiplier=2
	var captainPick *fantasy.ManagerPick
	for i := range picks {
		if picks[i].PlayerID == "10" {
			captainPick = &picks[i]
			break
		}
	}
	if captainPick == nil {
		t.Fatal("captain pick not found")
	}
	if !captainPick.IsCaptain || captainPick.Multiplier != 2 {
		t.Errorf("captain: IsCaptain=%v Multiplier=%d", captainPick.IsCaptain, captainPick.Multiplier)
	}

	// Vice (id=20) is in DEF slot
	var vicePick *fantasy.ManagerPick
	for i := range picks {
		if picks[i].PlayerID == "20" {
			vicePick = &picks[i]
			break
		}
	}
	if vicePick == nil || !vicePick.IsViceCaptain {
		t.Errorf("vice captain not found or flag not set")
	}

	// Bench starts at index 10 in order: 40, 41, 42, 43, multiplier=0
	if picks[10].PlayerID != "40" || picks[10].Multiplier != 0 {
		t.Errorf("bench[0] should be player 40 with mult=0, got %s mult=%d", picks[10].PlayerID, picks[10].Multiplier)
	}

	if picks[0].ManagerID != "999" {
		t.Errorf("expected ManagerID=999, got %s", picks[0].ManagerID)
	}
	if picks[0].GameID != "wcf" {
		t.Errorf("expected GameID=wcf")
	}
}

func TestMapPicks_tripleCaptain(t *testing.T) {
	captainID := 10
	entry := &wcfPickEntry{
		Captain:           &captainID,
		MaxCaptainBooster: true,
		Lineup:            map[string][]int{"GK": {10}, "DEF": {}, "MID": {}, "FWD": {}},
		BenchOrder:        []int{},
	}
	picks := mapPicks(1, 1, entry)
	if len(picks) == 0 {
		t.Fatal("no picks")
	}
	if picks[0].Multiplier != 3 {
		t.Errorf("triple captain: expected multiplier 3, got %d", picks[0].Multiplier)
	}
}

func TestMapPicks_nil(t *testing.T) {
	picks := mapPicks(1, 1, nil)
	if picks != nil {
		t.Errorf("expected nil picks for nil entry")
	}
}

func TestIsActiveChip(t *testing.T) {
	if isActiveChip(nil) {
		t.Error("nil should be inactive")
	}
	if isActiveChip(false) {
		t.Error("false should be inactive")
	}
	if !isActiveChip(true) {
		t.Error("true should be active")
	}
	if !isActiveChip(map[string]interface{}{"id": 1}) {
		t.Error("non-nil non-bool should be active")
	}
}

func TestPlayerName(t *testing.T) {
	known := "Neymar"
	empty := ""

	cases := []struct {
		p    wcfPlayer
		want string
	}{
		{wcfPlayer{FirstName: "Kylian", LastName: "Mbappé", KnownName: &known}, "Neymar"},
		{wcfPlayer{FirstName: "Kylian", LastName: "Mbappé", KnownName: nil}, "Kylian Mbappé"},
		{wcfPlayer{FirstName: "Kylian", LastName: "Mbappé", KnownName: &empty}, "Kylian Mbappé"},
	}
	for _, c := range cases {
		got := playerName(c.p)
		if got != c.want {
			t.Errorf("playerName: expected %q, got %q", c.want, got)
		}
	}
}

func TestMapTeams_updatedAt(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	teams := mapTeams([]wcfSquad{{ID: 1, Name: "X", Abbr: "X"}})
	if teams[0].UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt not set correctly")
	}
}
