package shotzone

import (
	"strconv"
	"strings"
	"time"
)

const understatDateLayout = "2006-01-02 15:04:05"

// ParseShots extracts open-play shot events from a match's shot stream, one
// Event per qualifying shot. Set pieces, free kicks, corners and penalties
// are excluded entirely (the setpiece board tracks those), and own goals are
// dropped since they aren't a shot a player should be ranked on.
func ParseShots(matchID string, shots []shot) []Event {
	events := make([]Event, 0, len(shots))
	for _, s := range shots {
		if s.Situation != situationOpenPlay {
			continue
		}
		if s.Result == "OwnGoal" {
			continue
		}

		teamFor, teamAgainst := s.HTeam, s.ATeam
		if s.HA == "a" {
			teamFor, teamAgainst = s.ATeam, s.HTeam
		}

		events = append(events, Event{
			MatchID:     matchID,
			Season:      s.Season,
			MatchDate:   parseDate(s.Date),
			Minute:      atoi(s.Minute),
			TeamFor:     canonicalTeam(teamFor),
			TeamAgainst: canonicalTeam(teamAgainst),
			Zone:        ClassifyZone(atof(s.X), atof(s.Y)),
			PlayerID:    s.PlayerID,
			PlayerName:  s.Player,
			IsGoal:      s.Result == "Goal",
			XG:          atof(s.XG),
		})
	}
	return events
}

func atoi(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func atof(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func parseDate(s string) time.Time {
	t, err := time.Parse(understatDateLayout, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
