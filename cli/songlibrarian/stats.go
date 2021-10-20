package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"No3371.github.com/song_librarian.bot/redirect"
)

type stats struct {
	StartAt             time.Time
	MessageEvents       uint64
	UnSubbedSkips       uint64
	MessageBuffered     uint64
	FirstFetchEmbeds0   uint64
	SecondFetchEmbeds0  uint64
	ThirdFetchEmbeds0   uint64
	BoundChannelMessage uint64
	FetchedAndAnalyzed  uint64
	AnalyzedEmbeds      uint64
	UrlRegexMatched     uint64
	SkippedDuplicate    uint64
	Pended              uint64
	Redirected          uint64
	GuessRight          uint64
}

var statSession *stats

func (s *stats) Print() {
	j, err := json.MarshalIndent(*s, "", "  ")
	if err != nil {
		j = []byte("Failed to marshal")
	}
	fmt.Printf("Has been running for %s...", time.Since(statSession.StartAt).Round(time.Second))
	fmt.Printf("\n[STATS] Redirect rate: %0.2f%%\n%s\n", 100*(float64(statSession.Redirected)/float64(statSession.AnalyzedEmbeds)), string(j))
}

type badGuessRecord struct {
	title string
	guess redirect.RedirectType
	result redirect.RedirectType
}

func init () {
	badGuesses = [256]badGuessRecord{}
}

var badGuesses [256]badGuessRecord
var badGuessIndex int

func noteGuessedWrong (r badGuessRecord) {
	badGuesses[badGuessIndex] = r
	badGuessIndex++
	if badGuessIndex >= 255 {
		badGuessIndex = 0
	}
}

func printBadGuesses () {
	sb := &strings.Builder {}
	for i := 0; i < 256; i++ {
		r := badGuesses[i]
		if r.title == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("\n%s -> %s / %s", redirect.RedirectTypetoString(r.guess), redirect.RedirectTypetoString(r.result), r.title))
	}
	fmt.Println(sb.String())
}