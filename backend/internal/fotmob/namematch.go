package fotmob

// canonicalNames maps FotMob team names to the canonical PL names v1 (and the
// frontend's team badges) use. Unknown names pass through unchanged.
var canonicalNames = map[string]string{
	"AFC Bournemouth":         "Bournemouth",
	"Brighton & Hove Albion":  "Brighton",
	"Coventry City":           "Coventry",
	"Hull City":               "Hull",
	"Ipswich Town":            "Ipswich",
	"Leeds United":            "Leeds",
	"Leicester City":          "Leicester",
	"Manchester City":         "Man City",
	"Manchester United":       "Man Utd",
	"Newcastle United":        "Newcastle",
	"Nottingham Forest":       "Nott'm Forest",
	"Tottenham Hotspur":       "Spurs",
	"West Ham United":         "West Ham",
	"Wolverhampton Wanderers": "Wolves",
}

func canonicalTeam(name string) string {
	if c, ok := canonicalNames[name]; ok {
		return c
	}
	return name
}
