package setpiece

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func loadFixtureShots(t *testing.T) []shot {
	t.Helper()
	body, err := os.ReadFile("testdata/match.json")
	if err != nil {
		t.Fatal(err)
	}
	shots, err := decodeShots(body)
	if err != nil {
		t.Fatalf("decodeShots: %v", err)
	}
	return shots
}

func TestParseShots_ClassifiesBothRoles(t *testing.T) {
	events := ParseShots("m1", loadFixtureShots(t))

	// 6 shots in fixture, 1 is OpenPlay -> 5 events.
	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}

	type key struct {
		role Role
		duty Duty
		pid  string
	}
	got := map[key]Event{}
	for _, e := range events {
		got[key{e.Role, e.Duty, e.PlayerID}] = e
	}

	// Penalty taker.
	if e, ok := got[key{RoleTaker, DutyPenalty, "1228"}]; !ok {
		t.Error("missing penalty taker event")
	} else if e.XG != 0.76 || e.UnderstatTeam != "Man Utd" {
		t.Errorf("penalty event wrong: %+v", e)
	}

	// Direct free-kick taker (away team -> Everton).
	if e, ok := got[key{RoleTaker, DutyDFK, "5000"}]; !ok {
		t.Error("missing DFK taker event")
	} else if e.UnderstatTeam != "Everton" {
		t.Errorf("DFK team: got %q want Everton", e.UnderstatTeam)
	}

	// Corner target man (header).
	if e, ok := got[key{RoleTarget, DutyCorner, "1668"}]; !ok {
		t.Error("missing corner target event")
	} else if !e.IsHeader {
		t.Error("corner target should be a header")
	}

	// Set-piece target man.
	if _, ok := got[key{RoleTarget, DutySetPiece, "1668"}]; !ok {
		t.Error("missing setpiece target event")
	}

	// Away corner target (not a header).
	if e, ok := got[key{RoleTarget, DutyCorner, "5001"}]; !ok {
		t.Error("missing away corner target event")
	} else if e.IsHeader {
		t.Error("away corner target should not be a header")
	}
}

func TestParseShots_SkipsTargetOwnGoals(t *testing.T) {
	shots := []shot{
		// Corner deflected in as an own goal, credited to a keeper — must be dropped.
		{Situation: situationFromCorner, Result: "OwnGoal", PlayerID: "99", Player: "Donnarumma", HTeam: "Man City", HA: "h", ShotType: "Head"},
		// Genuine corner target header — must be kept.
		{Situation: situationFromCorner, Result: "SavedShot", PlayerID: "1668", Player: "Target", HTeam: "Man City", HA: "h", ShotType: "Head"},
	}
	events := ParseShots("m1", shots)
	if len(events) != 1 {
		t.Fatalf("expected 1 event after skipping own goal, got %d", len(events))
	}
	if events[0].PlayerID != "1668" {
		t.Errorf("kept wrong event: %+v", events[0])
	}
}

func TestMatchShots_FetchesAndDecodes(t *testing.T) {
	fixture, _ := os.ReadFile("testdata/match.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Requested-With") != xRequestedWith {
			t.Errorf("missing X-Requested-With header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	client := NewClientWithBase(srv.URL, time.Minute, nil)
	shots, err := client.MatchShots(context.Background(), "m1")
	if err != nil {
		t.Fatalf("MatchShots: %v", err)
	}
	if len(shots) != 6 {
		t.Fatalf("expected 6 shots, got %d", len(shots))
	}
}
