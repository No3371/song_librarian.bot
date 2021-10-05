package redirect

type RedirectType int 

const (
	None RedirectType = iota
	Cover 
	Original
	Stream
	Unknown
)