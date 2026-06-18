package wcf

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"fantasy-league/internal/fantasy"
)

const gameID = "wcf"

// WCF positions are already the canonical short form.
var validPositions = map[string]bool{"GK": true, "DEF": true, "MID": true, "FWD": true}

// WCF statuses: "playing"→available, "suspended"/"transferred"→injured.
var statusMap = map[string]string{
	"playing":    "available",
	"suspended":  "injured",
	"transferred": "injured",
}

// lineupOrder defines the position sort order when flattening the lineup map.
var lineupOrder = []string{"GK", "DEF", "MID", "FWD"}

func mapTeams(raw []wcfSquad) []fantasy.Team {
	teams := make([]fantasy.Team, 0, len(raw))
	for _, t := range raw {
		teams = append(teams, fantasy.Team{
			GameID:     gameID,
			ExternalID: t.ID,
			Name:       t.Name,
			ShortName:  t.Abbr,
			UpdatedAt:  time.Now().UTC(),
		})
	}
	return teams
}

func mapPlayers(raw []wcfPlayer) []fantasy.Player {
	players := make([]fantasy.Player, 0, len(raw))
	for _, p := range raw {
		pos := p.Position
		if !validPositions[pos] {
			pos = "MID"
		}
		status := statusMap[p.Status]
		if status == "" {
			status = "available"
		}
		name := playerName(p)

		players = append(players, fantasy.Player{
			GameID:          gameID,
			ExternalID:      p.ID,
			Name:            name,
			TeamID:          strconv.Itoa(p.SquadID),
			Position:        pos,
			Price:           p.Price,
			Form:            p.Stats.Form,
			GlobalOwnership: p.PercentSelected,
			Status:          status,
			UpdatedAt:       time.Now().UTC(),
		})
	}
	return players
}

// playerName returns knownName if set, otherwise "FirstName LastName".
func playerName(p wcfPlayer) string {
	if p.KnownName != nil && *p.KnownName != "" {
		return *p.KnownName
	}
	return fmt.Sprintf("%s %s", p.FirstName, p.LastName)
}

// mapFixtures flattens all tournaments across all rounds into fixtures.
// WCF has no difficulty ratings; we default to 3 (medium) for all matches.
func mapFixtures(rounds []wcfRound) []fantasy.Fixture {
	var fixtures []fantasy.Fixture
	for _, round := range rounds {
		for _, t := range round.Tournaments {
			kickoff := time.Time{}
			if t.Date != "" {
				kickoff, _ = time.Parse(time.RFC3339, t.Date)
			}
			finished := t.Status == "complete"
			fixtures = append(fixtures, fantasy.Fixture{
				GameID:         gameID,
				ExternalID:     t.ID,
				GW:             round.ID,
				HomeTeamID:     strconv.Itoa(t.HomeSquadID),
				AwayTeamID:     strconv.Itoa(t.AwaySquadID),
				HomeDifficulty: 3,
				AwayDifficulty: 3,
				KickoffTime:    kickoff,
				Finished:       finished,
				HomeScore:      t.HomeScore,
				AwayScore:      t.AwayScore,
			})
		}
	}
	return fixtures
}

func mapManagers(entries []wcfRankEntry) []fantasy.Manager {
	managers := make([]fantasy.Manager, 0, len(entries))
	for i, e := range entries {
		managers = append(managers, fantasy.Manager{
			GameID:      gameID,
			ExternalID:  int(e.UserID),
			Name:        e.UserName,
			OverallRank: i + 1,
			UpdatedAt:   time.Now().UTC(),
		})
	}
	return managers
}

// mapPicks builds picks from the history response.
// Lineup players are ordered GK→DEF→MID→FWD (the map[string][]int keys).
// Bench players follow in BenchOrder sequence with multiplier 0.
// Captain gets multiplier 2 (3 if MaxCaptainBooster chip is active).
func mapPicks(managerID int, gw int, entry *wcfPickEntry) []fantasy.ManagerPick {
	if entry == nil {
		return nil
	}

	captainMult := 2
	if isActiveChip(entry.MaxCaptainBooster) {
		captainMult = 3
	}

	captainID := 0
	viceID := 0
	if entry.Captain != nil {
		captainID = *entry.Captain
	}
	if entry.Vice != nil {
		viceID = *entry.Vice
	}

	managerIDStr := strconv.Itoa(managerID)
	var picks []fantasy.ManagerPick

	// Flatten lineup in canonical position order.
	for _, pos := range lineupOrder {
		for _, playerID := range entry.Lineup[pos] {
			mult := 1
			if playerID == captainID {
				mult = captainMult
			}
			picks = append(picks, fantasy.ManagerPick{
				ManagerID:     managerIDStr,
				PlayerID:      strconv.Itoa(playerID),
				GameID:        gameID,
				GW:            gw,
				IsCaptain:     playerID == captainID,
				IsViceCaptain: playerID == viceID,
				Multiplier:    mult,
			})
		}
	}

	// Bench in bench order with multiplier 0.
	benchSeen := map[int]bool{}
	benched := sortedBench(entry)
	for _, playerID := range benched {
		if benchSeen[playerID] {
			continue
		}
		benchSeen[playerID] = true
		picks = append(picks, fantasy.ManagerPick{
			ManagerID:  managerIDStr,
			PlayerID:   strconv.Itoa(playerID),
			GameID:     gameID,
			GW:         gw,
			Multiplier: 0,
		})
	}

	return picks
}

// sortedBench returns bench player IDs from BenchOrder if available,
// otherwise falls back to flattening entry.Bench in position order.
func sortedBench(entry *wcfPickEntry) []int {
	if len(entry.BenchOrder) > 0 {
		return entry.BenchOrder
	}
	var ids []int
	for _, pos := range lineupOrder {
		ids = append(ids, entry.Bench[pos]...)
	}
	return ids
}

// isActiveChip returns true when a chip field holds a non-nil, non-false value.
func isActiveChip(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

// extractRanks sorts entries by OverallRank ascending before mapping.
func extractRanks(entries []wcfRankEntry) []wcfRankEntry {
	sorted := make([]wcfRankEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OverallRank < sorted[j].OverallRank
	})
	return sorted
}
