package memory

import (
	"encoding/json"
	"sync"
	"time"

	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/storage"
)


// var LoadedMemTracks map[int]*MemTrack
var sp storage.StorageProvider

type memState int 

const (
	None memState = iota
	Shared
	Pended
	Redirected
	Cancelled
	CancelledWithError
)

var MEM_SIZE = 4096

type Memory struct {
	Url string
	LatestState memState
	CancelledAt time.Time
	UnexpectedUnfinishedAt time.Time
	RedirectedAt time.Time
	SharedAt time.Time
	PendedAt time.Time
}


type MemTrack struct {
    TId int
    Mapping map[string]int
    List []*Memory
    MemPointer int
    LastTouchedMemPointer int
    lock *sync.RWMutex
}

func Setup (_sp storage.StorageProvider, memSize int) {
	sp = _sp
	MEM_SIZE = memSize
}

func NewMemTrack (tId int) *MemTrack {
	return &MemTrack{
        TId: tId,
        Mapping: make(map[string]int),
        List: make([]*Memory, MEM_SIZE),
        MemPointer: 0,
		LastTouchedMemPointer: 0,
		lock: &sync.RWMutex{},
    }
}

func (mt *MemTrack) SetupMemTrack (tId int) (updated bool) {
	if mt.Mapping == nil {
		mt.Mapping = make(map[string]int)
		updated = true
	}
	if mt.List == nil {
		mt.List = make([]*Memory, MEM_SIZE)
		updated = true
	}
	if mt.lock == nil {
		mt.lock = &sync.RWMutex{}
	}

	mt.TId = tId

	if latest, err := sp.GetLatestMemIndex(mt.TId); err != nil {
		logger.Logger.Errorf("Failed to read latest index of memtrack#%d: %v", mt.TId, err)
	} else {
		logger.Logger.Infof("Memtrack#%d: latest mem slot is %d, pointer is %d", tId, latest, mt.LastTouchedMemPointer)
		if (latest != mt.LastTouchedMemPointer - 1) {
			logger.Logger.Infof("Memtrack#%d: Found inconsistency between cached MemPointer and latest memory record, fixing...", mt.TId)
			mt.Mapping = make(map[string]int)
			mt.List = make([]*Memory, MEM_SIZE)
			err := sp.LoadMems(mt.TId, 0, MEM_SIZE, func (slot int, data string) error {
				innerErr := json.Unmarshal([]byte(data), &mt.List[slot])
				if innerErr != nil {
					logger.Logger.Errorf(" Failed to deserialize a mem record: %v", innerErr)
					return innerErr
				}
				mt.Mapping[mt.List[slot].Url] = slot
				updated = true
				return nil
			})
			mt.LastTouchedMemPointer = latest + 1
			if err != nil {
				logger.Logger.Errorf(" Failed to fix the memtrack: %v", err)
			}
		}
	}

	return
}

// func GetMemTrack (tId int) *MemTrack {
//     if m, err := LoadMemTrack(tId); err != nil {
// 		logger.Logger.Errorf("Failed to load MemTrack#%d: %s", tId, err)
// 		return nil
//     } else if (m != nil) {
// 		logger.Logger.Warnf("Trying to get a new memTrack#%d but the Id already used.", tId)
// 		return m
// 	}

//     return &MemTrack{
//         TId: tId,
//         Mapping: make(map[string]int, MEM_SIZE),
//         List: [MEM_SIZE]*Memory{},
//         MemPointer: 0,
//     }
// }

// func SaveMemTrack (tId int) (err error) {
//     if m, loaded := LoadedMemTracks[tId]; !loaded {
// 		return nil// The fact that the memtrack is not loaded should indicate it has not been used at all, therefore not changed at all
//     } else {
//         if data, err := json.Marshal(m); err != nil {
//             logger.Logger.Errorf("Failed to marshal memTrack#%d: %s", tId, err)
//             return err
//         } else {
//             if err = sp.SaveMemTrack(tId, string(data)); err != nil {
//                 logger.Logger.Errorf("Failed to save memTrack#%d: %s", tId, err)
//                 return err
//             }
//             return nil
//         }
//     }
// }

// func LoadMemTrack (tId int) (mt *MemTrack, err error) {
// 	var j string
// 	if j, err = sp.LoadMemTrack(tId); err != nil {
// 		logger.Logger.Errorf("Failed to load memtrack%d: %s", tId, err)
// 		return nil, err
// 	}

// 	err = json.Unmarshal([]byte(j), mt)
// 	if err != nil {
// 		logger.Logger.Errorf("Failed to unmarshal memtrack%d: %s", tId, err)
// 		return nil, err
// 	}

// 	return mt, nil
// }

// func SaveMem (tId int, index int) error {
// 	var mt *MemTrack
// 	var loaded bool
// 	if mt, loaded = LoadedMemTracks[tId]; !loaded {
// 		return nil// The fact that the memtrack is not loaded should indicate it has not been used at all, therefore not changed at all
// 	}

