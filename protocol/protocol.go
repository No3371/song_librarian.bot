package protocol

// import (
// 	"No3371.github.com/song_librarian.bot/binding"
// 	"No3371.github.com/song_librarian.bot/redirect"
// )

// type CommandType int

// const (
// 	BIND CommandType = iota
// 	UNBIND
// 	EXIT
// )

// type Command struct {
// 	CommandId int
// }

// type Bind struct {
// 	Command
// 	SrcChannelId uint64
// 	BindingId int
// 	RType redirect.RedirectType
// 	DestChannelId uint64
// }

// type Unbind struct {
// 	Command
// 	SrcChannelId uint64
// 	BindingId int
// 	RType redirect.RedirectType
// }


// type Exit struct {
// 	Command
// }

// func CreateBinding () int {
// 	return binding.NewBinding()
// }

// func SetRedirection (bId int, rType redirect.RedirectType, destCId uint64) (err error) {
	
// }

// func Bind (srcCId uint64, bId int) (err error) {
// 	binding.Bind(cId uint64, bId int)
// }