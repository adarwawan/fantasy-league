package fantasy

import "time"

type Player struct {
	ID              string
	GameID          string
	ExternalID      int
	Name            string
	TeamID          string
	Position        string
	Price           float64
	Form            float64
	GlobalOwnership float64
	Status          string
	News            string
	// Team-internal taker ranks for set-piece duties (1 = first choice), nil when
	// the player holds no such duty. Editorially maintained by FPL, so they can
	// lag real on-pitch changes.
	PenaltiesOrder       *int
	DirectFreekicksOrder *int
	CornersIndirectOrder *int
	UpdatedAt            time.Time
}

type Team struct {
	ID         string
	GameID     string
	ExternalID int
	Name       string
	ShortName  string
	AttForm    float64 // avg goals scored, last 5 GWs
	DefForm    float64 // avg goals conceded, last 5 GWs
	OvrForm    float64 // avg points, last 5 GWs
	UpdatedAt  time.Time
}

type Fixture struct {
	ID             string
	GameID         string
	ExternalID     int
	GW             int
	HomeTeamID     string
	AwayTeamID     string
	HomeDifficulty int
	AwayDifficulty int
	KickoffTime    time.Time
	Finished       bool
	HomeScore      *int
	AwayScore      *int
}

// PlayerGWStat is a player's stat line for one finished (or in-progress) gameweek.
// PlayerExternalID is the source's player ID; the store resolves it to the internal UUID.
type PlayerGWStat struct {
	GameID           string
	PlayerExternalID int
	GW               int
	Minutes          int
	Points           int
	Goals            int
	Assists          int
	Bonus            int
	CleanSheets      int
	// DefCon is the defensive-contribution points earned in the gameweek (0/2/4),
	// summed per fixture so double gameweeks are scored correctly — not the raw
	// CBIT/recovery action count.
	DefCon int
	// ICT index components for the gameweek (Opta-derived underlying stats).
	Influence  float64
	Creativity float64
	Threat     float64
}

type Manager struct {
	ID          string
	GameID      string
	ExternalID  int
	Name        string
	OverallRank int
	TeamValue   float64
	UpdatedAt   time.Time
}

type ManagerPick struct {
	ManagerID     string
	PlayerID      string
	GameID        string
	GW            int
	IsCaptain     bool
	IsViceCaptain bool
	Multiplier    int
}
