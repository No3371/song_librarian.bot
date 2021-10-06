package main

type stats struct {
	messageEvents uint64
	messageBuffered uint64
	firstFetchEmbeds0 uint64
	secondFetchEmbeds0 uint64
	thirdFetchEmbeds0 uint64
	analyzedEmbeds uint64
	urlRegexMatched uint64
	pended uint64
	redirected uint64
}

var statSession *stats