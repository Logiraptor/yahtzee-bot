package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

//go:generate stringer -type=ScoreLine

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

const (
	One = iota
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

func (r Roll) Len() int {
	return int(r[0] + r[1] + r[2] + r[3] + r[4] + r[5])
}

func (r Roll) Valid() bool {
	return r.Len() == 5
}

func (r Roll) Empty() bool {
	return r.Len() == 0
}

func (r Roll) String() string {

	if r.Empty() {
		return "empty"
	}

	return strings.Repeat("⚀", int(r[0])) +
		strings.Repeat("⚁", int(r[1])) +
		strings.Repeat("⚂", int(r[2])) +
		strings.Repeat("⚃", int(r[3])) +
		strings.Repeat("⚄", int(r[4])) +
		strings.Repeat("⚅", int(r[5]))
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
		return fmt.Sprintf("%d (%s)", score, roll)
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

func makeRandomMove(getRoll func(Roll) Roll, column *ScoreCardColumn, rolls int) {
	roll := getRoll(Roll{})

	for s := Ones; s < ScoreLineCount; s++ {
		if column.ScoreLines[s].Empty() {
			column.ScoreLines[s] = roll
			break
		}
	}
}

func makeGreedyMove(getRoll func(Roll) Roll, column *ScoreCardColumn, rolls int) {
	roll := getRoll(Roll{})

	maxScore := -1
	maxIndex := -1
	for i, existingRoll := range column.ScoreLines {
		if !existingRoll.Empty() {
			continue
		}
		if candidateScore := scorers[i](roll); candidateScore > maxScore {
			maxScore = candidateScore
			maxIndex = i
		}
	}

	column.ScoreLines[maxIndex] = roll
}

func makeRareMove(getRoll func(Roll) Roll, column *ScoreCardColumn, rolls int) {
	roll := getRoll(Roll{})

	maxScore := -1.0
	maxIndex := -1
	for i, existingRoll := range column.ScoreLines {
		if !existingRoll.Empty() {
			continue
		}

		score := scorers[i](roll)
		stats := calculateStats(scorers[i], rollDistributionGiven(Roll{}))
		pct := stats.Percentile(score)
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

func makeGreedyExpectedValueMove(getRoll func(Roll) Roll, column *ScoreCardColumn, rolls int) {
	var currentRoll = getRoll(Roll{})
	var currentLine ScoreLine
	var expectedScore float64

	for ; rolls < 3; rolls++ {
		var maxAlternative = -1.0
		var maxRoll Roll
		optionsToKeepFrom(currentRoll, func(keep Roll) {
			//   evaluate expected value
			candidateScore, _ := bestScoreLine(keep, column)
			if candidateScore > maxAlternative {
				maxAlternative = candidateScore
				maxRoll = keep
			}
		})

		expectedScore, currentLine = bestScoreLine(currentRoll, column)
		if maxAlternative > expectedScore {
			// Rolling is on average better, so roll some dice
			currentRoll = getRoll(maxRoll)
			expectedScore, currentLine = bestScoreLine(currentRoll, column)
		} else {
			break
		}
	}

	// We either found an above average roll, or we ran out of attempts
	column.ScoreLines[currentLine] = currentRoll
}

func bestScoreLine(start Roll, column *ScoreCardColumn) (float64, ScoreLine) {
	maxMean := -1.0
	maxIndex := -1
	for i, existingRoll := range column.ScoreLines {
		if !existingRoll.Empty() {
			continue
		}
		value := expectedValue(start, ScoreLine(i))
		if value > maxMean {
			maxIndex = i
			maxMean = value
		}
	}

	return maxMean, ScoreLine(maxIndex)
}

var expectedValues = make(map[Roll][ScoreLineCount]float64)

func expectedValue(start Roll, scoreLine ScoreLine) float64 {
	if results, ok := expectedValues[start]; ok {
		return results[scoreLine]
	}

	possibleRollDistribution := rollDistributionGiven(start)
	var results [ScoreLineCount]float64
	for s := ScoreLineMin; s < ScoreLineCount; s++ {
		stats := calculateStats(scorers[s], possibleRollDistribution)
		results[s] = stats.Mean()
	}

	expectedValues[start] = results
	return results[scoreLine]
}

func gemerateSummaryPlot() {
	ais := map[string]AI{
		"Random":      makeRandomMove,
		"Greedy":      makeGreedyMove,
		"Rare":        makeRareMove,
		"Greedy Mean": makeGreedyExpectedValueMove,
	}
	results := runIterations(1000, ais)
	allValues := make(map[string]plotter.Values)
	for name, scores := range results {
		values := make(plotter.Values, len(scores))
		for i, score := range scores {
			values[i] = float64(score)
		}
		allValues[name] = values
	}
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "AIs"
	p.Y.Label.Text = "Score"

	// Make boxes for our data and add them to the plot.
	w := vg.Points(20)
	var i = 0.0
	var names = []string{}
	for name, vals := range allValues {
		names = append(names, name)
		b0, err := plotter.NewBoxPlot(w, i, vals)
		if err != nil {
			panic(err)
		}
		p.Add(b0)
		i++
	}
	p.NominalX(names...)
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "boxplot.png"); err != nil {
		panic(err)
	}
}

func main() {
	gemerateSummaryPlot()
}

func generateRollsFrom(start Roll, len int, min int, f func(Roll)) {
	if len == 1 {
		for i := min; i <= Six; i++ {
			var roll = start
			roll[i]++
			f(roll)
		}
		return
	}

	if min == Six {
		var roll = start
		roll[min] = uint8(len)
		f(roll)
		return
	}

	var sub = start
	sub[min]++
	// generate rolls containing a min
	generateRollsFrom(sub, len-1, min, f)

	// generate rolls not containing a min
	generateRollsFrom(start, len, min+1, f)
}

func runIterations(count int, ais map[string]AI) map[string][]int {
	var output = make(map[string][]int)
	for i := 0; i < count; i++ {
		results := runIteration(int64(i), ais)
		for name, score := range results {
			output[name] = append(output[name], score)
		}
	}
	return output
}

func runIteration(seed int64, ais map[string]AI) map[string]int {
	var output = make(map[string]int)
	for name, ai := range ais {
		column := runGame(seed, name, ai)
		output[name] = column.Total()
	}
	return output
}

type diceRoller struct {
	rand *rand.Rand
	mut  *sync.Mutex
	dice []uint8
	pos  int
}

func NewDiceRoller(rand *rand.Rand) *diceRoller {
	return &diceRoller{
		rand: rand,
		mut:  new(sync.Mutex),
	}
}

func (d *diceRoller) ensureLen(l int) {
	d.mut.Lock()
	defer d.mut.Unlock()
	if len(d.dice)-d.pos > l {
		return
	}

	var buf [16]byte
	d.rand.Read(buf[:])
	d.dice = append(d.dice, buf[:]...)
}

func (d *diceRoller) read(buf []byte) {
	d.ensureLen(len(buf))

	copy(buf, d.dice[d.pos:])
	d.pos += len(buf)
}

func (d *diceRoller) randomRoll(start Roll) Roll {
	var data = make([]byte, 5-start.Len())
	d.read(data[:])
	for _, d := range data {
		start[uint8(d)%6]++
	}
	return start
}

func (d *diceRoller) withRoll(start Roll, f func(Roll)) {
	var subRoll = d.randomRoll(start)
	f(subRoll)
	d.pos -= 5 - start.Len()
}

type AI func(getRoll func(Roll) Roll, column *ScoreCardColumn, rolls int)

func runGame(seed int64, name string, ai AI) ScoreCardColumn {
	roller := NewDiceRoller(rand.New(rand.NewSource(seed)))
	getRoll := func(start Roll) Roll {
		return roller.randomRoll(start)
	}
	col := ScoreCardColumn{Name: name}
	for i := 0; i < 13; i++ {
		ai(getRoll, &col, 1)
	}
	return col
}

type optimalScoreAI struct {
	roller *diceRoller
	scores ScoreLines
}

func findOptimalScore(seed int64, name string) ScoreCardColumn {
	roller := NewDiceRoller(rand.New(rand.NewSource(seed)))

	o := optimalScoreAI{roller: roller}
	o.findOptimalScoreForTurn(0)

	return ScoreCardColumn{}
}

func (o *optimalScoreAI) findOptimalScoreForTurn(turn int) {
	if turn >= 13 {
		return
	}

	o.findOptimalRollForTurn(Roll{}, turn)
}

func (o *optimalScoreAI) findOptimalRollForTurn(base Roll, rolls int) {
	if rolls >= 3 {
		return
	}

	o.roller.withRoll(base, func(baseRoll Roll) {
		fmt.Println(rolls, baseRoll)

		// for all possible reroll attempts
		optionsToKeepFrom(baseRoll, func(keep Roll) {
			o.findOptimalRollForTurn(keep, rolls+1)
		})
	})
}