// 	j, err := json.Marshal(mt.List[index])
// 	if err != nil {
// 		logger.Logger.Errorf("Failed to marshal mem%d-%d: %s", tId, index, err)
// 		return err
// 	}

// 	err = sp.SaveMem(tId, index, string(j))
// 	if err != nil {
// 		logger.Logger.Errorf("Failed to save mem%d-%d: %s", tId, index, err)
// 		return err
// 	}

// 	return nil
// }

// func (mt *MemTrack) Save (tId int) error {
// 	j, err := json.Marshal(mt)
// 	if err != nil {
// 		logger.Logger.Errorf("Failed to marshal memtrack%d: %s", tId, err)
// 		return err
// 	}

// 	if err = sp.SaveMemTrack(tId, string(j)); err != nil {
// 		logger.Logger.Errorf("Failed to save memtrack%d: %s", tId, err)
// 		return err
// 	}

// 	return nil
// }

func (mt *MemTrack) GetLastState (url string) (state memState, ts time.Time) {
	if mt.lock == nil {
		mt.lock = &sync.RWMutex{}
	}
	mt.lock.RLock()
	defer mt.lock.RUnlock()
	var memorized bool
	var mIndex int
	var mem *Memory
	if mIndex, memorized = mt.Mapping[url]; memorized {
		mem = mt.List[mIndex]
		if mem == nil {
			return None, time.Time{}
		}
		state = mem.LatestState
		switch state {
		case Redirected:
			ts = mem.RedirectedAt
		case Shared:
			ts = mem.SharedAt
		case Pended:
			ts = mem.PendedAt
		case CancelledWithError:
			ts = mem.UnexpectedUnfinishedAt
		case Cancelled:
			ts = mem.CancelledAt
		default:
			logger.Logger.Errorf("WHAT IS THIS STATE? %s", state)
			ts = time.Time{}
		}
	} else {
		state = None
	}
	return

}

func (mt *MemTrack) GetLastTime (url string, state memState) time.Time {
	if mt.lock == nil {
		mt.lock = &sync.RWMutex{}
	}
	mt.lock.RLock()
	defer mt.lock.RUnlock()

	var mIndex int
	var memorized bool
	if mIndex, memorized = mt.Mapping[url]; memorized {
		mem := mt.List[mIndex]
		if mem == nil {
			return time.Time{}
		}
		switch state {
		case Redirected:
			return mem.RedirectedAt
		case Shared:
			return mem.SharedAt
		case Pended:
			return mem.PendedAt
		case CancelledWithError:
			return mem.UnexpectedUnfinishedAt
		case Cancelled:
			return mem.CancelledAt
		default:
			logger.Logger.Errorf("WHAT IS THIS STATE? %s", state)
			return time.Time{}
		}
	} else {
		return time.Time{}
	}
}

func (mt *MemTrack) Memorize (url string, state memState) (err error) {
	if mt.lock == nil {
		mt.lock = &sync.RWMutex{}
	}
	mt.lock.Lock()
	defer mt.lock.Unlock()

	var mIndex int
	var memorized bool
	if mIndex, memorized = mt.Mapping[url]; !memorized { // The url is not memorized
		mt.List[mt.MemPointer] = &Memory{ Url: url }
		mt.Mapping[url] = mt.MemPointer
		mIndex = mt.MemPointer
		mt.MemPointer++
		if mt.MemPointer >= MEM_SIZE {
			mt.MemPointer = 0
		}
	}

	mem := mt.List[mIndex]

	switch state {
	case Redirected:
		mem.RedirectedAt = time.Now()
	case Shared:
		mem.SharedAt = time.Now()
	case Pended:
		mem.PendedAt = time.Now()
	case CancelledWithError:
		mem.UnexpectedUnfinishedAt = time.Now()
	case Cancelled:
		mem.CancelledAt = time.Now()
	default:
		logger.Logger.Errorf("WHAT IS THIS STATE? %s", state)
		return
	}

	mem.LatestState = state

	j, err := json.Marshal(mem)
	if err != nil {
		logger.Logger.Errorf("Failed to marshal mem%d-%d: %s", mt.TId, mIndex, err)
		return err
	}

	err = sp.SaveMem(mt.TId, mIndex, string(j))
	if err != nil {
		logger.Logger.Errorf("Failed to save mem%d-%d: %s", mt.TId, mIndex, err)
		return err
	}

	mt.LastTouchedMemPointer = mIndex

	return nil
}

func (mt *MemTrack) Forget (url string, state memState) (err error) {
	if mt.lock == nil {
		mt.lock = &sync.RWMutex{}
	}
	mt.lock.Lock()
	defer mt.lock.Unlock()

	delete(mt.Mapping, url)

	return nil
}

func MemStateToString (ms memState) string {
	switch ms {
	case Shared:
		return "Shared"
	case Pended:
		return "Pended"
	case Redirected:
		return "Redirected"
	case Cancelled:
		return "Cancelled"
	case CancelledWithError:
		return "Cancelled(Error)"
	default:
		return "???"
	}
}