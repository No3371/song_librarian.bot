package binding

import (
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/redirect"
	"No3371.github.com/song_librarian.bot/storage"
	"github.com/vmihailenco/msgpack/v5"
)

var bindingCount int
var allBindings map[int]*ChannelBinding // [bId]
var mapping     map[uint64]map[int]struct{}
var changedBindings map[int]struct{}
var changedMappings map[uint64]struct{}
var sp storage.StorageProvider

func init () {
	var err error
	sp, err = storage.Sqlite()
	if err != nil {
		logger.Logger.Fatalf("%s", err)
	}

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
	EnabledUrlRegexes int
	Redirections map[redirect.RedirectType]uint64 // The index is RedirectType
}

// ChannelBinding defines these properties of a channel:
// - Which URL regexes are enabled
//	 - For any embed matching the URL regex:
//    - For each RedirectType
// 	    - What channel it redirects to
type ChannelBinding struct {
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
		} else if bIds == nil || len(bIds) == 0 {
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
	if exists {
		return b
	} else {
		var loaded *ExportedChannelBinding = &ExportedChannelBinding{}
		var err error
		if err = sp.LoadBinding(bId, loaded); err != nil {
			logger.Logger.Errorf("[BINDING] %v", err)
			allBindings[bId] = nil
			return nil
		} else {
			b = &ChannelBinding{
				enabledUrlRegexes: loaded.EnabledUrlRegexes,
				redirections: loaded.Redirections,
			}
			allBindings[bId] = b
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
		redirections: make(map[redirect.RedirectType]uint64),
	}
	changedBindings[bId] = struct {}{}
	return bId
}

func IterateAllBinding (iterator func (bId int, b *ChannelBinding)) {
	for bId, b := range allBindings {
		iterator(bId, b)
	}
}