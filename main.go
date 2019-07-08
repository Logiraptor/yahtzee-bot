package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type ScoreLine int

const (
	Ones ScoreLine = iota
	Twos
	Threes
	Fours
	Fives
	Sixes
	ThreeOfAKind
	FourOfAKind
	SmallStraight
	LargeStraight
	FullHouse
	Yahtzee
	Chance
	ScoreLineCount
	ScoreLineMin = Ones
	ScoreLineMax = Chance

	UpperSectionMin   = Ones
	UpperSectionMax   = Sixes
	UpperSectionCount = 1 + (UpperSectionMax - UpperSectionMin)
	LowerSectionMin   = ThreeOfAKind
	LowerSectionMax   = Chance
	LowerSectionCount = 1 + (LowerSectionMax - LowerSectionMin)
)

var everyRoll []Roll

var scoreStats [ScoreLineCount]Stats

func init() {
	var allRolls = make(map[Roll]struct{})

	for a := One; a <= Six; a++ {
		for b := One; b <= Six; b++ {
			for c := One; c <= Six; c++ {
				for d := One; d <= Six; d++ {
					for e := One; e <= Six; e++ {
						var roll Roll
						roll[a]++
						roll[b]++
						roll[c]++
						roll[d]++
						roll[e]++
						allRolls[roll] = struct{}{}
					}
				}
			}
		}
	}

	for r, _ := range allRolls {
		everyRoll = append(everyRoll, r)
	}

	for s := ScoreLineMin; s < ScoreLineCount; s++ {
		scoreStats[s] = calculateStats(scorers[s], everyRoll)
	}
}

type Die uint8

//go:generate stringer -type=Die

const (
	One Die = iota
	Two
	Three
	Four
	Five
	Six
)

type Roll [6]uint8

func (r Roll) Ones() int {
	return int(r[0])
}

func (r Roll) Twos() int {
	return int(r[1])
}

func (r Roll) Threes() int {
	return int(r[2])
}

func (r Roll) Fours() int {
	return int(r[3])
}

func (r Roll) Fives() int {
	return int(r[4])
}

func (r Roll) Sixes() int {
	return int(r[5])
}

func (r Roll) Valid() bool {
	return r[0]+r[1]+r[2]+r[3]+r[4]+r[5] == 5
}

func (r Roll) Empty() bool {
	return r[0]+r[1]+r[2]+r[3]+r[4]+r[5] == 0
}

func (r Roll) String() string {

	if r.Empty() {
		return "empty"
	}

	if !r.Valid() {
		return "invld"
	}

	return strings.Repeat("⚀", int(r[0])) +
		strings.Repeat("⚁", int(r[1])) +
		strings.Repeat("⚂", int(r[2])) +
		strings.Repeat("⚃", int(r[3])) +
		strings.Repeat("⚄", int(r[4])) +
		strings.Repeat("⚅", int(r[5]))
}

type Rolls struct {
	FirstRoll  Roll
	SecondRoll Roll
	ThirdRoll  Roll
}

func (r Rolls) FinalRoll() Roll {
	if !r.ThirdRoll.Empty() {
		return r.ThirdRoll
	}
	if !r.SecondRoll.Empty() {
		return r.SecondRoll
	}
	return r.FirstRoll
}

type ScoreLines [ScoreLineCount]Roll

func (u ScoreLines) UpperSectionBaseScore() int {
	total := 0
	total += ScoreOnes(u[Ones])
	total += ScoreTwos(u[Twos])
	total += ScoreThrees(u[Threes])
	total += ScoreFours(u[Fours])
	total += ScoreFives(u[Fives])
	total += ScoreSixes(u[Sixes])
	return total
}

func (u ScoreLines) UpperSectionBonus() int {
	if u.UpperSectionBaseScore() > 63 {
		return 35
	}
	return 0
}

func (u ScoreLines) UpperSectionTotal() int {
	return u.UpperSectionBaseScore() + u.UpperSectionBonus()
}

func (u ScoreLines) LowerSectionTotal() int {
	total := 0
	total += ScoreThreeOfAKind(u[ThreeOfAKind])
	total += ScoreFourOfAKind(u[FourOfAKind])
	total += ScoreSmallStraight(u[SmallStraight])
	total += ScoreLargeStraight(u[LargeStraight])
	total += ScoreFullHouse(u[FullHouse])
	total += ScoreYahtzee(u[Yahtzee])
	total += ScoreChance(u[Chance])
	return total
}

type ScoreCardColumn struct {
	Name string
	ScoreLines
	BonusYahtzee [3]bool
}

func (s ScoreCardColumn) Total() int {
	return s.UpperSectionTotal() + s.LowerSectionTotal()
}

type ScoreCard struct {
	Columns []ScoreCardColumn
}

