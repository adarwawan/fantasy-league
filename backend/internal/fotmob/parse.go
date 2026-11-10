package fotmob

import (
	"strconv"

	"fantasy-league/internal/setpiece"
	"fantasy-league/internal/shotzone"
)

// Pitch size FotMob's shot coordinates are expressed in; dividing by it gives
// the 0-1 coordinates shotzone.ClassifyZone expects (same zone spec as v1).
const (
	pitchLength = 105.0
	pitchWidth  = 68.0
)

// FotMob `situation` values, mapped to Understat's so v2 reports the same
// information as v1:
//
//	Penalty                                        -> taker/penalty
//	FreeKick (direct free-kick shot)               -> taker/dfk
//	FromCorner                                     -> target/corner
//	SetPiece (second phase of an indirect FK)      -> target/setpiece
//	RegularPlay, FastBreak, IndividualPlay         -> open play
//	ThrowInSetPiece                                -> open play: Understat has no
//	  throw-in category and files these under OpenPlay, so v2 does too.
//
// Anything unrecognised is dropped rather than guessed.
func classify(situation string) (sp *setpiece.Role, duty setpiece.Duty, openPlay bool) {
	taker, target := setpiece.RoleTaker, setpiece.RoleTarget
	switch situation {
	case "Penalty":
		return &taker, setpiece.DutyPenalty, false
	case "FreeKick":
		return &taker, setpiece.DutyDFK, false
	case "FromCorner":
		return &target, setpiece.DutyCorner, false
	case "SetPiece":
		return &target, setpiece.DutySetPiece, false
	case "RegularPlay", "FastBreak", "IndividualPlay", "ThrowInSetPiece":
		return nil, "", true
	}
	return nil, "", false
}

// Parse turns one match into set-piece and open-play shot events. Own goals are
// dropped from both, as in v1 (a deflection isn't a player's shot or a set-piece
// threat). Blocked shots are kept: Understat counts them too.
func Parse(matchID, season string, m matchData) ([]setpiece.Event, []shotzone.Event) {
	var sp []setpiece.Event
	var sz []shotzone.Event
	for _, s := range m.Shots {
		if s.IsOwnGoal {
			continue
		}
		forTeam, against := m.Home, m.Away
		if s.TeamID == m.Away.ID {
			forTeam, against = m.Away, m.Home
		}
		shotID := strconv.FormatInt(s.ID, 10)
		playerID := strconv.FormatInt(s.PlayerID, 10)
		isGoal := s.EventType == "Goal"

		role, duty, open := classify(s.Situation)
		switch {
		case role != nil:
			sp = append(sp, setpiece.Event{
				MatchID:       matchID,
				ShotID:        shotID,
				Season:        season,
				MatchDate:     m.Date,
				Minute:        s.Min,
				UnderstatTeam: canonicalTeam(forTeam.Name),
				Role:          *role,
				Duty:          duty,
				PlayerID:      playerID,
				PlayerName:    s.Name,
				IsHeader:      s.ShotType == "Header",
				IsGoal:        isGoal,
				XG:            s.XG,
			})
		case open:
			sz = append(sz, shotzone.Event{
				MatchID:     matchID,
				ShotID:      shotID,
				Season:      season,
				MatchDate:   m.Date,
				Minute:      s.Min,
				TeamFor:     canonicalTeam(forTeam.Name),
				TeamAgainst: canonicalTeam(against.Name),
				Zone:        shotzone.ClassifyZone(s.X/pitchLength, s.Y/pitchWidth),
				PlayerID:    playerID,
				PlayerName:  s.Name,
				IsGoal:      isGoal,
				XG:          s.XG,
			})
		}
	}
	return sp, sz
}
