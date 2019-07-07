package main

type OnesScorer struct {
	Roll
}

func (o OnesScorer) Score() int {
	return o.Ones()
}

type TwosScorer struct {
	Roll
}

func (t TwosScorer) Score() int {
	return t.Twos() * 2
}

type ThreesScorer struct {
	Roll
}

func (t ThreesScorer) Score() int {
	return t.Threes() * 3
}

type FoursScorer struct {
	Roll
}

func (f FoursScorer) Score() int {
	return f.Fours() * 4
}

type FivesScorer struct {
	Roll
}

func (f FivesScorer) Score() int {
	return f.Fives() * 5
}

type SixesScorer struct {
	Roll
}

func (s SixesScorer) Score() int {
	return s.Sixes() * 6
}

type ThreeOfAKindScorer struct {
	Roll
}

func (t ThreeOfAKindScorer) Score() int {
	valid := false
	sum := 0
	for _, d := range t.Roll {
		sum += int(d)
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

type FourOfAKindScorer struct {
	Roll
}

func (f FourOfAKindScorer) Score() int {
	valid := false
	sum := 0
	for _, d := range f.Roll {
		sum += int(d)
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

type SmallStraightScorer struct {
	Roll
}

func (s SmallStraightScorer) Score() int {
	runLength := 0
	for _, d := range s.Roll {
		if d == 0 {
			runLength = 0
		} else {
			runLength++
		}
	}
	if runLength >= 4 {
		// TODO: Check score for small straight
		return 35
	}
	return 0
}

type LargeStraightScorer struct {
	Roll
}

func (l LargeStraightScorer) Score() int {
	runLength := 0
	for _, d := range l.Roll {
		if d == 0 {
			runLength = 0
		} else {
			runLength++
		}
	}
	if runLength >= 5 {
		// TODO: Check score for large straight
		return 40
	}
	return 0
}

type FlushScorer struct {
	Roll
}

func (f FlushScorer) Score() int {
	foundPair := false
	foundTriple := false
	for _, d := range f.Roll {
		if d == 2 {
			foundPair = true
		} else if d == 3 {
			foundTriple = true
		}
	}
	if foundPair && foundTriple {
		// TODO: Check score for flush
		return 40
	}
	return 0
}

type YahtzeeScorer struct {
	Roll
}

func (y YahtzeeScorer) Score() int {
	for _, d := range y.Roll {
		if d == 5 {
			// TODO: check score for yahtzee
			return 50
		}
	}
	return 0
}

type ChanceScorer struct {
	Roll
}

func (c ChanceScorer) Score() int {
	sum := 0
	for _, d := range c.Roll {
		sum += int(d)
	}
	return sum
}
