package main

import (
	"sync"
	"time"

)


type memState int 

const (
	None memState = iota
	Shared 
	Pended
	Redirected
	Cancelled
	CancelledWithError
)

const MEM_SIZE = 256

var urlMap map[string]*memory
var urls [MEM_SIZE]string
var memPointer int
var lock *sync.RWMutex

type memory struct {
	index int
	latestState memState
	cancelledAt time.Time
	redirectedAt time.Time
	sharedAt time.Time
	pendedAt time.Time
}



func init() {
	urlMap = make(map[string]*memory, MEM_SIZE)
	urls = [MEM_SIZE]string {}
	lock = new(sync.RWMutex)
}

func getLastState (url string) (state memState, ts time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	var inMem bool
	var mem *memory
	if mem, inMem = urlMap[url]; inMem {
		state = mem.latestState
		ts = mem.sharedAt
	} else {
		state = None
	}
	return

}

func getLastShared (url string) (state memState, ts time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	var inMem bool
	var mem *memory
	if mem, inMem = urlMap[url]; inMem {
		state = mem.latestState
		ts = mem.sharedAt
	} else {
		state = None
	}
	return
}

func getLastResult (url string) (latestState memState, ts time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	var inMem bool
	var mem *memory
	if mem, inMem = urlMap[url]; inMem {
		latestState = mem.latestState
		if latestState == Redirected {
			ts = mem.redirectedAt
		} else if latestState == Cancelled {
			ts = mem.cancelledAt
		}
	} else {
		latestState = None
	}
	return
}

func getLastRedirected (url string) (latestState memState, ts time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	var inMem bool
	var mem *memory
	if mem, inMem = urlMap[url]; inMem {
		latestState = mem.latestState
		ts = mem.redirectedAt
	} else {
		latestState = None
	}
	return
}

func getLastCancelled (url string) (latestState memState, ts time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	var inMem bool
	var mem *memory
	if mem, inMem = urlMap[url]; inMem {
		latestState = mem.latestState
		ts = mem.cancelledAt
	} else {
		latestState = None
	}
	return
}

func getLastPended (url string) (latestState memState, ts time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	var inMem bool
	var mem *memory
	if mem, inMem = urlMap[url]; inMem {
		latestState = mem.latestState
		ts = mem.pendedAt
	} else {
		latestState = None
	}
	return
}

func memorizePended (url string) (err error) {
	lock.Lock()
	defer lock.Unlock()

	if m, exists := urlMap[url]; exists {
		m.pendedAt = time.Now() // This happens for unmarke item bursted and mem that is older then 24hr
		m.latestState = Pended
		return
	} else {
		if len(urls[memPointer]) > 0 {
			delete(urlMap, urls[memPointer])
		}
		urls[memPointer] = url
		urlMap[url] = &memory{ index: memPointer, pendedAt: time.Now(), latestState: Pended }
		memPointer++
		if memPointer >= MEM_SIZE {
			memPointer = 0
		}
	}
	return
}

func memorizeShared (url string) (err error) {
	lock.Lock()
	defer lock.Unlock()

	if m, exists := urlMap[url]; exists {
		m.sharedAt = time.Now() // This happens for unmarke item bursted and mem that is older then 24hr
		m.latestState = Shared
		return
	} else {
		if len(urls[memPointer]) > 0 {
			delete(urlMap, urls[memPointer])
		}
		urls[memPointer] = url
		urlMap[url] = &memory{ index: memPointer, sharedAt: time.Now(), latestState: Shared }
		memPointer++
		if memPointer >= MEM_SIZE {
			memPointer = 0
		}
	}
	return
}


func memorizeResult (url string, state memState) (err error) {
	lock.Lock()
	defer lock.Unlock()

	if m, exists := urlMap[url]; exists {
		if state == Redirected {
			m.redirectedAt = time.Now()
		} else if state == Cancelled {
			m.cancelledAt = time.Now()
		}
		m.latestState = state
		return
	} else {
		if len(urls[memPointer]) > 0 {
			delete(urlMap, urls[memPointer])
		}
		urls[memPointer] = url
		if state == Redirected {
			urlMap[url] = &memory{ index: memPointer, redirectedAt: time.Now(), latestState: state }
		} else if state == Cancelled {
			urlMap[url] = &memory{ index: memPointer, cancelledAt: time.Now(), latestState: state }
		}
		memPointer++
		if memPointer >= MEM_SIZE {
			memPointer = 0
		}
	}
	return
}

func forget (url string) {
	lock.Lock()
	defer lock.Unlock()
	if mem, exists := urlMap[url]; exists {
		return
	} else {
		urls[mem.index] = ""
		delete(urlMap, url)		
	}
}

func memStateToString (ms memState) string {
	switch ms {
	case Shared:
		return "Shared"
	case Pended:
		return "Pended"
	case Redirected:
		return "Redirected"
	case Cancelled:
		return "Cancelled"
	default:
		return "???"
	}
}