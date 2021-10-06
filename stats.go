package main

import (
	"encoding/json"
	"fmt"
)

type stats struct {
	MessageEvents      uint64
	MessageBuffered    uint64
	FirstFetchEmbeds0  uint64
	SecondFetchEmbeds0 uint64
	ThirdFetchEmbeds0  uint64
	AnalyzedEmbeds     uint64
	UrlRegexMatched    uint64
	Pended             uint64
	Redirected         uint64
}

var statSession *stats

func (s *stats) Print() {
	j, err := json.Marshal(*s)
	if err != nil {
		j = []byte("Failed to marshal")
	}
	fmt.Printf("\n[STATS] Redirect rate: %f , \n%s\n", float64(statSession.AnalyzedEmbeds)/float64(statSession.MessageEvents), string(j))
}