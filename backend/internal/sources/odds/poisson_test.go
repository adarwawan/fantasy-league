package odds_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"fantasy-league/internal/sources/odds"
)

func approx(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// --- poissonCDF (via EstimateLambdaTotal as a proxy) ---

// TestEstimateLambdaTotal_KnownLine verifies that a 2.5 line at 1.70/2.10
// yields λ ≈ 2.89 (within ±0.10).
// Calculation: pUnder = (1/2.10) / (1/1.70 + 1/2.10) ≈ 0.4475
// solveLambda(2.5, 0.4475) → P(X<=2|λ)=0.4475 → λ ≈ 2.89
func TestEstimateLambdaTotal_KnownLine(t *testing.T) {
	bks := []odds.Bookmaker{
		{
			Key:        "bet365",
			LastUpdate: time.Now(),
			Markets: []odds.Market{
				{
					Key:        "totals",
					LastUpdate: time.Now(),
					Outcomes: []odds.Outcome{
						{Name: "Over", Description: "2.5", Price: 1.70},
						{Name: "Under", Description: "2.5", Price: 2.10},
					},
				},
			},
		},
	}
	got, ok := odds.EstimateLambdaTotal(bks)
	if !ok {
		t.Fatal("EstimateLambdaTotal returned ok=false")
	}
	if !approx(got, 2.89, 0.10) {
		t.Errorf("λ: got %.4f, want ~2.89 (±0.10)", got)
	}
}

// TestEstimateLambdaTotal_MultipleLines verifies averaging of multiple lines
// from the same bookmaker.
func TestEstimateLambdaTotal_MultipleLines(t *testing.T) {
	bks := []odds.Bookmaker{
		{
			Key:        "unibet",
			LastUpdate: time.Now(),
			Markets: []odds.Market{
				{
					Key:        "totals",
					LastUpdate: time.Now(),
					Outcomes: []odds.Outcome{
						{Name: "Over", Description: "1.5", Price: 1.30},
						{Name: "Under", Description: "1.5", Price: 3.20},
						{Name: "Over", Description: "2.5", Price: 1.70},
						{Name: "Under", Description: "2.5", Price: 2.10},
					},
				},
			},
		},
	}
	got, ok := odds.EstimateLambdaTotal(bks)
	if !ok {
		t.Fatal("EstimateLambdaTotal returned ok=false")
	}
	// Both lines point to a similar λ; result should be in a reasonable range.
	if got < 1.5 || got > 4.0 {
		t.Errorf("λ out of range: got %.4f", got)
	}
}

// TestEstimateLambdaTotal_WeightedMedian checks that a fresher bookmaker
// dominates when two bookmakers disagree.
func TestEstimateLambdaTotal_WeightedMedian(t *testing.T) {
	old := time.Now().Add(-10 * time.Minute)
	fresh := time.Now()

	// Old bookmaker implies very low λ (very likely Under 2.5 → low scoring).
	// Fresh bookmaker implies higher λ (balanced market).
	bks := []odds.Bookmaker{
		{
			Key:        "old-bk",
			LastUpdate: old,
			Markets: []odds.Market{
				{
					Key:        "totals",
					LastUpdate: old,
					Outcomes: []odds.Outcome{
						{Name: "Over", Description: "2.5", Price: 4.00},
						{Name: "Under", Description: "2.5", Price: 1.22},
					},
				},
			},
		},
		{
			Key:        "fresh-bk",
			LastUpdate: fresh,
			Markets: []odds.Market{
				{
					Key:        "totals",
					LastUpdate: fresh,
					Outcomes: []odds.Outcome{
						{Name: "Over", Description: "2.5", Price: 1.90},
						{Name: "Under", Description: "2.5", Price: 1.90},
					},
				},
			},
		},
	}
	got, ok := odds.EstimateLambdaTotal(bks)
	if !ok {
		t.Fatal("EstimateLambdaTotal returned ok=false")
	}
	// Fresh bookmaker (50/50 → λ ≈ 2.5) should dominate the old one.
	if !approx(got, 2.50, 0.20) {
		t.Errorf("λ: got %.4f, want ~2.50 (±0.20)", got)
	}
}

// TestEstimateLambdaTotal_NoTotals returns false when no totals market exists.
func TestEstimateLambdaTotal_NoTotals(t *testing.T) {
	bks := []odds.Bookmaker{
		{
			Key:        "bk",
			LastUpdate: time.Now(),
			Markets: []odds.Market{
				{Key: "h2h", Outcomes: []odds.Outcome{
					{Name: "Home", Price: 2.0},
					{Name: "Draw", Price: 3.0},
					{Name: "Away", Price: 4.0},
				}},
			},
		},
	}
	_, ok := odds.EstimateLambdaTotal(bks)
	if ok {
		t.Error("expected ok=false for missing totals market")
	}
}

// --- SplitLambda ---

func TestSplitLambda_SumsToTotal(t *testing.T) {
	bks := bookmakerWithH2H("Portugal", 1.40, 4.50, "DR Congo", 8.00)
	lHome, lAway := odds.SplitLambda(3.0, bks)
	if !approx(lHome+lAway, 3.0, 1e-9) {
		t.Errorf("lHome+lAway = %.6f, want 3.0", lHome+lAway)
	}
}

func TestSplitLambda_FavouriteScoresMore(t *testing.T) {
	// Portugal heavy favourite → should have higher λ than DR Congo.
	bks := bookmakerWithH2H("Portugal", 1.40, 4.50, "DR Congo", 8.00)
	lHome, lAway := odds.SplitLambda(3.0, bks)
	if lHome <= lAway {
		t.Errorf("expected lHome (%.4f) > lAway (%.4f) for heavy favourite", lHome, lAway)
	}
}

func TestSplitLambda_EvenOddsEqualSplit(t *testing.T) {
	bks := bookmakerWithH2H("Brazil", 2.00, 3.00, "Argentina", 2.00)
	lHome, lAway := odds.SplitLambda(2.5, bks)
	// Symmetric odds → near-equal split (draw odds dilute but ratio stays 1).
	if !approx(lHome, lAway, 0.01) {
		t.Errorf("expected near-equal split; lHome=%.4f, lAway=%.4f", lHome, lAway)
	}
}

func TestSplitLambda_NoH2H_EqualFallback(t *testing.T) {
	bks := []odds.Bookmaker{{Key: "bk", Markets: []odds.Market{{Key: "totals"}}}}
	lHome, lAway := odds.SplitLambda(2.0, bks)
	if !approx(lHome, 1.0, 1e-9) || !approx(lAway, 1.0, 1e-9) {
		t.Errorf("expected 1/1 fallback; got %.4f / %.4f", lHome, lAway)
	}
}

// --- CleanSheetPct ---

func TestCleanSheetPct_ZeroLambda(t *testing.T) {
	pct := odds.CleanSheetPct(0)
	if !approx(pct, 100.0, 1e-9) {
		t.Errorf("CleanSheetPct(0) = %.4f, want 100.0", pct)
	}
}

func TestCleanSheetPct_KnownValues(t *testing.T) {
	cases := []struct {
		lambda float64
		want   float64 // e^(-λ)*100
	}{
		{1.0, 36.79},
		{2.0, 13.53},
		{0.5, 60.65},
	}
	for _, tc := range cases {
		got := odds.CleanSheetPct(tc.lambda)
		if !approx(got, tc.want, 0.01) {
			t.Errorf("CleanSheetPct(%.1f) = %.4f, want %.2f", tc.lambda, got, tc.want)
		}
	}
}

// --- AggregateBookmakers ---

func TestAggregateBookmakers_FromFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/betting.json")
	if err != nil {
		t.Fatal(err)
	}
	var matches []odds.OddsMatch
	if err := json.Unmarshal(raw, &matches); err != nil {
		t.Fatal(err)
	}

	result := odds.AggregateBookmakers(matches)

	if len(result) != 2 {
		t.Fatalf("expected 2 MatchOdds, got %d", len(result))
	}

	// Match 1: Portugal vs DR Congo — Portugal heavy favourite.
	m := result[0]
	if m.OddsMatchID != "abc123" {
		t.Errorf("OddsMatchID: got %q", m.OddsMatchID)
	}
	if m.LambdaHome <= m.LambdaAway {
		t.Errorf("Portugal should score more: home=%.4f, away=%.4f", m.LambdaHome, m.LambdaAway)
	}
	// CS% for Portugal (home) = e^(-lambdaAway)*100, should be high (low λAway).
	if m.HomeCSPct < 50 {
		t.Errorf("Portugal CS%% should be high, got %.2f", m.HomeCSPct)
	}
	// λ sanity: total should be ~ 2.7.
	total := m.LambdaHome + m.LambdaAway
	if !approx(total, 2.7, 0.3) {
		t.Errorf("total λ: got %.4f, want ~2.7 ±0.3", total)
	}

	// Match 2: Brazil vs Argentina — roughly even.
	m2 := result[1]
	if m2.OddsMatchID != "def456" {
		t.Errorf("OddsMatchID: got %q", m2.OddsMatchID)
	}
	// Even 1.90/1.90 over/under → λ ≈ 2.5.
	total2 := m2.LambdaHome + m2.LambdaAway
	if !approx(total2, 2.5, 0.3) {
		t.Errorf("total λ: got %.4f, want ~2.5 ±0.3", total2)
	}

	// FetchedAt should be recent.
	if time.Since(m.FetchedAt) > 5*time.Second {
		t.Error("FetchedAt is stale")
	}
}

func TestAggregateBookmakers_SkipsMatchWithNoTotals(t *testing.T) {
	matches := []odds.OddsMatch{
		{
			ID:       "no-totals",
			HomeTeam: "A",
			AwayTeam: "B",
			Bookmakers: []odds.Bookmaker{
				{Key: "bk", Markets: []odds.Market{{Key: "h2h"}}},
			},
		},
	}
	result := odds.AggregateBookmakers(matches)
	if len(result) != 0 {
		t.Errorf("expected 0 results for match with no totals, got %d", len(result))
	}
}

// --- helpers ---

func bookmakerWithH2H(home string, homePrice, drawPrice float64, away string, awayPrice float64) []odds.Bookmaker {
	return []odds.Bookmaker{
		{
			Key:        "bk",
			LastUpdate: time.Now(),
			Markets: []odds.Market{
				{
					Key:        "h2h",
					LastUpdate: time.Now(),
					Outcomes: []odds.Outcome{
						{Name: home, Price: homePrice},
						{Name: "Draw", Price: drawPrice},
						{Name: away, Price: awayPrice},
					},
				},
			},
		},
	}
}
