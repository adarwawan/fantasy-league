package handler

import (
	"testing"

	"fantasy-league/internal/store"
)

func TestComputeTeamICTShares(t *testing.T) {
	lines := []store.PlayerStatGW{
		// ARS: two players. Star sums ICT 60+90=150, Support 50. Team total 200.
		{PlayerID: "p1", Position: "MID", Name: "Star", TeamShortName: "ARS", Influence: 20, Creativity: 20, Threat: 20},
		{PlayerID: "p1", Position: "MID", Name: "Star", TeamShortName: "ARS", Influence: 30, Creativity: 30, Threat: 30},
		{PlayerID: "p2", Position: "FWD", Name: "Support", TeamShortName: "ARS", Influence: 10, Creativity: 10, Threat: 30},
		// ARS: zero-ICT player counts toward neither the list nor the total.
		{PlayerID: "p3", Position: "GK", Name: "Idle", TeamShortName: "ARS"},
		// BOU: single player, decimals that need rounding.
		{PlayerID: "p4", Position: "FWD", Name: "Solo", TeamShortName: "BOU", Influence: 10.15, Threat: 20.25},
	}

	teams := computeTeamICTShares(lines, 5)

	if len(teams) != 2 || teams[0].Team != "ARS" || teams[1].Team != "BOU" {
		t.Fatalf("expected [ARS BOU], got %+v", teams)
	}

	ars := teams[0]
	if ars.TotalICT != 200 {
		t.Errorf("expected ARS total 200, got %v", ars.TotalICT)
	}
	if len(ars.Players) != 2 || ars.Players[0].Name != "Star" || ars.Players[1].Name != "Support" {
		t.Fatalf("unexpected ARS players: %+v", ars.Players)
	}
	if ars.Players[0].ICT != 150 || ars.Players[0].Share != 75.0 {
		t.Errorf("expected Star ict=150 share=75, got %+v", ars.Players[0])
	}
	if ars.Players[0].Influence != 50 || ars.Players[0].Creativity != 50 || ars.Players[0].Threat != 50 {
		t.Errorf("unexpected Star component totals: %+v", ars.Players[0])
	}
	if ars.Players[1].Share != 25.0 {
		t.Errorf("expected Support share=25, got %+v", ars.Players[1])
	}

	bou := teams[1]
	if len(bou.Players) != 1 || bou.Players[0].ICT != 30.4 || bou.Players[0].Share != 100.0 {
		t.Errorf("unexpected BOU players: %+v", bou.Players)
	}

	// Badges: Star out-totals Support in every component → rank 1 across the
	// board, Support rank 2. Solo leads BOU in influence and threat but has
	// zero creativity, which must earn no rank.
	if s := ars.Players[0]; s.InfluenceRank != 1 || s.CreativityRank != 1 || s.ThreatRank != 1 {
		t.Errorf("expected Star rank 1 in all components, got %+v", s)
	}
	if s := ars.Players[1]; s.InfluenceRank != 2 || s.CreativityRank != 2 || s.ThreatRank != 2 {
		t.Errorf("expected Support rank 2 in all components, got %+v", s)
	}
	if s := bou.Players[0]; s.InfluenceRank != 1 || s.ThreatRank != 1 || s.CreativityRank != 0 {
		t.Errorf("expected Solo influence/threat rank 1 and no creativity rank, got %+v", s)
	}
}

func TestComputeTeamICTSharesLimitKeepsFullTeamTotal(t *testing.T) {
	// Six players of 10 ICT each; limit 2. Shares are against the full team
	// total (60), not the emitted subset, so each listed player has ~16.7%.
	var lines []store.PlayerStatGW
	for i := 0; i < 6; i++ {
		lines = append(lines, store.PlayerStatGW{
			PlayerID: string(rune('a' + i)), Position: "MID",
			Name: string(rune('a' + i)), TeamShortName: "LIV", Threat: 10,
		})
	}

	teams := computeTeamICTShares(lines, 2)
	// Limit is 2, but "c" holds threat rank 3 (name tiebreak) and is appended
	// past the cutoff; d/e/f hold no badge and stay excluded.
	if len(teams) != 1 || len(teams[0].Players) != 3 {
		t.Fatalf("expected 1 team with 2+1 players, got %+v", teams)
	}
	if teams[0].TotalICT != 60 {
		t.Errorf("expected team total 60, got %v", teams[0].TotalICT)
	}
	if teams[0].Players[0].Share != 16.7 {
		t.Errorf("expected share 16.7 against full total, got %v", teams[0].Players[0].Share)
	}
	// Equal threat values: ranks 1-3 break ties by name, and only the top 3 of
	// the six earn a rank. Zero influence earns none.
	p := teams[0].Players
	if p[0].ThreatRank != 1 || p[1].ThreatRank != 2 {
		t.Errorf("expected threat ranks 1,2 by name tiebreak, got %+v", p)
	}
	if p[2].Name != "c" || p[2].ThreatRank != 3 || p[2].Share != 16.7 {
		t.Errorf("expected badge-holder c appended past the limit with a share, got %+v", p[2])
	}
	if p[0].InfluenceRank != 0 {
		t.Errorf("expected no influence rank for zero influence, got %+v", p[0])
	}
}
