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

// trimToWindow keeps only events from each key's (as extracted by teamOf)
// most recent `window` distinct match dates. window <= 0 keeps everything.
// Taken and conceded aggregation call this with different teamOf functions,
// so each side's rolling window is measured in its own matches.
func trimToWindow(events []Event, window int, teamOf func(Event) string) []Event {
	if window <= 0 {
		return events
	}
	teamDates := map[string]map[string]time.Time{}
	for _, e := range events {
		team := teamOf(e)
		m := teamDates[team]
		if m == nil {
			m = map[string]time.Time{}
			teamDates[team] = m
		}
		m[e.MatchID] = e.MatchDate
	}
	cutoff := map[string]time.Time{}
	for team, matches := range teamDates {
		dates := make([]time.Time, 0, len(matches))
		for _, d := range matches {
			dates = append(dates, d)
		}
		sort.Slice(dates, func(i, j int) bool { return dates[i].After(dates[j]) })
		if len(dates) > window {
			cutoff[team] = dates[window-1]
		}
	}
	out := make([]Event, 0, len(events))
	for _, e := range events {
		team := teamOf(e)
		if c, ok := cutoff[team]; ok && e.MatchDate.Before(c) {
			continue
		}
		out = append(out, e)
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
