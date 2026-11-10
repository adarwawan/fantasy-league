package setpiece

import (
	"math"
	"sort"
	"time"
)

// AggregateConfig tunes the rolling-window ranking.
type AggregateConfig struct {
	// WindowMatches is how many of a team's most recent matches feed the board.
	WindowMatches int
	// RecencyHalfLife is the half-life (in days, relative to the newest event)
	// for the exponential recency weight. Zero disables weighting (all 1.0).
	RecencyHalfLife time.Duration
	// TeamMatches, when set, is each team's window of real matches (team -> match
	// ids) and replaces the events-derived window. Without it a team's "matches"
	// are only those where it had a set-piece shot, so a quiet match shrinks the
	// window and the board reaches back further than WindowMatches.
	TeamMatches map[string]map[string]bool
}

// Aggregate ranks events into board rows per team/role/duty over the rolling
// window. Takers rank by weighted count; target men rank by a volume×xG blend.
// Target men additionally get a cross-duty 'all' aggregate used for the
// league-wide table.
func Aggregate(events []Event, cfg AggregateConfig) []BoardRow {
	if len(events) == 0 {
		return nil
	}

	// The recency reference is the single most recent event date across the set.
	var newest time.Time
	for _, e := range events {
		if e.MatchDate.After(newest) {
			newest = e.MatchDate
		}
	}

	// Trim to the last WindowMatches distinct match dates per team, so a team's
	// window is measured in its own matches (not calendar time).
	windowed := trimToWindow(events, cfg.WindowMatches, cfg.TeamMatches)

	// Accumulate per (team, role, duty, player). Target men also feed a synthetic
	// duty='all' bucket for the league-wide aggregate.
	type accKey struct {
		team string
		role Role
		duty Duty
		pid  string
	}
	type acc struct {
		name     string
		weighted float64 // recency-weighted count
		xgW      float64 // recency-weighted xG sum
		raw      int
		xgSum    float64
		headers  int
		goals    int
		lastSeen time.Time
	}
	accs := map[accKey]*acc{}

	add := func(k accKey, e Event, w float64) {
		a := accs[k]
		if a == nil {
			a = &acc{name: e.PlayerName}
			accs[k] = a
		}
		a.weighted += w
		a.xgW += w * e.XG
		a.raw++
		a.xgSum += e.XG
		if e.IsHeader {
			a.headers++
		}
		if e.IsGoal {
			a.goals++
		}
		if e.MatchDate.After(a.lastSeen) {
			a.lastSeen = e.MatchDate
		}
	}

	for _, e := range windowed {
		w := recencyWeight(newest, e.MatchDate, cfg.RecencyHalfLife)
		add(accKey{e.UnderstatTeam, e.Role, e.Duty, e.PlayerID}, e, w)
		if e.Role == RoleTarget {
			add(accKey{e.UnderstatTeam, RoleTarget, DutyAll, e.PlayerID}, e, w)
		}
	}

	rows := make([]BoardRow, 0, len(accs))
	for k, a := range accs {
		row := BoardRow{
			UnderstatTeam: k.team,
			Role:          k.role,
			Duty:          k.duty,
			PlayerID:      k.pid,
			PlayerName:    a.name,
			RawCount:      a.raw,
			Goals:         a.goals,
			XGSum:         round2(a.xgSum),
			LastSeen:      a.lastSeen,
		}
		if k.role == RoleTaker {
			// Weighted count — a taker is decided by how often they take, not xG.
			row.WeightedScore = round2(a.weighted)
		} else {
			// Volume × xG blend: weighted volume plus weighted xG, so a defender
			// with many headers ranks even with low individual xG, while genuine
			// threat still lifts the score.
			row.WeightedScore = round2(a.weighted + a.xgW)
			hp := 0.0
			if a.raw > 0 {
				hp = round2(float64(a.headers) / float64(a.raw) * 100)
			}
			row.HeaderPct = &hp
		}
		rows = append(rows, row)
	}

	assignRanks(rows)
	return rows
}

// trimToWindow keeps only events inside each team's window. With `allowed`
// (real match lists) that is a plain membership test; otherwise the window is
// the team's last `window` matches that have events, ordered by (date, match
// id) so same-day matches can't both straddle the cutoff. window <= 0 keeps
// everything.
func trimToWindow(events []Event, window int, allowed map[string]map[string]bool) []Event {
	if allowed == nil {
		if window <= 0 {
			return events
		}
		allowed = eventWindows(events, window)
	}
	out := make([]Event, 0, len(events))
	for _, e := range events {
		if allowed[e.UnderstatTeam][e.MatchID] {
			out = append(out, e)
		}
	}
	return out
}

func eventWindows(events []Event, window int) map[string]map[string]bool {
	type ref struct {
		id   string
		date time.Time
	}
	seen := map[string]map[string]bool{}
	byTeam := map[string][]ref{}
	for _, e := range events {
		if seen[e.UnderstatTeam] == nil {
			seen[e.UnderstatTeam] = map[string]bool{}
		}
		if !seen[e.UnderstatTeam][e.MatchID] {
			seen[e.UnderstatTeam][e.MatchID] = true
			byTeam[e.UnderstatTeam] = append(byTeam[e.UnderstatTeam], ref{e.MatchID, e.MatchDate})
		}
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

// assignRanks sorts rows within each (team, role, duty) group by weighted score
// (desc) and assigns 1-based ranks. Ties break on raw count, then xG sum.
func assignRanks(rows []BoardRow) {
	groups := map[[3]string][]int{}
	for i, r := range rows {
		key := [3]string{r.UnderstatTeam, string(r.Role), string(r.Duty)}
		groups[key] = append(groups[key], i)
	}
	for _, idxs := range groups {
		sort.SliceStable(idxs, func(a, b int) bool {
			ra, rb := rows[idxs[a]], rows[idxs[b]]
			if ra.WeightedScore != rb.WeightedScore {
				return ra.WeightedScore > rb.WeightedScore
			}
			if ra.RawCount != rb.RawCount {
				return ra.RawCount > rb.RawCount
			}
			return ra.XGSum > rb.XGSum
		})
		for rank, idx := range idxs {
			rows[idx].Rank = rank + 1
		}
	}
}

// recencyWeight is 0.5^(ageDays/halfLifeDays); 1.0 when half-life is zero.
func recencyWeight(newest, when time.Time, halfLife time.Duration) float64 {
	if halFDays := halfLife.Hours() / 24; halFDays > 0 && !when.IsZero() {
		ageDays := newest.Sub(when).Hours() / 24
		if ageDays < 0 {
			ageDays = 0
		}
		return math.Pow(0.5, ageDays/halFDays)
	}
	return 1.0
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
