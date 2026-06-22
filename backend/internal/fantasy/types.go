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
	UpdatedAt       time.Time
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
