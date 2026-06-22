package odds_test

import (
	"testing"

	"fantasy-league/internal/fantasy"
	"fantasy-league/internal/sources/odds"
)

// helpers

func team(id, gameID, name, short string) fantasy.Team {
	return fantasy.Team{ID: id, GameID: gameID, Name: name, ShortName: short}
}

func matchOdds(id, home, away, homeID, awayID string) odds.MatchOdds {
	return odds.MatchOdds{
		OddsMatchID: id,
		HomeTeam:    home,
		AwayTeam:    away,
		HomeTeamID:  homeID,
		AwayTeamID:  awayID,
		LambdaHome:  2.0,
		LambdaAway:  1.0,
		HomeCSPct:   36.8,
		AwayCSPct:   13.5,
	}
}

func fixture(id, homeID, awayID string) fantasy.Fixture {
	return fantasy.Fixture{ID: id, HomeTeamID: homeID, AwayTeamID: awayID}
}

// --- MapTeams ---

func TestMapTeams_ExactMatchViaOverride(t *testing.T) {
	// "Portugal" is in nameOverrides and maps to "Portugal"; both teams resolve.
	teams := []fantasy.Team{
		team("t1a", "wcf", "Portugal", "POR"),
		team("t1b", "wcf", "Brazil", "BRA"),
	}
	input := []odds.MatchOdds{{OddsMatchID: "m1", HomeTeam: "Portugal", AwayTeam: "Brazil"}}
	result := odds.MapTeams(input, teams, "wcf")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].HomeTeamID != "t1a" {
		t.Errorf("HomeTeamID: got %q, want %q", result[0].HomeTeamID, "t1a")
	}
	if result[0].AwayTeamID != "t1b" {
		t.Errorf("AwayTeamID: got %q, want %q", result[0].AwayTeamID, "t1b")
	}
}

func TestMapTeams_OverrideTranslatesName(t *testing.T) {
	// "DR Congo" → override → "Congo DR" → exact hit.
	teams := []fantasy.Team{
		team("t2a", "wcf", "Congo DR", "CGO"),
		team("t2b", "wcf", "Brazil", "BRA"),
	}
	input := []odds.MatchOdds{{OddsMatchID: "m2", HomeTeam: "DR Congo", AwayTeam: "Brazil"}}
	result := odds.MapTeams(input, teams, "wcf")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].HomeTeamID != "t2a" {
		t.Errorf("override: HomeTeamID got %q, want %q", result[0].HomeTeamID, "t2a")
	}
}

func TestMapTeams_OverrideTranslatesIvoryCoast(t *testing.T) {
	// "Ivory Coast" → override → "Côte d'Ivoire" → exact hit.
	teams := []fantasy.Team{
		team("t3a", "wcf", "Côte d'Ivoire", "CIV"),
		team("t3b", "wcf", "Brazil", "BRA"),
	}
	input := []odds.MatchOdds{{OddsMatchID: "m3", HomeTeam: "Ivory Coast", AwayTeam: "Brazil"}}
	result := odds.MapTeams(input, teams, "wcf")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].HomeTeamID != "t3a" {
		t.Errorf("override→exact: HomeTeamID got %q, want %q", result[0].HomeTeamID, "t3a")
	}
}

func TestMapTeams_UnknownNameSkipped(t *testing.T) {
	teams := []fantasy.Team{team("t4", "wcf", "Brazil", "BRA")}
	input := []odds.MatchOdds{{OddsMatchID: "m4", HomeTeam: "Zzzland", AwayTeam: "Brazil"}}
	result := odds.MapTeams(input, teams, "wcf")
	if len(result) != 0 {
		t.Errorf("unknown name: expected record skipped, got %d records", len(result))
	}
}

func TestMapTeams_WrongGameIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for wrong gameID")
		}
	}()
	teams := []fantasy.Team{team("t5", "fpl", "Arsenal", "ARS")}
	odds.MapTeams(nil, teams, "wcf") // fpl team in wcf call → panic
}

func TestMapTeams_UnresolvedRecordSkipped(t *testing.T) {
	teams := []fantasy.Team{team("t6", "wcf", "Brazil", "BRA")}
	input := []odds.MatchOdds{
		{OddsMatchID: "m5a", HomeTeam: "Brazil", AwayTeam: "Zzzland"},
		{OddsMatchID: "m5b", HomeTeam: "Brazil", AwayTeam: "Brazil"},
	}
	result := odds.MapTeams(input, teams, "wcf")
	if len(result) != 1 {
		t.Errorf("expected 1 resolved record, got %d", len(result))
	}
	if result[0].OddsMatchID != "m5b" {
		t.Errorf("expected m5b to survive, got %q", result[0].OddsMatchID)
	}
}

