package main

import (
	"sync"
	"time"

	"No3371.github.com/song_librarian.bot/logger"
)

const MEM_SIZE = 64

var redirectedMap map[string]memory
var redirectedUrls [MEM_SIZE]string
var memPointer int
var lock *sync.RWMutex

type memory struct {
	index int
	markedAt time.Time
}



func init() {
	redirectedMap = make(map[string]memory, MEM_SIZE)
	redirectedUrls = [MEM_SIZE]string {}
	lock = new(sync.RWMutex)
}

func isDuplicate(url string) (last time.Time, redirected bool) {
	lock.RLock()
	defer lock.RUnlock()
	item, redirected := redirectedMap[url]
	last = item.markedAt
	return
}

func markRedirected(url string) (err error) {
	defer func () {
		if p := recover(); p != nil {
			if err != nil {
				logger.Logger.Errorf("Error when marking redirected: %v", err)
			}
			err = p.(error)
			logger.Logger.Errorf("Error when marking redirected: %v", err)
		}
	} ()

	lock.Lock()
	defer lock.Unlock()
	if m, exists := redirectedMap[url]; exists {
		redirectedMap[url] =  memory{ index: m.index, markedAt: time.Now() } // This happens for unmarke item bursted and mem that is older then 24hr
		return
	}

	if len(redirectedUrls[memPointer]) > 0 {
		delete(redirectedMap, redirectedUrls[memPointer])
	}
	redirectedUrls[memPointer] = url
	redirectedMap[url] = memory{ index: memPointer, markedAt: time.Now() }
	memPointer++
	if memPointer >= MEM_SIZE {
		memPointer = 0
	}
	return
}

func forget (url string) {
	lock.Lock()
	defer lock.Unlock()
	if mem, exists := redirectedMap[url]; exists {
		return
	} else {
		redirectedUrls[mem.index] = ""
		delete(redirectedMap, url)		
	}
}