func (s ScoreCard) writeRow(wr io.Writer, name string, rowText func(ScoreCardColumn) string) error {
	_, err := fmt.Fprint(wr, name+"\t")
	if err != nil {
		return err
	}
	for _, col := range s.Columns {
		_, err = fmt.Fprint(wr, rowText(col)+"\t")
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(wr)
	return err
}

func (s ScoreCard) writeScore(wr io.Writer, name string, line ScoreLine) {
	s.writeRow(wr, name, func(col ScoreCardColumn) string {
		roll := col.ScoreLines[line]
		score := scorers[line](roll)
		stats := scoreStats[line]
		return fmt.Sprintf("%d (%s) (%.2f)", score, roll, stats.Percentile(score))
	})
}

func (s ScoreCard) writeHeader(wr io.Writer, name string) error {
	_, err := fmt.Fprintln(wr, name+"\t"+strings.Repeat("--\t", len(s.Columns)))
	return err
}

func (s ScoreCard) String() string {
	buf := new(bytes.Buffer)
	wr := tabwriter.NewWriter(buf, 0, 1, 1, ' ', tabwriter.AlignRight)
	s.writeHeader(wr, "Score")

	s.writeHeader(wr, "Upper Section")
	s.writeRow(wr, "Name", func(col ScoreCardColumn) string {
		return col.Name
	})
	s.writeScore(wr, "Ones", Ones)
	s.writeScore(wr, "Twos", Twos)
	s.writeScore(wr, "Threes", Threes)
	s.writeScore(wr, "Fours", Fours)
	s.writeScore(wr, "Fives", Fives)
	s.writeScore(wr, "Sixes", Sixes)
	s.writeRow(wr, "SubTotal", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.UpperSectionBaseScore())
	})
	s.writeRow(wr, "Bonus", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.UpperSectionBonus())
	})
	s.writeRow(wr, "Total Score", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.UpperSectionTotal())
	})

	s.writeHeader(wr, "Lower Section")
	s.writeScore(wr, "3 of a Kind", ThreeOfAKind)
	s.writeScore(wr, "4 of a Kind", FourOfAKind)
	s.writeScore(wr, "Full House", FullHouse)
	s.writeScore(wr, "SM Straight", SmallStraight)
	s.writeScore(wr, "LG Straight", LargeStraight)
	s.writeScore(wr, "Yahtzee", Yahtzee)
	s.writeScore(wr, "Chance", Chance)
	s.writeRow(wr, "SubTotal", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.LowerSectionTotal())
	})
	s.writeRow(wr, "Total", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Total())
	})
	// TODO: Bonus
	wr.Flush()
	return buf.String()
}

func randomRoll(r *rand.Rand) Roll {
	var data [5]byte
	r.Read(data[:])
	var roll Roll
	for _, d := range data {
		roll[uint8(d)%6]++
	}
	return roll
}

func makeRandomMove(getRoll func() Roll, column *ScoreCardColumn, rolls int) {
	roll := getRoll()

	for s := Ones; s < ScoreLineCount; s++ {
		if column.ScoreLines[s].Empty() {
			column.ScoreLines[s] = roll
			break
		}
	}
}

func makeGreedyMove(getRoll func() Roll, column *ScoreCardColumn, rolls int) {
	roll := getRoll()

	maxScore := -1
	maxIndex := -1
	for i := range column.ScoreLines {
		if candidateScore := scorers[i](roll); candidateScore > maxScore {
			maxScore = candidateScore
			maxIndex = i
		}
	}

	column.ScoreLines[maxIndex] = roll
}

func makeRareMove(getRoll func() Roll, column *ScoreCardColumn, rolls int) {
	roll := getRoll()

	maxScore := -1.0
	maxIndex := -1
	for i, existingRoll := range column.ScoreLines {
		if !existingRoll.Empty() {
			continue
		}

		score := scorers[i](roll)
		pct := scoreStats[i].Percentile(score)
		weight := pct * float64(score)
		if weight > maxScore {
			maxScore = weight
			maxIndex = i
		}
	}
	if maxScore == 0 && rolls < 3 {
		makeRareMove(getRoll, column, rolls+1)
		return
	}

	column.ScoreLines[maxIndex] = roll
}

func main() {
	var randomScoreCard = runGame("Random", makeRandomMove)
	var greedyScoreCard = runGame("Greedy", makeGreedyMove)
	var rareScoreCard = runGame("Rare", makeRareMove)

	var scoreCard = ScoreCard{
		Columns: []ScoreCardColumn{randomScoreCard, greedyScoreCard, rareScoreCard},
	}
	fmt.Println(scoreCard)
}

type AI func(getRoll func() Roll, column *ScoreCardColumn, rolls int)

func runGame(name string, ai AI) ScoreCardColumn {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	getRoll := func() Roll {
		return randomRoll(rnd)
	}
	col := ScoreCardColumn{Name: name}
	for i := 0; i < 13; i++ {
		ai(getRoll, &col, 1)
	}
	return col
}

type Stats struct {
	scores []int
}

func (s Stats) Min() int {
	return s.scores[0]
}

func (s Stats) Max() int {
	return s.scores[len(s.scores)-1]
}

func (s Stats) Median() int {
	medianIndex := len(s.scores) / 2
	return s.scores[medianIndex]
}

func (s Stats) Percentile(score int) float64 {
	for i, d := range s.scores {
		if d >= score {
			return float64(i) / float64(len(s.scores))
		}
	}
	return 1
}

func printStats(name string, scorer func(Roll) int, rolls []Roll) {
	stats := calculateStats(scorer, rolls)
	fmt.Println(name, stats.Min(), stats.Median(), stats.Max())
}

func calculateStats(scorer func(Roll) int, rolls []Roll) Stats {
	var scores []int
	for _, r := range rolls {
		scores = append(scores, scorer(r))
	}
	sort.Ints(scores)
	return Stats{scores}
}
