package shotzone

import (
	"math"
	"sort"
	"time"
)

// AggregateTaken ranks each team's own players by zone over the rolling
// window (that team's own most recent windowMatches). Also computes the
// cross-zone ZoneOverall bucket per player, mirroring setpiece's DutyAll.
func AggregateTaken(events []Event, windowMatches int) []TakenRow {
	windowed := trimToWindow(events, windowMatches, func(e Event) string { return e.TeamFor })

	type key struct {
		team, player string
		zone         Zone
	}
	type acc struct {
		name     string
		shots    int
		goals    int
		xg       float64
		lastSeen time.Time
	}
	accs := map[key]*acc{}
	add := func(k key, e Event) {
		a := accs[k]
		if a == nil {
			a = &acc{name: e.PlayerName}
			accs[k] = a
		}
		a.shots++
		a.xg += e.XG
		if e.IsGoal {
			a.goals++
		}
		if e.MatchDate.After(a.lastSeen) {
			a.lastSeen = e.MatchDate
		}
	}
	for _, e := range windowed {
		add(key{e.TeamFor, e.PlayerID, e.Zone}, e)
		add(key{e.TeamFor, e.PlayerID, ZoneOverall}, e)
	}

	rows := make([]TakenRow, 0, len(accs))
	for k, a := range accs {
		rows = append(rows, TakenRow{
			Team: k.team, Zone: k.zone, PlayerID: k.player, PlayerName: a.name,
			Shots: a.shots, Goals: a.goals, XG: round2(a.xg), LastSeen: a.lastSeen,
		})
	}
	assignTakenRanks(rows)
	return rows
}

// AggregateConceded totals shots/goals/xG a team has faced per zone over its
// own rolling window (its own most recent windowMatches) — a team-weakness
// signal, not a leaderboard, so no player breakdown.
func AggregateConceded(events []Event, windowMatches int) []ConcededRow {
	windowed := trimToWindow(events, windowMatches, func(e Event) string { return e.TeamAgainst })

	type key struct {
		team string
		zone Zone
	}
	type acc struct {
		shots    int
		goals    int
		xg       float64
		lastSeen time.Time
	}
	accs := map[key]*acc{}
	add := func(k key, e Event) {
		a := accs[k]
		if a == nil {
			a = &acc{}
			accs[k] = a
		}
		a.shots++
		a.xg += e.XG
		if e.IsGoal {
			a.goals++
		}
		if e.MatchDate.After(a.lastSeen) {
			a.lastSeen = e.MatchDate
		}
	}
	for _, e := range windowed {
		add(key{e.TeamAgainst, e.Zone}, e)
		add(key{e.TeamAgainst, ZoneOverall}, e)
	}

	rows := make([]ConcededRow, 0, len(accs))
	for k, a := range accs {
		rows = append(rows, ConcededRow{
			Team: k.team, Zone: k.zone, Shots: a.shots, Goals: a.goals,
			XG: round2(a.xg), LastSeen: a.lastSeen,
		})
	}
	return rows
}

// TeamWindows returns, per team, the ids of its most recent `window` matches
// (both sides of every match count, so a team's window is its real fixture
// list, not just matches where it produced events). Matches are ordered by
// (date, match id) so two same-day matches can't both straddle the cutoff.
// window <= 0 returns nil (no windowing).
func TeamWindows(events []Event, window int) map[string]map[string]bool {
	if window <= 0 {
		return nil
	}
	type ref struct {
		id   string
		date time.Time
	}
	seen := map[string]map[string]bool{}
	byTeam := map[string][]ref{}
	note := func(team string, e Event) {
		if seen[team] == nil {
			seen[team] = map[string]bool{}
		}
		if !seen[team][e.MatchID] {
			seen[team][e.MatchID] = true
			byTeam[team] = append(byTeam[team], ref{e.MatchID, e.MatchDate})
		}
	}
	for _, e := range events {
		note(e.TeamFor, e)
		note(e.TeamAgainst, e)
	}
	out := make(map[string]map[string]bool, len(byTeam))
	for team, refs := range byTeam {
		sort.Slice(refs, func(i, j int) bool {
			if !refs[i].date.Equal(refs[j].date) {
				return refs[i].date.After(refs[j].date)
			}
			return refs[i].id > refs[j].id
		})
		if len(refs) > window {
			refs = refs[:window]
		}
		set := make(map[string]bool, len(refs))
		for _, r := range refs {
			set[r.id] = true
		}
		out[team] = set
	}
	return out
}

// trimToWindow keeps only events whose match is in the key's (as extracted by
// teamOf) window. Taken and conceded aggregation call this with different
// teamOf functions, so each side's rolling window is measured in its own matches.
func trimToWindow(events []Event, window int, teamOf func(Event) string) []Event {
	windows := TeamWindows(events, window)
	if windows == nil {
		return events
	}
	out := make([]Event, 0, len(events))
	for _, e := range events {
		if windows[teamOf(e)][e.MatchID] {
			out = append(out, e)
		}
	}
	return out
}

// assignTakenRanks sorts rows within each (team, zone) group by shots desc
// (ties: goals, then xG) and assigns 1-based ranks.
func assignTakenRanks(rows []TakenRow) {
	groups := map[[2]string][]int{}
	for i, r := range rows {
		key := [2]string{r.Team, string(r.Zone)}
		groups[key] = append(groups[key], i)
	}
	for _, idxs := range groups {
		sort.SliceStable(idxs, func(a, b int) bool {
			ra, rb := rows[idxs[a]], rows[idxs[b]]
			if ra.Shots != rb.Shots {
				return ra.Shots > rb.Shots
			}
			if ra.Goals != rb.Goals {
				return ra.Goals > rb.Goals
			}
			return ra.XG > rb.XG
		})
		for rank, idx := range idxs {
			rows[idx].Rank = rank + 1
		}
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
