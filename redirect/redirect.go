package redirect

type RedirectType int 

const (
	Cover RedirectType = iota
	Original
	Stream
	None
	Unknown
)