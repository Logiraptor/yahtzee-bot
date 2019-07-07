package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

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

type ScoreCardUpperSection struct {
	Ones   OnesScorer
	Twos   TwosScorer
	Threes ThreesScorer
	Fours  FoursScorer
	Fives  FivesScorer
	Sixes  SixesScorer
}

func (u ScoreCardUpperSection) UpperSectionBaseScore() int {
	total := 0
	total += u.Ones.Score()
	total += u.Twos.Score()
	total += u.Threes.Score()
	total += u.Fours.Score()
	total += u.Fives.Score()
	total += u.Sixes.Score()
	return total
}

func (u ScoreCardUpperSection) UpperSectionBonus() int {
	if u.UpperSectionBaseScore() > 63 {
		// TODO: check upper section bonus score
		return 35
	}
	return 0
}

func (u ScoreCardUpperSection) UpperSectionTotal() int {
	return u.UpperSectionBaseScore() + u.UpperSectionBonus()
}

type ScoreCardLowerSection struct {
	ThreeOfAKind  ThreeOfAKindScorer
	FourOfAKind   FourOfAKindScorer
	SmallStraight SmallStraightScorer
	LargeStraight LargeStraightScorer
	Flush         FlushScorer
	Yahtzee       YahtzeeScorer
	Chance        ChanceScorer
}

type ScoreCardColumn struct {
	Name string
	ScoreCardUpperSection
	ScoreCardLowerSection
	BonusYahtzee [3]bool
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

func (s ScoreCard) writeHeader(wr io.Writer, name string) error {
	_, err := fmt.Fprintln(wr, name+"\t")
	return err
}

func (s ScoreCard) String() string {
	buf := new(bytes.Buffer)
	wr := tabwriter.NewWriter(buf, 0, 1, 1, ' ', 0)
	s.writeHeader(wr, "Score")
	s.writeHeader(wr, "--------")

	s.writeHeader(wr, "Upper Section")

	s.writeRow(wr, "Name", func(col ScoreCardColumn) string {
		return col.Name
	})

	s.writeRow(wr, "Ones", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Ones.Score())
	})
	s.writeRow(wr, "Twos", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Twos.Score())
	})
	s.writeRow(wr, "Threes", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Threes.Score())
	})
	s.writeRow(wr, "Fours", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Fours.Score())
	})
	s.writeRow(wr, "Fives", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Fives.Score())
	})
	s.writeRow(wr, "Sixes", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.Sixes.Score())
	})

	s.writeRow(wr, "SubTotal", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.UpperSectionBaseScore())
	})

	s.writeRow(wr, "Bonus", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.UpperSectionBonus())
	})

	s.writeRow(wr, "Total", func(col ScoreCardColumn) string {
		return strconv.Itoa(col.UpperSectionTotal())
	})

	s.writeHeader(wr, "Lower Section")

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

func main() {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	var scoreCard = ScoreCard{
		Columns: []ScoreCardColumn{
			{
				Name: "P1",
				ScoreCardUpperSection: ScoreCardUpperSection{
					Ones:   OnesScorer{randomRoll(rnd)},
					Twos:   TwosScorer{randomRoll(rnd)},
					Threes: ThreesScorer{randomRoll(rnd)},
					Fours:  FoursScorer{randomRoll(rnd)},
					Fives:  FivesScorer{randomRoll(rnd)},
					Sixes:  SixesScorer{randomRoll(rnd)},
				},
			},
			{
				Name: "P2",
			},
		},
	}

	fmt.Println(scoreCard)
}
