package setpiece

import (
	"testing"
	"time"
)

func d(n int) time.Time { return time.Date(2026, 8, n, 0, 0, 0, 0, time.UTC) }

func countFor(board []BoardRow, team, player string) int {
	for _, r := range board {
		if r.UnderstatTeam == team && r.PlayerID == player && r.Role == RoleTaker {
			return r.RawCount
		}
	}
	return 0
}

func TestRealMatchWindowIncludesQuietMatches(t *testing.T) {
	// Team A's last 2 real matches are m3 and m4, but it had no set-piece shot in
	// m4. Events-derived window = {m2, m3} (wrong: reaches back to m2); the real
	// match list = {m3, m4}, so m2's penalty must fall out.
	ev := []Event{
		{MatchID: "2", MatchDate: d(2), UnderstatTeam: "A", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "p2"},
		{MatchID: "3", MatchDate: d(3), UnderstatTeam: "A", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "p3"},
	}
	derived := Aggregate(ev, AggregateConfig{WindowMatches: 2})
	if countFor(derived, "A", "p2") != 1 {
		t.Fatal("precondition: events-derived window should still include m2")
	}
	real := Aggregate(ev, AggregateConfig{WindowMatches: 2, TeamMatches: map[string]map[string]bool{
		"A": {"3": true, "4": true},
	}})
	if countFor(real, "A", "p2") != 0 || countFor(real, "A", "p3") != 1 {
		t.Fatalf("real window board = %+v", real)
	}
}

func TestEventWindowSameDayTieKeepsExactlyN(t *testing.T) {
	ev := []Event{
		{MatchID: "1", MatchDate: d(1), UnderstatTeam: "A", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "p1"},
		{MatchID: "2", MatchDate: d(5), UnderstatTeam: "A", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "p2"},
		{MatchID: "3", MatchDate: d(5), UnderstatTeam: "A", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "p3"},
	}
	// Window 1 over two same-day matches: exactly one match survives, not both.
	if got := len(trimToWindow(ev, 1, nil)); got != 1 {
		t.Fatalf("kept %d events, want 1", got)
	}
}
