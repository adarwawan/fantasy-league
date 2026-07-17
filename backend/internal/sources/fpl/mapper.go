package fpl

import (
	"strconv"
	"strings"
	"time"

	"fantasy-league/internal/fantasy"
)

const gameID = "fpl"

var positionMap = map[int]string{
	1: "GK",
	2: "DEF",
	3: "MID",
	4: "FWD",
}

var statusMap = map[string]string{
	"a": "available",
	"d": "doubt",
	"i": "injured",
	"s": "injured",
	"n": "injured",
	"u": "injured",
}

func mapTeams(raw []fplTeam) []fantasy.Team {
	teams := make([]fantasy.Team, 0, len(raw))
	for _, t := range raw {
		teams = append(teams, fantasy.Team{
			GameID:     gameID,
			ExternalID: t.ID,
			Name:       t.Name,
			ShortName:  t.ShortName,
			UpdatedAt:  time.Now().UTC(),
		})
	}
	return teams
}

func mapPlayers(raw []fplPlayer) []fantasy.Player {
	players := make([]fantasy.Player, 0, len(raw))
	for _, p := range raw {
		pos := positionMap[p.ElementType]
		if pos == "" {
			pos = "MID"
		}
		status := statusMap[p.Status]
		if status == "" {
			status = "available"
		}
		form := parseFloat(p.Form)
		globalOwn := parseFloat(p.SelectedByPercent)

		players = append(players, fantasy.Player{
			GameID:               gameID,
			ExternalID:           p.ID,
			Name:                 p.WebName,
			TeamID:               strconv.Itoa(p.Team),
			Position:             pos,
			Price:                float64(p.NowCost) / 10.0,
			Form:                 form,
			GlobalOwnership:      globalOwn,
			Status:               status,
			News:                 p.News,
			PenaltiesOrder:       p.PenaltiesOrder,
			DirectFreekicksOrder: p.DirectFreekicksOrder,
			CornersIndirectOrder: p.CornersIndirectOrder,
			UpdatedAt:            time.Now().UTC(),
		})
	}
	return players
}

func mapFixtures(raw []fplFixture) []fantasy.Fixture {
	fixtures := make([]fantasy.Fixture, 0, len(raw))
	for _, f := range raw {
		if f.Event == nil {
			continue
		}
		kickoff := time.Time{}
		if f.KickoffTime != "" {
			kickoff, _ = time.Parse(time.RFC3339, f.KickoffTime)
		}
		fixtures = append(fixtures, fantasy.Fixture{
			GameID:         gameID,
			ExternalID:     f.ID,
			GW:             *f.Event,
			HomeTeamID:     strconv.Itoa(f.TeamH),
			AwayTeamID:     strconv.Itoa(f.TeamA),
			HomeDifficulty: f.TeamHDifficulty,
			AwayDifficulty: f.TeamADifficulty,
			KickoffTime:    kickoff,
			Finished:       f.Finished,
			HomeScore:      f.TeamHScore,
			AwayScore:      f.TeamAScore,
		})
	}
	return fixtures
}

func mapGWStats(gw int, elements []fplLiveElement) []fantasy.PlayerGWStat {
	stats := make([]fantasy.PlayerGWStat, 0, len(elements))
	for _, e := range elements {
		stats = append(stats, fantasy.PlayerGWStat{
			GameID:           gameID,
			PlayerExternalID: e.ID,
			GW:               gw,
			Minutes:          e.Stats.Minutes,
			Points:           e.Stats.TotalPoints,
			Goals:            e.Stats.GoalsScored,
			Assists:          e.Stats.Assists,
			Bonus:            e.Stats.Bonus,
			CleanSheets:      e.Stats.CleanSheets,
			DefCon:           defConPoints(e.Explain),
			Influence:        parseFloat(e.Stats.Influence),
			Creativity:       parseFloat(e.Stats.Creativity),
			Threat:           parseFloat(e.Stats.Threat),
		})
	}
	return stats
}

// defConPoints sums the defensive-contribution points a player earned across all
// of the gameweek's fixtures. FPL evaluates the CBIT/recovery threshold per
// fixture, so in a double gameweek the correct total is the sum of per-fixture
// points (0, 2 or 4) — not the threshold applied once to the aggregate action
// count, which is what stats.defensive_contribution would give.
func defConPoints(explain []fplExplain) int {
	total := 0
	for _, ex := range explain {
		for _, s := range ex.Stats {
			if s.Identifier == "defensive_contribution" {
				total += s.Points
			}
		}
	}
	return total
}

func mapManagers(entries []fplStandingEntry) []fantasy.Manager {
	managers := make([]fantasy.Manager, 0, len(entries))
	for i, e := range entries {
		managers = append(managers, fantasy.Manager{
			GameID:      gameID,
			ExternalID:  e.Entry,
			Name:        e.EntryName,
			OverallRank: i + 1,
			UpdatedAt:   time.Now().UTC(),
		})
	}
	return managers
}

func mapPicks(managerExternalID int, gw int, raw []fplPick) []fantasy.ManagerPick {
	picks := make([]fantasy.ManagerPick, 0, len(raw))
	for _, p := range raw {
		picks = append(picks, fantasy.ManagerPick{
			ManagerID:     strconv.Itoa(managerExternalID),
			PlayerID:      strconv.Itoa(p.Element),
			GameID:        gameID,
			GW:            gw,
			IsCaptain:     p.IsCaptain,
			IsViceCaptain: p.IsViceCaptain,
			Multiplier:    p.Multiplier,
		})
	}
	return picks
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func currentGW(events []fplEvent) int {
	for _, e := range events {
		if e.IsCurrent {
			return e.ID
		}
	}
	for _, e := range events {
		if e.IsNext {
			return e.ID - 1
		}
	}
	return 1
}
