package main

import (
	"encoding/json"
	"fmt"
)

type stats struct {
	messageEvents      uint64
	messageBuffered    uint64
	firstFetchEmbeds0  uint64
	secondFetchEmbeds0 uint64
	thirdFetchEmbeds0  uint64
	analyzedEmbeds     uint64
	urlRegexMatched    uint64
	pended             uint64
	redirected         uint64
}

var statSession *stats

func (s *stats) Print() {
	j, err := json.Marshal(*s)
	if err != nil {
		j = []byte("Failed to unmarshal")
	}
	fmt.Printf("\n[STATS]\n%s\n", string(j))
}