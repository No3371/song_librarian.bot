package redirect

type RedirectType int 

const (
	None RedirectType = iota
	Cover 
	Original
	Stream
	Unknown
	Clip
)

func RedirectTypetoString(rt RedirectType) string {
	switch rt {
	case None:
		return "NONE"
	case Cover:
		return "COVER"
	case Original:
		return "ORIGINAL"
	case Stream:
		return "STREAM"
	case Unknown:
		return "UNKNOWN"
	case Clip:
		return "CLIP"
	default:
		return "Unknown RedirectType"
	}
}