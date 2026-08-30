package setpiece

import (
	"context"
	"testing"
	"time"
)

// TestParseShots_SkipsOwnGoals_Live hits real Understat and verifies that no
// set-piece target-man event survives with an OwnGoal result. Gated behind
// -short (skips) so it stays out of the fast unit path; run it explicitly:
//
//	go test ./internal/setpiece/ -run Live -v
func TestParseShots_SkipsOwnGoals_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Understat test in -short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client := NewClient(0, nil)

	// Current season (starting year). Today is the 2026/27 season.
	ids, err := client.FinishedMatches(ctx, "2026")
	if err != nil {
		t.Fatalf("FinishedMatches: %v", err)
	}
	if len(ids) == 0 {
		t.Skip("no finished matches yet for season 2026")
	}

	var (
		rawTargetOwnGoals int // set-piece OwnGoal shots on the target role, in the raw feed
		keptOwnGoals      int // any that survived ParseShots (must be 0)
		sampleTeam        string
		samplePlayer      string
	)

	for _, id := range ids {
		shots, err := client.MatchShots(ctx, id)
		if err != nil {
			t.Logf("skip match %s: %v", id, err)
			continue
		}

		// Expected target-man events = raw target shots minus own goals.
		var wantTargets int
		for _, s := range shots {
			role, _, ok := classify(s.Situation)
			if !ok || role != RoleTarget {
				continue
			}
			if s.Result == "OwnGoal" {
				rawTargetOwnGoals++
				sampleTeam = s.HTeam
				if s.HA == "a" {
					sampleTeam = s.ATeam
				}
				samplePlayer = s.Player
				continue // must be dropped
			}
			wantTargets++
		}

		var gotTargets int
		for _, e := range ParseShots(id, shots) {
			if e.Role == RoleTarget {
				gotTargets++
			}
		}
		if gotTargets > wantTargets {
			keptOwnGoals += gotTargets - wantTargets
		}
	}

	t.Logf("season 2026: %d finished matches, %d raw target-man own goals (e.g. %q for %s)",
		len(ids), rawTargetOwnGoals, samplePlayer, sampleTeam)

	if keptOwnGoals != 0 {
		t.Fatalf("expected 0 target-man own goals to survive ParseShots, got %d", keptOwnGoals)
	}
}
