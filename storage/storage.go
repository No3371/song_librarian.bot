package storage


type StorageProvider interface {
	SaveChannelMapping(cId uint64, bIDs map[int]struct{}) (err error)
	LoadChannelMapping(cId uint64) (bIDs map[int]struct{}, err error)

	SaveBinding(bId int, b interface{}) (err error)
	LoadBinding(bId int, b interface{}) (err error)
	RemoveBinding(bId int) (err error)

	GetBindingCount() (count int, err error)

	SaveCommandId(defId int, cmdId uint64, version uint32) (err error)
	LoadCommandId(defId int) (cmdId uint64, version uint32, err error)

	SaveSubState(uId uint64, state bool) (err error)
	LoadSubState(uId uint64) (state bool, err error)

	// SaveMem save a single record into the database
	SaveMem(tId int, id int, data string) (err error)
	LoadMems (tId int, from int, to int, deserializer func (slot int, data string) error) (err error)
	GetLatestMemIndex (tId int) (int, error)
	// LoadMem get a single record from the database
	// LoadMem(tId uint64, id int) (data string, err error)

	// // SaveMemTrack save the whole serialized track into the database
	// SaveMemTrack (id int, trackJson string) (err error)
	// // LoadMemTrack get the whole serialized track from the database
	// // By utilise this, we can avoid reading all rows on startup
	// LoadMemTrack (id int) (trackJson string, err error)

	SaveAll() (err error)
	Close() (err error)
}