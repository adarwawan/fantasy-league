package shotzone

import (
	"testing"
	"time"
)

func day(n int) time.Time { return time.Date(2026, 8, n, 0, 0, 0, 0, time.UTC) }

func TestTeamWindowsCountsBothSidesAndBreaksDateTies(t *testing.T) {
	// A plays m1..m4; m3 and m4 share a date. Window 2 must be exactly m3+m4
	// (never 3 matches), and a match where A only conceded still counts.
	ev := []Event{
		{MatchID: "1", MatchDate: day(1), TeamFor: "A", TeamAgainst: "B"},
		{MatchID: "2", MatchDate: day(2), TeamFor: "B", TeamAgainst: "A"}, // A only conceded
		{MatchID: "3", MatchDate: day(5), TeamFor: "A", TeamAgainst: "C"},
		{MatchID: "4", MatchDate: day(5), TeamFor: "A", TeamAgainst: "D"},
	}
	w := TeamWindows(ev, 2)["A"]
	if len(w) != 2 || !w["3"] || !w["4"] {
		t.Fatalf("window = %v, want {3,4}", w)
	}
	// Window 3 reaches back to m2 even though A took no shot in it.
	if w3 := TeamWindows(ev, 3)["A"]; len(w3) != 3 || !w3["2"] {
		t.Fatalf("window3 = %v, want {2,3,4}", w3)
	}
	// Same-day tie at the boundary: window 1 keeps exactly one of m3/m4.
	if w1 := TeamWindows(ev, 1)["A"]; len(w1) != 1 || !w1["4"] {
		t.Fatalf("window1 = %v, want {4}", w1)
	}
	if TeamWindows(ev, 0) != nil {
		t.Fatal("window 0 must disable windowing")
	}
}
