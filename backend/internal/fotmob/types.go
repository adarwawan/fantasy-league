// Package fotmob is v2 of the set-piece and shot-zone data feeds, sourced from
// FotMob because Understat is in maintenance. One match fetch yields both
// features' events; storage, aggregation and read handlers are the setpiece and
// shotzone packages' own, pointed at their *2 tables (migration 014).
//
// Like those modules it is isolated: it never touches the players/fixtures
// pipeline.
package fotmob

import "time"

// --- FotMob wire types -------------------------------------------------------

// leaguePayload is /api/data/leagues; only the fixture list is needed.
type leaguePayload struct {
	Fixtures struct {
		AllMatches []leagueMatch `json:"allMatches"`
	} `json:"fixtures"`
}

type leagueMatch struct {
	ID     string `json:"id"`
	Status struct {
		Finished  bool `json:"finished"`
		Cancelled bool `json:"cancelled"`
	} `json:"status"`
}

// matchPayload is /api/data/matchDetails, reduced to what we use.
type matchPayload struct {
	General struct {
		HomeTeam     team   `json:"homeTeam"`
		AwayTeam     team   `json:"awayTeam"`
		MatchTimeUTC string `json:"matchTimeUTCDate"`
		Finished     bool   `json:"finished"`
	} `json:"general"`
	Content struct {
		Shotmap *struct {
			Shots []shot `json:"shots"`
		} `json:"shotmap"`
	} `json:"content"`
}

type team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// shot is one shotmap entry. Coordinates are in metres on a 105x68 pitch,
// already normalised so the shooting team attacks x=105.
type shot struct {
	ID        int64   `json:"id"`
	EventType string  `json:"eventType"` // Goal | AttemptSaved | Miss | Post
	TeamID    int     `json:"teamId"`
	PlayerID  int64   `json:"playerId"`
	Name      string  `json:"playerName"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Min       int     `json:"min"`
	XG        float64 `json:"expectedGoals"`
	ShotType  string  `json:"shotType"`  // LeftFoot | RightFoot | Header | OtherBodyParts
	Situation string  `json:"situation"` // see parse.go
	IsOwnGoal bool    `json:"isOwnGoal"`
}

// matchData is the slim form we cache in Redis (the raw payload is ~300KB).
type matchData struct {
	Date  time.Time `json:"date"`
	Home  team      `json:"home"`
	Away  team      `json:"away"`
	Shots []shot    `json:"shots"`
}
