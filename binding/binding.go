package binding

import (
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/memory"
	"No3371.github.com/song_librarian.bot/redirect"
	"No3371.github.com/song_librarian.bot/storage"
	"github.com/vmihailenco/msgpack/v5"
)

var bindingCount int
var allBindings map[int]*ChannelBinding // [bId]
var mapping     map[uint64]map[int]struct{} // Channel(s) to Binding
var changedBindings map[int]struct{}
var changedMappings map[uint64]struct{}
var sp storage.StorageProvider


func Setup (s storage.StorageProvider) {
	sp = s
	allBindings = make(map[int]*ChannelBinding)
	mapping= make(map[uint64]map[int]struct{})
	changedBindings = make(map[int]struct{})
	changedMappings = make(map[uint64]struct{})
}

type MutableChannelBinding struct {
	*ChannelBinding
}

func (mcb *MutableChannelBinding) EnableUrlRegexes (indexes ...int) {
	for i:= 0; i < len(indexes); i++ {
		mcb.enabledUrlRegexes |= 1 << indexes[i]
	}
}

func (mcb *MutableChannelBinding) DisableUrlRegexes (indexes ...int) {
	for i:= 0; i < len(indexes); i++ {
		mcb.enabledUrlRegexes &= ^(1 << indexes[i])
	}
}

func (mcb *MutableChannelBinding) SetRedirection (r redirect.RedirectType, dest uint64) {
	mcb.redirections[r] = dest
}

func (mcb *MutableChannelBinding) RemoveRedirection (r redirect.RedirectType) {
	delete(mcb.redirections, r)
}

type ExportedChannelBinding struct {
	memory.MemTrack
	EnabledUrlRegexes int
	Redirections map[redirect.RedirectType]uint64 // The index is RedirectType
}

// ChannelBinding defines these properties of a channel:
// - Which URL regexes are enabled
//	 - For any embed matching the URL regex:
//    - For each RedirectType
// 	    - What channel it redirects to
type ChannelBinding struct {
	memory.MemTrack
	enabledUrlRegexes int
	redirections map[redirect.RedirectType]uint64 // The index is RedirectType
}

func (cb *ChannelBinding) UrlRegexEnabled (rIndex int) bool {
	bit := 1 << rIndex
	bit = cb.enabledUrlRegexes & bit
	return bit != 0
}
func (cb *ChannelBinding) DestChannelId (rType redirect.RedirectType) (cId uint64, exist bool) {
	cId, exist = cb.redirections[rType]
	return cId, exist
}

func Serialize (b *ChannelBinding) (bytes []byte, err error) {
	return msgpack.Marshal(b)
}

func SaveAll () (err error) {
	for bId := range changedBindings {
		b := QueryBinding(bId) // It gets loaded if it's not
		logger.Logger.Infof("[BINDING] Saving Binding#%d", bId)
		err = sp.SaveBinding(bId, ExportedChannelBinding {
			MemTrack: b.MemTrack,
			EnabledUrlRegexes: b.enabledUrlRegexes,
			Redirections: b.redirections,
		})
		if err != nil {
			logger.Logger.Errorf("%v", err)
		}
		delete(changedBindings, bId)
	}

	for cId := range changedMappings {
		logger.Logger.Infof("[BINDING] Saving Mapping#%d", cId)
		err = sp.SaveChannelMapping(cId, mapping[cId])
		if err != nil {
			logger.Logger.Errorf("%v", err)
		}
		delete(changedMappings, cId)
	}

	return nil
}

func Deserialize (bytes []byte) (b *ChannelBinding, err error) {
	b = &ChannelBinding{}
	err = msgpack.Unmarshal(bytes, b)
	return b, err
}

func Bind (cId uint64, bId int) {
	_, exists := mapping[cId]
	if !exists {
		mapping[cId] = make(map[int]struct{})
	}
	mapping[cId][bId] = struct{}{}
	changedMappings[cId]=struct{}{}
	logger.Logger.Infof("Binding#%d is bound to channel %d", bId, cId)
}

func Unbind (cId uint64, bId int) {
	_, exists := mapping[cId]
	if exists {
		delete(mapping[cId], bId)
		changedMappings[cId]=struct{}{}
	}
}

func QuickQueryRedirect (cId uint64, rt redirect.RedirectType) bool {
	if bIds := GetMappedBindingIDs(cId); bIds == nil {
		return false
	} else {
		for bId := range bIds {
			b := QueryBinding(bId)
			if b != nil {
				if _, redirecting := b.redirections[rt]; redirecting {
					return true
				}
			}
		}
		return false
	} 
}
func GetMappedBindingIDs (cId uint64) map[int]struct{} {
	ids, exists := mapping[cId]
	if exists {
		return ids
	} else {
		if bIds, err := sp.LoadChannelMapping(cId); err != nil {
			logger.Logger.Errorf("[BINDING] %v", err)
			mapping[cId] = nil
			return nil
		} else if len(bIds) == 0 {
			return nil
		} else {
			ids = bIds
			mapping[cId] = ids
			return ids
		}
	}
}

func QueryBinding (bId int) *ChannelBinding {
	b, exists := allBindings[bId]
	if exists { // Loaded
		return b
	} else { // Need to load or create
		var loaded *ExportedChannelBinding = &ExportedChannelBinding{}
		var err error
		if err = sp.LoadBinding(bId, loaded); err != nil { // Failed to load
			logger.Logger.Errorf("[BINDING] %v", err)
			allBindings[bId] = nil
			return nil
		} else { // Loaded
			b = &ChannelBinding{
				MemTrack: loaded.MemTrack,
				enabledUrlRegexes: loaded.EnabledUrlRegexes,
				redirections: loaded.Redirections,
			}
			if b.MemTrack.SetupMemTrack (bId) {
				changedBindings[bId] = struct{}{}
			}
			// if b.MemTrack == nil {
			// 	changedBindings[bId] = struct{}{}
			// 	b.MemTrack = *memory.NewMemTrack(bId)
			// }
			allBindings[bId] = b
			logger.Logger.Infof("[BINDING] Loaded#%d", bId)
			return b
		}
	}
}

func GetModifiableBinding (bId int) *MutableChannelBinding {
	b := QueryBinding(bId)
	if b == nil {
		return nil
	}
	changedBindings[bId] = struct{}{}
	return &MutableChannelBinding {
		b,
	}
}

func NewBinding () (bId int) {
	bindingCount ++
	bId = bindingCount
	allBindings[bId] = &ChannelBinding{
		MemTrack: *memory.NewMemTrack(bId),
		redirections: make(map[redirect.RedirectType]uint64),
	}
	changedBindings[bId] = struct {}{}
	return bId
}

func IterateAllMapping (loadAllBounds bool, iterator func (cId uint64, b *ChannelBinding)) {
	for cId, bIds := range mapping {
		for bId := range bIds {
			if loadAllBounds {
				b := QueryBinding(bId)
				iterator(cId, b)
			}
		}
	}
}