// --- LinkFixtures ---

func TestLinkFixtures_WCF_SidesAgree(t *testing.T) {
	m := matchOdds("o1", "A", "B", "team-a", "team-b")
	f := fixture("fix-1", "team-a", "team-b")
	cfg := odds.WCFOddsConfig

	result := odds.LinkFixtures([]odds.MatchOdds{m}, []fantasy.Fixture{f}, cfg)
	if result[0].FixtureID != "fix-1" {
		t.Errorf("FixtureID: got %q, want %q", result[0].FixtureID, "fix-1")
	}
	// No swap — lambdas unchanged.
	if result[0].LambdaHome != 2.0 || result[0].LambdaAway != 1.0 {
		t.Errorf("lambdas should be unchanged; got home=%.1f away=%.1f", result[0].LambdaHome, result[0].LambdaAway)
	}
}

func TestLinkFixtures_WCF_SidesFlipped_SwapApplied(t *testing.T) {
	// Odds list B as home, A as away; fixture says A is home.
	m := matchOdds("o2", "B", "A", "team-b", "team-a")
	m.LambdaHome, m.LambdaAway = 2.0, 1.0
	m.HomeCSPct, m.AwayCSPct = 36.0, 13.0
	f := fixture("fix-2", "team-a", "team-b") // A is home in fantasy
	cfg := odds.WCFOddsConfig

	result := odds.LinkFixtures([]odds.MatchOdds{m}, []fantasy.Fixture{f}, cfg)
	if result[0].FixtureID != "fix-2" {
		t.Errorf("FixtureID: got %q", result[0].FixtureID)
	}
	// After swap: fixture home (A=team-a=odds away) gets odds-away values.
	if result[0].LambdaHome != 1.0 || result[0].LambdaAway != 2.0 {
		t.Errorf("swap not applied: home=%.1f away=%.1f", result[0].LambdaHome, result[0].LambdaAway)
	}
	if result[0].HomeCSPct != 13.0 || result[0].AwayCSPct != 36.0 {
		t.Errorf("CS%% swap not applied: home=%.1f away=%.1f", result[0].HomeCSPct, result[0].AwayCSPct)
	}
}

func TestLinkFixtures_FPL_SidesAgree(t *testing.T) {
	m := matchOdds("o3", "C", "D", "team-c", "team-d")
	f := fixture("fix-3", "team-c", "team-d")
	cfg := odds.FPLOddsConfig

	result := odds.LinkFixtures([]odds.MatchOdds{m}, []fantasy.Fixture{f}, cfg)
	if result[0].FixtureID != "fix-3" {
		t.Errorf("FixtureID: got %q, want %q", result[0].FixtureID, "fix-3")
	}
}

func TestLinkFixtures_FPL_SidesFlipped_LeftUnlinked(t *testing.T) {
	// Strict mode: odds say C is home but fixture says D is home → no link.
	m := matchOdds("o4", "C", "D", "team-c", "team-d")
	f := fixture("fix-4", "team-d", "team-c") // D is home in fixture
	cfg := odds.FPLOddsConfig

	result := odds.LinkFixtures([]odds.MatchOdds{m}, []fantasy.Fixture{f}, cfg)
	if result[0].FixtureID != "" {
		t.Errorf("strict mismatch should leave FixtureID empty, got %q", result[0].FixtureID)
	}
	// No side swap.
	if result[0].LambdaHome != 2.0 {
		t.Errorf("lambdas must not be swapped in strict mode")
	}
}

func TestLinkFixtures_NoFixtureMatch_FixtureIDEmpty(t *testing.T) {
	m := matchOdds("o5", "E", "F", "team-e", "team-f")
	f := fixture("fix-5", "team-x", "team-y") // different teams entirely
	cfg := odds.WCFOddsConfig

	result := odds.LinkFixtures([]odds.MatchOdds{m}, []fantasy.Fixture{f}, cfg)
	if result[0].FixtureID != "" {
		t.Errorf("unmatched: FixtureID should be empty, got %q", result[0].FixtureID)
	}
}
