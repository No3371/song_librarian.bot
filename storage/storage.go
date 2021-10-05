package storage

type StorageProvider interface {
	SaveChannelMapping (cId uint64, bIDs map[int]struct{}) (err error)
	SaveBinding (bId int, b interface{}) (err error)
	LoadChannelMapping (cId uint64) (bIDs map[int]struct{}, err error) 
	LoadBinding (bId int, b interface{}) (err error)
	RemoveBinding (bId int) (err error)
	GetBindingCount () (count int, err error)
	SaveAll () (err error)
	Close () (err error)
}