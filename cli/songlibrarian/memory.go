package main

// import (
// 	"encoding/json"
// 	"sync"
// 	"time"

// 	"No3371.github.com/song_librarian.bot/logger"
// )


// type memState int 

// const (
// 	None memState = iota
// 	Shared 
// 	Pended
// 	Redirected
// 	Cancelled
// 	CancelledWithError
// )

// const MEM_SIZE = 2048

// var urlMap map[string]*memory
// var urls [MEM_SIZE]string
// var memPointer int
// var lock *sync.RWMutex

// type memory struct {
// 	Url string
// 	LatestState memState
// 	CancelledAt time.Time
// 	UnexpectedUnfinishedAt time.Time
// 	RedirectedAt time.Time
// 	SharedAt time.Time
// 	PendedAt time.Time
// }



// func init() {
// 	urlMap = make(map[string]*memory, MEM_SIZE)
// 	urls = [MEM_SIZE]string {}
// 	lock = new(sync.RWMutex)
// }

// func saveAllMemories () {
// 	var count = 0
// 	var job = func (i uint) {
// 		if urls[i] == "" || urlMap[urls[i]] == nil {
// 			return
// 		}

// 		m := urlMap[urls[i]]
// 		j, err := json.Marshal(m)
// 		if err != nil {
// 			logger.Logger.Errorf("Failed to marshal memory: %s", err)
// 			return
// 		}
// 		err = sp.SaveMem(uint64(count), string(j))
// 		if err != nil {
// 			logger.Logger.Errorf("Failed to save memory: %s", err)
// 			return
// 		}

// 		count++
// 	}
// 	for i := memPointer; i > 0; i-- {
// 		job(uint(i))
// 	}
	
// 	for i := MEM_SIZE - 1; i > memPointer; i-- {
// 		job(uint(i))
// 	}
// }

// func loadAllMemories () {
// 	datas, err := sp.LoadMemAll()
// 	if err != nil {
// 		logger.Logger.Errorf("Failed to load memories: %s", err)
// 		return
// 	}

// 	for _, data := range datas {
// 		var m *memory
// 		err = json.Unmarshal([]byte(data), m)
// 		if err != nil {
// 			logger.Logger.Errorf("Failed to unmarshal memory: %s", err)
// 			continue
// 		}
// 	}
// }

// func getLastState (url string) (state memState, ts time.Time) {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	var inMem bool
// 	var mem *memory
// 	if mem, inMem = urlMap[url]; inMem {
// 		state = mem.LatestState
// 		ts = mem.SharedAt
// 	} else {
// 		state = None
// 	}
// 	return

// }

// func getLastShared (url string) (state memState, ts time.Time) {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	var inMem bool
// 	var mem *memory
// 	if mem, inMem = urlMap[url]; inMem {
// 		state = mem.LatestState
// 		ts = mem.SharedAt
// 	} else {
// 		state = None
// 	}
// 	return
// }

// func getLastResult (url string) (latestState memState, ts time.Time) {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	var inMem bool
// 	var mem *memory
// 	if mem, inMem = urlMap[url]; inMem {
// 		latestState = mem.LatestState
// 		if latestState == Redirected {
// 			ts = mem.RedirectedAt
// 		} else if latestState == Cancelled {
// 			ts = mem.CancelledAt
// 		}
// 	} else {
// 		latestState = None
// 	}
// 	return
// }

// func getLastRedirected (url string) (latestState memState, ts time.Time) {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	var inMem bool
// 	var mem *memory
// 	if mem, inMem = urlMap[url]; inMem {
// 		latestState = mem.LatestState
// 		ts = mem.RedirectedAt
// 	} else {
// 		latestState = None
// 	}
// 	return
// }

// func getLastCancelled (url string) (latestState memState, ts time.Time) {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	var inMem bool
// 	var mem *memory
// 	if mem, inMem = urlMap[url]; inMem {
// 		latestState = mem.LatestState
// 		ts = mem.CancelledAt
// 	} else {
// 		latestState = None
// 	}
// 	return
// }

// func getLastPended (url string) (latestState memState, ts time.Time) {
// 	lock.RLock()
// 	defer lock.RUnlock()
// 	var inMem bool
// 	var mem *memory
// 	if mem, inMem = urlMap[url]; inMem {
// 		latestState = mem.LatestState
// 		ts = mem.PendedAt
// 	} else {
// 		latestState = None
// 	}
// 	return
// }

// func memorizePended (url string) (err error) {
// 	lock.Lock()
// 	defer lock.Unlock()

// 	if m, exists := urlMap[url]; exists {
// 		m.PendedAt = time.Now() // This happens for unmarke item bursted and mem that is older then 24hr
// 		m.LatestState = Pended
// 		return
// 	} else {
// 		if len(urls[memPointer]) > 0 {
// 			delete(urlMap, urls[memPointer])
// 		}
// 		urls[memPointer] = url
// 		urlMap[url] = &memory{ Index: memPointer, PendedAt: time.Now(), LatestState: Pended }
// 		memPointer++
// 		if memPointer >= MEM_SIZE {
// 			memPointer = 0
// 		}
// 	}
// 	return
// }

// func memorizeShared (url string) (err error) {
// 	lock.Lock()
// 	defer lock.Unlock()

// 	if m, exists := urlMap[url]; exists {
// 		m.SharedAt = time.Now() // This happens for unmarke item bursted and mem that is older then 24hr
// 		m.LatestState = Shared
// 		return
// 	} else {
// 		if len(urls[memPointer]) > 0 {
// 			delete(urlMap, urls[memPointer])
// 		}
// 		urls[memPointer] = url
// 		urlMap[url] = &memory{ Index: memPointer, SharedAt: time.Now(), LatestState: Shared }
// 		memPointer++
// 		if memPointer >= MEM_SIZE {
// 			memPointer = 0
// 		}
// 	}
// 	return
// }


// func memorizeResult (url string, state memState) (err error) {
// 	lock.Lock()
// 	defer lock.Unlock()

// 	if m, exists := urlMap[url]; exists {
// 		if state == Redirected {
// 			m.RedirectedAt = time.Now()
// 		} else if state == Cancelled {
// 			m.CancelledAt = time.Now()
// 		}
// 		m.LatestState = state
// 		return
// 	} else {
// 		if len(urls[memPointer]) > 0 {
// 			delete(urlMap, urls[memPointer])
// 		}
// 		urls[memPointer] = url
// 		if state == Redirected {
// 			urlMap[url] = &memory{ Index: memPointer, RedirectedAt: time.Now(), LatestState: state }
// 		} else if state == Cancelled {
// 			urlMap[url] = &memory{ Index: memPointer, CancelledAt: time.Now(), LatestState: state }
// 		}
// 		memPointer++
// 		if memPointer >= MEM_SIZE {
// 			memPointer = 0
// 		}
// 	}
// 	return
// }

// func forget (url string) {
// 	lock.Lock()
// 	defer lock.Unlock()
// 	if mem, exists := urlMap[url]; exists {
// 		return
// 	} else {
// 		urls[mem.Index] = ""
// 		delete(urlMap, url)
// 	}
// }

// func memStateToString (ms memState) string {
// 	switch ms {
// 	case Shared:
// 		return "Shared"
// 	case Pended:
// 		return "Pended"
// 	case Redirected:
// 		return "Redirected"
// 	case Cancelled:
// 		return "Cancelled"
// 	default:
// 		return "???"
// 	}
// }