package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"fantasy-league/internal/fantasy"
	"fantasy-league/internal/handler"
	"fantasy-league/internal/store"
)

// Integration test: requires DATABASE_URL to be set.
// Run with: DATABASE_URL=... go test ./internal/handler/...
func TestPlayersEndpoint(t *testing.T) {
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

	// Seed data
	gameID := "test_e2e"
	now := time.Now().UTC()

	teams := []fantasy.Team{
		{GameID: gameID, ExternalID: 1, Name: "Team A", ShortName: "TMA", UpdatedAt: now},
		{GameID: gameID, ExternalID: 2, Name: "Team B", ShortName: "TMB", UpdatedAt: now},
	}
	if err := pg.UpsertTeams(ctx, teams); err != nil {
		t.Fatalf("seed teams: %v", err)
	}

	players := []fantasy.Player{
		{GameID: gameID, ExternalID: 101, Name: "Alpha", TeamID: "1", Position: "MID", Price: 10.5, Form: 8.0, GlobalOwnership: 50.0, Status: "available", UpdatedAt: now},
		{GameID: gameID, ExternalID: 102, Name: "Beta", TeamID: "2", Position: "FWD", Price: 9.0, Form: 5.0, GlobalOwnership: 30.0, Status: "available", UpdatedAt: now},
	}
	if err := pg.UpsertPlayers(ctx, players); err != nil {
		t.Fatalf("seed players: %v", err)
	}

	// Finished GWs 1-5 plus an upcoming easy GW 6 fixture, so Alpha can qualify as must-have.
	fixtures := make([]fantasy.Fixture, 0, 6)
	for gw := 1; gw <= 5; gw++ {
		score := 1
		fixtures = append(fixtures, fantasy.Fixture{
			GameID: gameID, ExternalID: gw, GW: gw,
			HomeTeamID: "1", AwayTeamID: "2",
			HomeDifficulty: 3, AwayDifficulty: 3,
			KickoffTime: now.AddDate(0, 0, -7*(6-gw)), Finished: true,
			HomeScore: &score, AwayScore: &score,
		})
	}
	fixtures = append(fixtures, fantasy.Fixture{
		GameID: gameID, ExternalID: 6, GW: 6,
		HomeTeamID: "1", AwayTeamID: "2",
		HomeDifficulty: 2, AwayDifficulty: 2,
		KickoffTime: now.AddDate(0, 0, 7), Finished: false,
	})
	if err := pg.UpsertFixtures(ctx, fixtures); err != nil {
		t.Fatalf("seed fixtures: %v", err)
	}

	// Alpha: 6+ points in 3 of last 5 GWs (must-have). Beta: only 1 hit (not must-have).
	var stats []fantasy.PlayerGWStat
	for gw, pts := range map[int]int{1: 8, 2: 2, 3: 7, 4: 6, 5: 3} {
		stats = append(stats, fantasy.PlayerGWStat{GameID: gameID, PlayerExternalID: 101, GW: gw, Minutes: 90, Points: pts})
	}
	for gw, pts := range map[int]int{1: 2, 2: 2, 3: 9, 4: 1, 5: 0} {
		stats = append(stats, fantasy.PlayerGWStat{GameID: gameID, PlayerExternalID: 102, GW: gw, Minutes: 90, Points: pts})
	}
	if err := pg.UpsertPlayerGWStats(ctx, stats); err != nil {
		t.Fatalf("seed gw stats: %v", err)
	}

	// Build handler with no-op cache and default must-have thresholds
	cache := &noopCache{}
	h := handler.NewPlayersHandler(pg, cache, nil)

	r := chi.NewRouter()
	r.Get("/api/{game}/players", h.List)
	r.Get("/api/{game}/players/scatter", h.Scatter)

	// Test /api/{game}/players
	req := httptest.NewRequest(http.MethodGet, "/api/"+gameID+"/players", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Players []struct {
			Name            string  `json:"name"`
			GlobalOwnership float64 `json:"global_ownership"`
			MustHave        bool    `json:"must_have"`
		} `json:"players"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Meta.Total != 2 {
		t.Errorf("expected 2 players, got %d", resp.Meta.Total)
	}
	// Default sort is global_ownership DESC — Alpha (50) should be first
	if len(resp.Players) > 0 && resp.Players[0].Name != "Alpha" {
		t.Errorf("expected Alpha first (highest ownership), got %s", resp.Players[0].Name)
	}
	// Alpha: top-owned MID, 3/5 GWs >= 6 pts, next FDR 2, available → must-have.
	// Beta: only 1/5 GWs >= 6 pts → not must-have.
	for _, p := range resp.Players {
		switch p.Name {
		case "Alpha":
			if !p.MustHave {
				t.Errorf("expected Alpha to be must-have")
			}
		case "Beta":
			if p.MustHave {
				t.Errorf("expected Beta to not be must-have")
			}
		}
	}

	// Cleanup seeded data
	pg.DeleteTestGame(ctx, gameID)
}

type noopCache struct{}

func (n *noopCache) Get(_ context.Context, _ string) ([]byte, error)                       { return nil, nil }
func (n *noopCache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error      { return nil }
