package setpiece

// teamNameOverrides maps Understat team titles to canonical PL team names, used
// only for grouping and display (same static-map pattern as odds/namemap.go).
// Understat's titles already match most PL names; entries here cover the few
// that differ. Unknown names pass through unchanged.
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
