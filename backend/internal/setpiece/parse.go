package setpiece

import (
	"strconv"
	"strings"
	"time"
)

// understatDateLayout is the format Understat uses for a shot's `date`.
const understatDateLayout = "2006-01-02 15:04:05"

// ParseShots classifies a match's shots into set-piece events, one Event per
// qualifying shot. Both roles are extracted in a single pass since they share
// the shot stream (docs §6, P1). Shots in open play are dropped.
func ParseShots(matchID string, shots []shot) []Event {
	events := make([]Event, 0, len(shots))
	for _, s := range shots {
		role, duty, ok := classify(s.Situation)
		if !ok {
			continue
		}

		team := s.HTeam
		if s.HA == "a" {
			team = s.ATeam
		}

		events = append(events, Event{
			MatchID:       matchID,
			Season:        s.Season,
			MatchDate:     parseDate(s.Date),
			Minute:        atoi(s.Minute),
			UnderstatTeam: canonicalTeam(team),
			Role:          role,
			Duty:          duty,
			PlayerID:      s.PlayerID,
			PlayerName:    s.Player,
			IsHeader:      strings.EqualFold(s.ShotType, "Head"),
			IsGoal:        s.Result == "Goal",
			XG:            atof(s.XG),
		})
	}
	return events
}

// classify maps an Understat `situation` to a (role, duty). ok is false for
// situations we don't track (OpenPlay).
func classify(situation string) (Role, Duty, bool) {
	switch situation {
	case situationPenalty:
		return RoleTaker, DutyPenalty, true
	case situationDirectFreekick:
		return RoleTaker, DutyDFK, true
	case situationFromCorner:
		return RoleTarget, DutyCorner, true
	case situationSetPiece:
		return RoleTarget, DutySetPiece, true
	default:
		return "", "", false
	}
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
