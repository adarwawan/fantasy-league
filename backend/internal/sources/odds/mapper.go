package odds

import (
	"fmt"
	"log"

	"fantasy-league/internal/fantasy"
)

// MapTeams resolves HomeTeam/AwayTeam strings on each MatchOdds record to
// fantasy.Team IDs. The nameOverrides map (namemap.go) translates every known
// betting API name to the exact fantasy.Team.Name; a single map lookup plus
// an exact index hit is all that is needed.
//
// The caller must pass teams scoped to a single game; the function panics if
// any team belongs to a different game.
func MapTeams(odds []MatchOdds, teams []fantasy.Team, gameID string) []MatchOdds {
	for _, t := range teams {
		if t.GameID != gameID {
			panic(fmt.Sprintf("odds.MapTeams: team %q has GameID %q, expected %q", t.Name, t.GameID, gameID))
		}
	}

	byName := make(map[string]string, len(teams)) // team.Name → team.ID
	for _, t := range teams {
		byName[t.Name] = t.ID
	}

	var result []MatchOdds
	for _, m := range odds {
		m.HomeTeamID = lookupTeam(m.HomeTeam, byName)
		if m.HomeTeamID == "" {
			log.Printf("odds mapper: unresolved home team %q (game=%s, match=%s) — skipping", m.HomeTeam, gameID, m.OddsMatchID)
			continue
		}
		m.AwayTeamID = lookupTeam(m.AwayTeam, byName)
		if m.AwayTeamID == "" {
			log.Printf("odds mapper: unresolved away team %q (game=%s, match=%s) — skipping", m.AwayTeam, gameID, m.OddsMatchID)
			continue
		}
		result = append(result, m)
	}
	return result
}

// lookupTeam translates a raw betting API team name to a fantasy team ID via
// the override map then an exact name index hit.
func lookupTeam(raw string, byName map[string]string) string {
	canonical := raw
	if mapped, ok := nameOverrides[raw]; ok {
		canonical = mapped
	}
	return byName[canonical]
}

// ---------------------------------------------------------------------------
// Fixture linking
// ---------------------------------------------------------------------------

// LinkFixtures associates each MatchOdds record with a fantasy.Fixture ID.
// Behaviour is controlled by cfg.StrictSides:
//   - false (WCF): unordered pair key; swaps λ/CS% if sides are flipped.
//   - true  (FPL): ordered pair key; logs a warning on mismatch, leaves FixtureID empty.
func LinkFixtures(odds []MatchOdds, fixtures []fantasy.Fixture, cfg GameOddsConfig) []MatchOdds {
	type fixtureEntry struct {
		fixture fantasy.Fixture
		homeID  string
	}

	if cfg.StrictSides {
		index := make(map[string]fixtureEntry, len(fixtures))
		for _, f := range fixtures {
			index[orderedKey(f.HomeTeamID, f.AwayTeamID)] = fixtureEntry{f, f.HomeTeamID}
		}
		result := make([]MatchOdds, len(odds))
		for i, m := range odds {
			if entry, ok := index[orderedKey(m.HomeTeamID, m.AwayTeamID)]; ok {
				m.FixtureID = entry.fixture.ID
			} else if _, ok := index[orderedKey(m.AwayTeamID, m.HomeTeamID)]; ok {
				log.Printf("odds linker: strict-sides mismatch for match %s (home/away flipped vs fixture); leaving unlinked", m.OddsMatchID)
			}
			result[i] = m
		}
		return result
	}

	index := make(map[string]fixtureEntry, len(fixtures))
	for _, f := range fixtures {
		index[unorderedKey(f.HomeTeamID, f.AwayTeamID)] = fixtureEntry{f, f.HomeTeamID}
	}

	result := make([]MatchOdds, len(odds))
	for i, m := range odds {
		entry, ok := index[unorderedKey(m.HomeTeamID, m.AwayTeamID)]
		if !ok {
			result[i] = m
			continue
		}
		m.FixtureID = entry.fixture.ID
		if m.HomeTeamID != entry.homeID {
			m.LambdaHome, m.LambdaAway = m.LambdaAway, m.LambdaHome
			m.HomeCSPct, m.AwayCSPct = m.AwayCSPct, m.HomeCSPct
		}
		result[i] = m
	}
	return result
}

func unorderedKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func orderedKey(home, away string) string {
	return home + "|" + away
}
