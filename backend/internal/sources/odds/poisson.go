package odds

import (
	"math"
	"sort"
	"time"
)

// poissonCDF returns P(X <= k) for a Poisson distribution with rate lambda.
// k must be a non-negative integer.
func poissonCDF(lambda float64, k int) float64 {
	if lambda <= 0 {
		return 1
	}
	sum := 0.0
	term := math.Exp(-lambda)
	for i := 0; i <= k; i++ {
		sum += term
		if i < k {
			term *= lambda / float64(i+1)
		}
	}
	return sum
}

// solveLambda finds λ such that poissonCDF(λ, floor(line)) == targetCDF
// via binary search over [0, 15].
func solveLambda(line, targetCDF float64) float64 {
	k := int(math.Floor(line))
	lo, hi := 0.0, 15.0
	for i := 0; i < 64; i++ {
		mid := (lo + hi) / 2
		if poissonCDF(mid, k) > targetCDF {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// lambdaFromTotalsOutcomes derives a single λ estimate for one bookmaker's
// totals market. Each over/under pair at a given line yields one estimate;
// if multiple lines are present they are averaged.
func lambdaFromTotalsOutcomes(outcomes []Outcome) (float64, bool) {
	type pair struct{ over, under float64 }
	pairs := map[float64]*pair{}
	for _, o := range outcomes {
		if o.Point == nil {
			continue
		}
		key := *o.Point
		if _, ok := pairs[key]; !ok {
			pairs[key] = &pair{}
		}
		switch o.Name {
		case "Over":
			pairs[key].over = o.Price
		case "Under":
			pairs[key].under = o.Price
		}
	}

	var lambdas []float64
	for line, p := range pairs {
		if p.over <= 0 || p.under <= 0 {
			continue
		}
		// Remove bookmaker margin.
		rawOver := 1 / p.over
		rawUnder := 1 / p.under
		norm := rawOver + rawUnder
		pUnder := rawUnder / norm
		// P(X <= floor(line)) == pUnder
		lambdas = append(lambdas, solveLambda(line, pUnder))
	}

	if len(lambdas) == 0 {
		return 0, false
	}
	sum := 0.0
	for _, l := range lambdas {
		sum += l
	}
	return sum / float64(len(lambdas)), true
}

// weightedMedian returns the median of values weighted by their weights.
// Uses the standard weighted median algorithm (sort by value, accumulate weights).
func weightedMedian(values, weights []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	type vw struct{ v, w float64 }
	pairs := make([]vw, len(values))
	totalW := 0.0
	for i := range values {
		pairs[i] = vw{values[i], weights[i]}
		totalW += weights[i]
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].v < pairs[j].v })

	cum := 0.0
	for _, p := range pairs {
		cum += p.w
		if cum >= totalW/2 {
			return p.v
		}
	}
	return pairs[len(pairs)-1].v
}

// EstimateLambdaTotal aggregates totals markets across all bookmakers into a
// single market-implied total-goals estimate using a weighted median.
// Weights are based on bookmaker recency (later last_update = higher weight).
func EstimateLambdaTotal(bookmakers []Bookmaker) (float64, bool) {
	var lambdas, weights []float64

	// Find the most recent update time to compute relative weights.
	var maxTime time.Time
	for _, bk := range bookmakers {
		for _, m := range bk.Markets {
			if m.Key == "totals" && m.LastUpdate.After(maxTime) {
				maxTime = m.LastUpdate
			}
		}
	}

	for _, bk := range bookmakers {
		for _, m := range bk.Markets {
			if m.Key != "totals" {
				continue
			}
			lambda, ok := lambdaFromTotalsOutcomes(m.Outcomes)
			if !ok {
				continue
			}
			// Weight = seconds since epoch difference from max; at least 1.
			age := maxTime.Sub(m.LastUpdate).Seconds()
			w := 1.0 / (1.0 + age) // fresher → higher weight
			lambdas = append(lambdas, lambda)
			weights = append(weights, w)
		}
	}

	if len(lambdas) == 0 {
		return 0, false
	}
	return weightedMedian(lambdas, weights), true
}

// SplitLambda splits lambdaTotal into home and away components using h2h
// market probabilities from the given bookmakers (Dixon-Coles simplified split).
// homeTeam and awayTeam are the raw team names from the OddsMatch record and
// are used to identify which h2h outcome belongs to which side.
func SplitLambda(lambdaTotal float64, bookmakers []Bookmaker, homeTeam, awayTeam string) (home, away float64) {
	var pHomeSum, pAwaySum float64
	count := 0
	for _, bk := range bookmakers {
		for _, m := range bk.Markets {
			if m.Key != "h2h" {
				continue
			}
			ph, pd, pa := h2hProbs(m.Outcomes, homeTeam, awayTeam)
			if ph <= 0 || pd <= 0 || pa <= 0 {
				continue
			}
			norm := ph + pd + pa
			pHomeSum += ph / norm
			pAwaySum += pa / norm
			count++
		}
	}
	if count == 0 || pAwaySum == 0 {
		return lambdaTotal / 2, lambdaTotal / 2
	}
	pHome := pHomeSum / float64(count)
	pAway := pAwaySum / float64(count)

	ratio := pHome / pAway
	home = lambdaTotal * ratio / (1 + ratio)
	away = lambdaTotal / (1 + ratio)
	return home, away
}

// h2hProbs returns raw (un-normalised) implied probabilities for a h2h market.
// It identifies home and away by matching outcome names against homeTeam and
// awayTeam rather than relying on position order.
func h2hProbs(outcomes []Outcome, homeTeam, awayTeam string) (pHome, pDraw, pAway float64) {
	for _, o := range outcomes {
		if o.Price <= 0 {
			continue
		}
		switch o.Name {
		case "Draw":
			pDraw = 1 / o.Price
		case homeTeam:
			pHome = 1 / o.Price
		case awayTeam:
			pAway = 1 / o.Price
		}
	}
	return
}

// CleanSheetPct returns the probability (0–100) of a clean sheet given the
// opponent's expected goals (lambda).
func CleanSheetPct(lambdaOpponent float64) float64 {
	return math.Exp(-lambdaOpponent) * 100
}

// AggregateBookmakers converts a slice of raw OddsMatch records into MatchOdds
// with computed lambda and clean sheet values.
func AggregateBookmakers(matches []OddsMatch) []MatchOdds {
	now := time.Now().UTC()
	result := make([]MatchOdds, 0, len(matches))
	for _, m := range matches {
		lambdaTotal, ok := EstimateLambdaTotal(m.Bookmakers)
		if !ok {
			continue
		}
		lHome, lAway := SplitLambda(lambdaTotal, m.Bookmakers, m.HomeTeam, m.AwayTeam)
		result = append(result, MatchOdds{
			OddsMatchID: m.ID,
			HomeTeam:    m.HomeTeam,
			AwayTeam:    m.AwayTeam,
			KickoffTime: m.CommenceTime,
			LambdaHome:  lHome,
			LambdaAway:  lAway,
			HomeCSPct:   CleanSheetPct(lAway),
			AwayCSPct:   CleanSheetPct(lHome),
			FetchedAt:   now,
		})
	}
	return result
}
