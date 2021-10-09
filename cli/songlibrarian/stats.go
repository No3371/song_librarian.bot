package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type stats struct {
	StartAt			   time.Time
	MessageEvents      uint64
	MessageBuffered    uint64
	FirstFetchEmbeds0  uint64
	SecondFetchEmbeds0 uint64
	ThirdFetchEmbeds0  uint64
	FetchedAndAnalyzed  uint64
	AnalyzedEmbeds     uint64
	UrlRegexMatched    uint64
	SkippedDuplicate    uint64
	Pended             uint64
	Redirected         uint64
}

var statSession *stats

func (s *stats) Print() {
	j, err := json.MarshalIndent(*s, "", "  ")
	if err != nil {
		j = []byte("Failed to marshal")
	}
	fmt.Printf("Has been running for %s...", time.Now().Sub(statSession.StartAt).Round(time.Second))
	fmt.Printf("\n[STATS] Redirect rate: %0.2f%%\n%s\n", 100*(float64(statSession.Redirected)/float64(statSession.AnalyzedEmbeds)), string(j))
}
