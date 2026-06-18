package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"fantasy-league/internal/fantasy"
	"fantasy-league/internal/store"
)

// Integration test: requires DATABASE_URL and migration 002 to be applied.
// Run with: DATABASE_URL=... go test ./internal/store/...
func TestRecomputeTeamForm(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pg, err := store.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pg.Close()

	gameID := "test_teamform"
	now := time.Now().UTC()

	teams := []fantasy.Team{
		{GameID: gameID, ExternalID: 1, Name: "Team A", ShortName: "TMA", UpdatedAt: now},
		{GameID: gameID, ExternalID: 2, Name: "Team B", ShortName: "TMB", UpdatedAt: now},
	}
	if err := pg.UpsertTeams(ctx, teams); err != nil {
		t.Fatalf("UpsertTeams: %v", err)
	}
	t.Cleanup(func() { pg.DeleteTestGame(ctx, gameID) })

	home2, away0 := 2, 0
	home1, away1 := 1, 1
	home3, away1b := 3, 1

	// GW 1: Team A (home) 2-0 Team B  → A: W(3pts), B: L(0pts)
	// GW 2: Team B (home) 1-1 Team A  → B: D(1pt),  A: D(1pt)
	// GW 3: Team A (home) 3-1 Team B  → A: W(3pts), B: L(0pts)
	//
	// Team A: scored 2+1+3=6, conceded 0+1+1=2, points 3+1+3=7  → avg over 3 games: 2.00 / 0.67 / 2.33
	// Team B: scored 0+1+1=2, conceded 2+1+3=6, points 0+1+0=1  → avg over 3 games: 0.67 / 2.00 / 0.33
	fixtures := []fantasy.Fixture{
		{
			GameID: gameID, ExternalID: 1, GW: 1,
			HomeTeamID: "1", AwayTeamID: "2",
			HomeDifficulty: 3, AwayDifficulty: 3,
			KickoffTime: now.Add(-48 * time.Hour),
			Finished:    true,
			HomeScore:   &home2, AwayScore: &away0,
		},
		{
			GameID: gameID, ExternalID: 2, GW: 2,
			HomeTeamID: "2", AwayTeamID: "1",
			HomeDifficulty: 3, AwayDifficulty: 3,
			KickoffTime: now.Add(-24 * time.Hour),
			Finished:    true,
			HomeScore:   &home1, AwayScore: &away1,
		},
		{
			GameID: gameID, ExternalID: 3, GW: 3,
			HomeTeamID: "1", AwayTeamID: "2",
			HomeDifficulty: 3, AwayDifficulty: 3,
			KickoffTime: now.Add(-1 * time.Hour),
			Finished:    true,
			HomeScore:   &home3, AwayScore: &away1b,
		},
	}
	if err := pg.UpsertFixtures(ctx, fixtures); err != nil {
		t.Fatalf("UpsertFixtures: %v", err)
	}

	if err := pg.RecomputeTeamForm(ctx, gameID, 5); err != nil {
		t.Fatalf("RecomputeTeamForm: %v", err)
	}

	rows, err := pg.QueryTeams(ctx, gameID)
	if err != nil {
		t.Fatalf("QueryTeams: %v", err)
	}

	teamByName := map[string]store.TeamRow{}
	for _, row := range rows {
		teamByName[row.Name] = row
	}

	teamA, ok := teamByName["Team A"]
	if !ok {
		t.Fatal("Team A not found in results")
	}
	if teamA.AttForm < 1.99 || teamA.AttForm > 2.01 {
		t.Errorf("Team A att_form: expected ~2.00, got %f", teamA.AttForm)
	}
	if teamA.DefForm < 0.66 || teamA.DefForm > 0.68 {
		t.Errorf("Team A def_form: expected ~0.67, got %f", teamA.DefForm)
	}
	if teamA.OvrForm < 2.32 || teamA.OvrForm > 2.34 {
		t.Errorf("Team A ovr_form: expected ~2.33, got %f", teamA.OvrForm)
	}

	teamB, ok := teamByName["Team B"]
	if !ok {
		t.Fatal("Team B not found in results")
	}
	if teamB.AttForm < 0.66 || teamB.AttForm > 0.68 {
		t.Errorf("Team B att_form: expected ~0.67, got %f", teamB.AttForm)
	}
	if teamB.DefForm < 1.99 || teamB.DefForm > 2.01 {
		t.Errorf("Team B def_form: expected ~2.00, got %f", teamB.DefForm)
	}
	if teamB.OvrForm < 0.32 || teamB.OvrForm > 0.34 {
		t.Errorf("Team B ovr_form: expected ~0.33, got %f", teamB.OvrForm)
	}
}
