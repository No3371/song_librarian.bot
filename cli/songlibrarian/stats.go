package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"No3371.github.com/song_librarian.bot/redirect"
)

type stats struct {
	StartAt             time.Time
	MessageEvents       uint64
	UnSubbedSkips       uint64
	MessageBuffered     uint64
	Unlinked            uint64
	SkippedLinks        uint64
	FirstFetchEmbeds0   uint64
	SecondFetchEmbeds0  uint64
	ThirdFetchEmbeds0   uint64
	BoundChannelMessage uint64
	FetchedAndAnalyzed  uint64
	AnalyzedEmbeds      uint64
	UrlRegexMatched     uint64
	SkippedDuplicate    uint64
	PendedSpoilerFlag	uint64
	Pended              uint64
	Redirected          uint64
	GuessRight          uint64
	GueseWrongType      uint64
}

var statSession *stats

func (s *stats) Print() {
	j, err := json.MarshalIndent(*s, "", "  ")
	if err != nil {
		j = []byte("Failed to marshal")
	}
	passed := time.Since(statSession.StartAt).Round(time.Second)
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("\nUptime: %s", passed))
	sb.WriteString(fmt.Sprintf("\nAverage msg/hour: %0.2f", float64(statSession.BoundChannelMessage) / math.Max(1, passed.Hours())))
	sb.WriteString(fmt.Sprintf("\nAverage msg/hour: %0.2f", float64(statSession.BoundChannelMessage) / math.Max(1, passed.Hours())))
	sb.WriteString(fmt.Sprintf("\nGuess correctness: %0.2f%%", 100*(float64(statSession.GuessRight)/float64(statSession.Pended))))
	sb.WriteString(fmt.Sprintf("\nGuess wrongness: %0.2f%%\n", 100*(float64(statSession.GueseWrongType)/float64(statSession.Pended))))
	sb.WriteString(string(j))
	fmt.Print(sb.String())
}

type badGuessRecord struct {
	title string
	guess redirect.RedirectType
	result redirect.RedirectType
}

func init () {
	badGuesses = [1024]badGuessRecord{}
}

var badGuesses [1024]badGuessRecord
var badGuessIndex int

func noteGuessedWrong (r badGuessRecord) {
	badGuesses[badGuessIndex] = r
	badGuessIndex++
	if badGuessIndex >= 1023 {
		badGuessIndex = 0
	}
}

func printBadGuesses () {
	sb := &strings.Builder {}
	for i := 0; i < 1024; i++ {
		r := badGuesses[i]
		if r.title == "" {
			continue
		}

		if r.result != redirect.None {
			sb.WriteString(fmt.Sprintf("\n  %s -> %s / %s", redirect.RedirectTypetoString(r.guess), redirect.RedirectTypetoString(r.result), r.title))
		} else {
			sb.WriteString(fmt.Sprintf("\n%s -> %s / %s", redirect.RedirectTypetoString(r.guess), redirect.RedirectTypetoString(r.result), r.title))
		}
	}
	fmt.Println(sb.String())
}