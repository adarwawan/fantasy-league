package shotzone

// teamNameOverrides maps Understat team titles to canonical PL team names.
// Same static-map pattern as setpiece/namematch.go — duplicated rather than
// imported, keeping the two Understat modules fully independent.
var teamNameOverrides = map[string]string{
	"Manchester City":         "Man City",
	"Manchester United":       "Man Utd",
	"Newcastle United":        "Newcastle",
	"Tottenham":               "Spurs",
	"Wolverhampton Wanderers": "Wolves",
	"Nottingham Forest":       "Nott'm Forest",
	"Brighton":                "Brighton",
	"Leeds":                   "Leeds",
	"Leicester":               "Leicester",
}

// canonicalTeam resolves an Understat team title to the canonical PL name.
func canonicalTeam(understatTitle string) string {
	if c, ok := teamNameOverrides[understatTitle]; ok {
		return c
	}
	return understatTitle
}
