package main

type ScoreFunc func(r Roll) int

var scorers = [ScoreLineCount]ScoreFunc{
	Ones:          ScoreOnes,
	Twos:          ScoreTwos,
	Threes:        ScoreThrees,
	Fours:         ScoreFours,
	Fives:         ScoreFives,
	Sixes:         ScoreSixes,
	ThreeOfAKind:  ScoreThreeOfAKind,
	FourOfAKind:   ScoreFourOfAKind,
	SmallStraight: ScoreSmallStraight,
	LargeStraight: ScoreLargeStraight,
	FullHouse:     ScoreFullHouse,
	Yahtzee:       ScoreYahtzee,
	Chance:        ScoreChance,
}

func ScoreOnes(o Roll) int {
	return o.Ones()
}

func ScoreTwos(t Roll) int {
	return t.Twos() * 2
}

func ScoreThrees(t Roll) int {
	return t.Threes() * 3
}

func ScoreFours(f Roll) int {
	return f.Fours() * 4
}

func ScoreFives(f Roll) int {
	return f.Fives() * 5
}

func ScoreSixes(s Roll) int {
	return s.Sixes() * 6
}

func ScoreThreeOfAKind(t Roll) int {
	valid := false
	sum := 0
	for i, d := range t {
		sum += int(d) * (i + 1)
		if d >= 3 {
			valid = true
		}
	}

	if !valid {
		// Could not find three of a kind, this must be a 0
		return 0
	}
	return sum
}

func ScoreFourOfAKind(f Roll) int {
	valid := false
	sum := 0
	for i, d := range f {
		sum += int(d) * (i + 1)
		if d >= 4 {
			valid = true
		}
	}

	if !valid {
		// Could not find three of a kind, this must be a 0
		return 0
	}
	return sum
}

func ScoreSmallStraight(s Roll) int {
	runLength := 0
	for _, d := range s {
		if d == 0 {
			runLength = 0
		} else {
			runLength++
		}
	}
	if runLength >= 4 {
		return 30
	}
	return 0
}

func ScoreLargeStraight(l Roll) int {
	runLength := 0
	for _, d := range l {
		if d == 0 {
			runLength = 0
		} else {
			runLength++
		}
	}
	if runLength >= 5 {
		return 40
	}
	return 0
}

func ScoreFullHouse(f Roll) int {
	foundPair := false
	foundTriple := false
	for _, d := range f {
		if d == 2 {
			foundPair = true
		} else if d == 3 {
			foundTriple = true
		}
	}
	if foundPair && foundTriple {
		return 25
	}
	return 0
}

func ScoreYahtzee(y Roll) int {
	for _, d := range y {
		if d == 5 {
			return 50
		}
	}
	return 0
}

func ScoreChance(c Roll) int {
	sum := 0
	for i, d := range c {
		sum += int(d) * (i + 1)
	}
	return sum
}
