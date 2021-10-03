package storage

// import (
// 	"strconv"

// 	"No3371.github.com/song_librarian.bot/guilds"
// 	"github.com/dgraph-io/badger/v3"
// 	"github.com/pkg/errors"
// )

// type Badger struct {}

// var instance *badger.DB

// func (b *Badger) SaveGuild(g *guilds.Guild) (err error) {
// 	err = b.assureInstance()
// 	if err != nil {
// 		return err
// 	}

// 	return instance.Update(func(txn *badger.Txn) error {
// 		b, err := guilds.Serialize(g)
// 		if err != nil {
// 			return err
// 		}
// 		err = txn.Set([]byte(g.ID.String()), b)
// 		return err
// 	})
// }

// func (b *Badger) LoadGuild(gId int64) (g *guilds.Guild, err error)  {
// 	err = b.assureInstance()
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = instance.View(func(txn *badger.Txn) error {

// 		i, err := txn.Get([]byte(strconv.FormatInt(gId, 10)))
// 		if err != nil {
// 			return err
// 		}

// 		err = i.Value(func(val []byte) error {
// 			g, err = guilds.Deserialize(val)
// 			return err
// 		})

// 		return err
// 	})

// 	return g, err
// }

// func (b *Badger) assureInstance () (err error) {
// 	if instance == nil || instance.IsClosed() {
// 		// Open the Badger database located in the /tmp/badger directory.
// 		// It will be created if it doesn't exist.
// 		instance, err = badger.Open(badger.DefaultOptions("/tmp/badger"))
// 		if err != nil {
// 			return errors.Wrap(err, "An error occured when opening Badger")
// 		}
// 		return nil
// 	} else {
// 		return nil
// 	}
// }
