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

	// Build handler with no-op cache
	cache := &noopCache{}
	h := handler.NewPlayersHandler(pg, cache)

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

	// Cleanup seeded data
	pg.DeleteTestGame(ctx, gameID)
}

type noopCache struct{}

func (n *noopCache) Get(_ context.Context, _ string) ([]byte, error)                       { return nil, nil }
func (n *noopCache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error      { return nil }
