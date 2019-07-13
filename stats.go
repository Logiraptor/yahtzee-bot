package main

import (
	"sort"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/combin"
)

func optionsToKeepFrom(dice Roll, f func(Roll)) {
	var generated = make(map[Roll]struct{})
	var rec func(dice Roll)
	rec = func(dice Roll) {
		generated[dice] = struct{}{}
		for i, c := range dice {
			if c > 0 {
				var smaller = dice
				smaller[i]--
				if _, ok := generated[smaller]; ok {
					continue
				}
				f(smaller)
				rec(smaller)
			}
		}
	}
	rec(dice)
}

var rollDistributions = make(map[Roll]map[Roll]int)

func rollDistributionGiven(dice Roll) map[Roll]int {
	if dist, ok := rollDistributions[dice]; ok {
		return dist
	}
	var allRolls = make(map[Roll]int)

	availableRollsGiven(dice, func(roll Roll) {
		allRolls[roll]++
	})
	rollDistributions[dice] = allRolls
	return allRolls
}

func availableRollsGiven(dice Roll, f func(Roll)) {
	diceToGenerate := 5 - dice.Len()
	if diceToGenerate == 0 {
		f(dice)
		return
	}
	var data [][]float64
	for i := 0; i < diceToGenerate; i++ {
		data = append(data, []float64{One, Two, Three, Four, Five, Six})
	}
	possibleRolls := combin.Cartesian(nil, data)
	rows, _ := possibleRolls.Dims()

	for i := 0; i < rows; i++ {
		var roll = dice
		row := possibleRolls.RawRowView(i)
		for _, d := range row {
			roll[int(d)]++
		}
		f(roll)
	}
}

type Stats struct {
	scores []float64
	rolls  []Roll
	counts []float64
}

func (s Stats) Mean() float64 {
	return stat.Mean(s.scores, s.counts)
}

func (s Stats) Less(i, j int) bool {
	return s.scores[i] < s.scores[j]
}

func (s Stats) Swap(i, j int) {
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
	s.rolls[i], s.rolls[j] = s.rolls[j], s.rolls[i]
	s.counts[i], s.counts[j] = s.counts[j], s.counts[i]
}

func (s Stats) Len() int {
	return len(s.scores)
}

func (s Stats) Percentile(score int) float64 {
	scoreF := float64(score)
	return stat.CDF(scoreF, stat.Empirical, s.scores, s.counts)
}

func calculateStats(scorer func(Roll) int, rolls map[Roll]int) Stats {
	stats := Stats{}
	for r, c := range rolls {
		stats.scores = append(stats.scores, float64(scorer(r)))
		stats.rolls = append(stats.rolls, r)
		stats.counts = append(stats.counts, float64(c))
	}
	sort.Sort(stats)
	return stats
}
