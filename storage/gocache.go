package storage

// import (
// 	"fmt"

// 	"No3371.github.com/song_librarian.bot/binding"
// 	"github.com/patrickmn/go-cache"
// )

// func (g *goCache) SaveChannelMapping(cId int64, bId int64) (err error) {
// 	g.Set(fmt.Sprintf("c%d", cId), bId, cache.NoExpiration)
// 	return nil
// }

// func (g *goCache) SaveBinding(b *binding.Binding) (err error) {
// 	panic("not implemented") // TODO: Implement
// }

// func (g *goCache) LoadChannelMapping(cId int64) (bId int64, err error) {
// 	panic("not implemented") // TODO: Implement
// }

// func (g *goCache) LoadBinding(bId int64) (b *binding.Binding, err error) {
// 	panic("not implemented") // TODO: Implement
// }

// func (g *goCache) SaveToDisk() (err error) {
// 	g.Items()
// }


// type goCache struct {
// 	*cache.Cache
// }

// func GoCache () (gc storageProvider) {
// 	gc = &goCache {
// 		cache.New(cache.NoExpiration, 0),
// 	}
// 	return gc
// }