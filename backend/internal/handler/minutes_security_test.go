package handler

import (
	"testing"

	"fantasy-league/internal/store"
)

func TestMinutesSecurity(t *testing.T) {
	f := func(v float64) *float64 { return &v }

	tests := []struct {
		name     string
		mins     []store.GWMinutes
		wantRate *float64
		wantAvg  int
	}{
		{
			name:     "no history",
			mins:     nil,
			wantRate: nil,
			wantAvg:  0,
		},
		{
			name: "every fixture started",
			mins: []store.GWMinutes{
				{GW: 1, Minutes: 90, Starts: 1, Fixtures: 1},
				{GW: 2, Minutes: 88, Starts: 1, Fixtures: 1},
			},
			wantRate: f(1),
			wantAvg:  89,
		},
		{
			name: "double gameweek counts twice",
			mins: []store.GWMinutes{
				{GW: 1, Minutes: 160, Starts: 2, Fixtures: 2}, // started both legs
				{GW: 2, Minutes: 0, Starts: 0, Fixtures: 1},   // benched
			},
			wantRate: f(2.0 / 3.0), // 2 starts of 3 fixtures
			wantAvg:  53,           // (160 + 0) / 3 fixtures, not per gameweek
		},
		{
			name: "blank gameweek excluded from rate and average",
			mins: []store.GWMinutes{
				{GW: 1, Minutes: 90, Starts: 1, Fixtures: 1},
				{GW: 2, Minutes: 0, Starts: 0, Fixtures: 0}, // blank: club didn't play
			},
			wantRate: f(1), // 1 start of 1 played fixture
			wantAvg:  90,   // blank not averaged in
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotRate, gotAvg := minutesSecurity(tt.mins)
			switch {
			case tt.wantRate == nil && gotRate != nil:
				t.Fatalf("start rate: want nil, got %v", *gotRate)
			case tt.wantRate != nil && gotRate == nil:
				t.Fatalf("start rate: want %v, got nil", *tt.wantRate)
			case tt.wantRate != nil && gotRate != nil:
				if diff := *gotRate - *tt.wantRate; diff > 1e-9 || diff < -1e-9 {
					t.Errorf("start rate: want %v, got %v", *tt.wantRate, *gotRate)
				}
			}
			if gotAvg != tt.wantAvg {
				t.Errorf("avg minutes: want %d, got %d", tt.wantAvg, gotAvg)
			}
		})
	}
}
