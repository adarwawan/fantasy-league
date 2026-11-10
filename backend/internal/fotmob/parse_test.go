package fotmob

import (
	"os"
	"testing"

	"fantasy-league/internal/setpiece"
	"fantasy-league/internal/shotzone"
)

func TestParseRealMatch(t *testing.T) {
	b, err := os.ReadFile("testdata/match.json") // Man Utd 4-4 Aston Villa, 2026-11-08
	if err != nil {
		t.Fatal(err)
	}
	m, err := decodeMatch(b)
	if err != nil {
		t.Fatal(err)
	}
	sp, sz := Parse("5795512", "2026", m)

	// 38 shots: 26 open play (24 RegularPlay + 2 FastBreak), 7 SetPiece,
	// 3 FromCorner, 2 Penalty.
	if len(sz) != 26 || len(sp) != 12 {
		t.Fatalf("sz=%d sp=%d, want 26 and 12", len(sz), len(sp))
	}
	duties := map[setpiece.Duty]int{}
	for _, e := range sp {
		duties[e.Duty]++
		if e.ShotID == "" || e.PlayerID == "" || e.UnderstatTeam == "" {
			t.Errorf("missing ids: %+v", e)
		}
	}
	if duties[setpiece.DutyPenalty] != 2 || duties[setpiece.DutyCorner] != 3 || duties[setpiece.DutySetPiece] != 7 {
		t.Errorf("duties = %v", duties)
	}
	teams := map[string]bool{}
	for _, e := range sz {
		teams[e.TeamFor] = true
		if e.TeamFor == e.TeamAgainst {
			t.Errorf("team for == against: %+v", e)
		}
	}
	if !teams["Man Utd"] || !teams["Aston Villa"] || len(teams) != 2 {
		t.Errorf("teams = %v", teams)
	}
}

func TestParseRules(t *testing.T) {
	home, away := team{ID: 1, Name: "Tottenham Hotspur"}, team{ID: 2, Name: "Everton"}
	m := matchData{Home: home, Away: away, Shots: []shot{
		{ID: 1, TeamID: 1, PlayerID: 10, Situation: "FreeKick", EventType: "Goal", XG: 0.1},
		{ID: 2, TeamID: 2, PlayerID: 20, Situation: "FromCorner", ShotType: "Header"},
		{ID: 3, TeamID: 1, PlayerID: 11, Situation: "RegularPlay", IsOwnGoal: true},
		{ID: 4, TeamID: 1, PlayerID: 11, Situation: "ThrowInSetPiece", X: 100, Y: 34},
		{ID: 5, TeamID: 1, PlayerID: 11, Situation: "Mystery"},
		// Same player, same minute, same zone: v1's key would merge these.
		{ID: 6, TeamID: 2, PlayerID: 21, Situation: "RegularPlay", Min: 5, X: 100, Y: 34},
		{ID: 7, TeamID: 2, PlayerID: 21, Situation: "RegularPlay", Min: 5, X: 100, Y: 34},
	}}
	sp, sz := Parse("m", "2026", m)

	if len(sp) != 2 || sp[0].Role != setpiece.RoleTaker || sp[0].Duty != setpiece.DutyDFK || !sp[0].IsGoal ||
		sp[0].UnderstatTeam != "Spurs" {
		t.Errorf("free kick: %+v", sp)
	}
	if sp[1].Duty != setpiece.DutyCorner || !sp[1].IsHeader || sp[1].UnderstatTeam != "Everton" {
		t.Errorf("corner: %+v", sp[1])
	}
	// own goal and unknown dropped; throw-in + two rebounds are open play.
	if len(sz) != 3 {
		t.Fatalf("sz = %d, want 3", len(sz))
	}
	if sz[0].Zone != shotzone.ZoneCenter || sz[0].TeamAgainst != "Everton" {
		t.Errorf("throw-in shot: %+v", sz[0])
	}
	if sz[1].ShotID == sz[2].ShotID {
		t.Error("same-minute shots must keep distinct shot ids")
	}
}

func TestZoneBoundaries(t *testing.T) {
	// 105x68 metres -> shared 0-1 zone spec (box line 0.83 ~ 87m).
	cases := []struct {
		x, y float64
		want shotzone.Zone
	}{
		{80, 34, shotzone.ZoneOutside},
		{95, 10, shotzone.ZoneLeft},
		{95, 34, shotzone.ZoneCenter},
		{95, 60, shotzone.ZoneRight},
	}
	for _, c := range cases {
		if got := shotzone.ClassifyZone(c.x/pitchLength, c.y/pitchWidth); got != c.want {
			t.Errorf("(%v,%v) = %s, want %s", c.x, c.y, got, c.want)
		}
	}
}

func TestDecodeMatchNoShotmap(t *testing.T) {
	if _, err := decodeMatch([]byte(`{"error":true,"message":"Data not found"}`)); err == nil {
		t.Fatal("want error so the match stays unprocessed")
	}
}
