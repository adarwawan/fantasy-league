package setpiece

import (
	"testing"
	"time"
)

func TestAggregate_RanksTakersAndTargets(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2025, 9, d, 15, 0, 0, 0, time.UTC) }

	events := []Event{
		// Man Utd penalties: Bruno x2, Rashford x1 -> Bruno is #1.
		{MatchID: "1", UnderstatTeam: "Man Utd", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "b", PlayerName: "Bruno", MatchDate: day(1), XG: 0.76},
		{MatchID: "2", UnderstatTeam: "Man Utd", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "b", PlayerName: "Bruno", MatchDate: day(8), XG: 0.76},
		{MatchID: "3", UnderstatTeam: "Man Utd", Role: RoleTaker, Duty: DutyPenalty, PlayerID: "r", PlayerName: "Rashford", MatchDate: day(3), XG: 0.76},

		// Man Utd corner target men: Maguire 3 headers, Casemiro 1 -> Maguire #1.
		{MatchID: "1", UnderstatTeam: "Man Utd", Role: RoleTarget, Duty: DutyCorner, PlayerID: "m", PlayerName: "Maguire", MatchDate: day(1), IsHeader: true, XG: 0.1},
		{MatchID: "2", UnderstatTeam: "Man Utd", Role: RoleTarget, Duty: DutyCorner, PlayerID: "m", PlayerName: "Maguire", MatchDate: day(8), IsHeader: true, XG: 0.1},
		{MatchID: "3", UnderstatTeam: "Man Utd", Role: RoleTarget, Duty: DutySetPiece, PlayerID: "m", PlayerName: "Maguire", MatchDate: day(3), IsHeader: true, XG: 0.1},
		{MatchID: "1", UnderstatTeam: "Man Utd", Role: RoleTarget, Duty: DutyCorner, PlayerID: "c", PlayerName: "Casemiro", MatchDate: day(1), IsHeader: false, XG: 0.2},
	}

	rows := Aggregate(events, AggregateConfig{WindowMatches: 6})

	find := func(role Role, duty Duty, pid string) *BoardRow {
		for i := range rows {
			r := &rows[i]
			if r.Role == role && r.Duty == duty && r.PlayerID == pid {
				return r
			}
		}
		return nil
	}

	if r := find(RoleTaker, DutyPenalty, "b"); r == nil || r.Rank != 1 {
		t.Errorf("Bruno should be #1 penalty taker: %+v", r)
	}
	if r := find(RoleTaker, DutyPenalty, "r"); r == nil || r.Rank != 2 {
		t.Errorf("Rashford should be #2 penalty taker: %+v", r)
	}

	// Target 'all' aggregate: Maguire (3 shots across corner+setpiece) is #1.
	mAll := find(RoleTarget, DutyAll, "m")
	if mAll == nil || mAll.Rank != 1 {
		t.Fatalf("Maguire should be #1 target overall: %+v", mAll)
	}
	if mAll.RawCount != 3 {
		t.Errorf("Maguire target shots: got %d want 3", mAll.RawCount)
	}
	if mAll.HeaderPct == nil || *mAll.HeaderPct != 100 {
		t.Errorf("Maguire header_pct: got %v want 100", mAll.HeaderPct)
	}
	if find(RoleTarget, DutyAll, "c") == nil {
		t.Error("Casemiro should have a target 'all' row")
	}
}

func TestRecencyWeight(t *testing.T) {
	newest := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
	// One half-life earlier -> weight 0.5.
	older := newest.Add(-45 * 24 * time.Hour)
	w := recencyWeight(newest, older, 1080*time.Hour)
	if w < 0.49 || w > 0.51 {
		t.Errorf("expected ~0.5, got %f", w)
	}
	// Zero half-life disables weighting.
	if w := recencyWeight(newest, older, 0); w != 1.0 {
		t.Errorf("expected 1.0 with zero half-life, got %f", w)
	}
}
