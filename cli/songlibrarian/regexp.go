package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1

var regexUrlMapping = []*regexp2.Regexp{
	regexYoutubeUrl_u0,
}

var regexYoutubeUrl_u0 = regexp2.MustCompile(`((?:https?:)\/\/)((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)(?!c|post|playlist)([\w\-]+)(\S+)?$`, regexp2.Multiline)

var regexCoverPlus = regexp2.MustCompile(`γ\s?(cover|πππ£ππ|ζ­γγΎγγ|ζ­γ£γ¦γΏγ|ηΏ»ε±)\s?γ`, regexp2.IgnoreCase | regexp2.Multiline)
var regexOriginalPlus = regexp2.MustCompile(`γ\s?(original|εε΅|γͺγͺγΈγγ«)\s?γ`, regexp2.IgnoreCase | regexp2.Multiline)
var regexStreamPlus = regexp2.MustCompile(`γ\s?(ζ­ζ |sing(ing)?|ζ­ε)\s?γ`, regexp2.IgnoreCase | regexp2.Multiline)

var regexCover_s0 = regexp2.MustCompile(`(cover|πππ£ππ)(?!\s?live)|γγγ£?γ¦γΏγ(?!.{0,4}(ζ |ιδΏ‘))|ζ­γγΎγγ(?!.{0,4}(ζ |ιδΏ‘))|ζ­γ£γ¦γΏγ(?!.{0,4}(ζ |ιδΏ‘))|ζ­γΏγ|θΈγ£γ¦γΏγ|ηΏ»ε±|ηΏ»ε±|γ«γ|θ©¦θε±δΊ|θ―ηε±δΊ|(?<!(full|short|tv.{0,4}|remix\s?))ver\.|(arrange\s|εΌΎγθͺγ)\s?ver\s`, regexp2.IgnoreCase | regexp2.Multiline)
var regexOriginal_s1 = regexp2.MustCompile(`original(?!\s?mv)|γͺγͺγΈγγ«|εε΅|music\s?video|mv|οΌ­οΌΆ|official|feat\.|ft\.|new single|M.?V.+from.+album|(1st|2nd|3rd|4th|5th|6th|7th|8th)\s?single|γ(1st|2nd|3rd|4th|5th|6th|7th|8th)\s?γ’γ«γγ \s?ει²ζ²γ`, regexp2.IgnoreCase | regexp2.Multiline)
var regexStream_s2 = regexp2.MustCompile(
`ζ­γ£γ¦γ(γ|η₯)γγ|γ\s?(sing(ing)?|song|ζ­|ε±ζ­|ζ­ζ )\s?γ|(γγ|ζ­|ζ­ζ |sing)\s?(γͺγγ|γ¨|οΌ|\+|οΌ|and)\s?(γ―γͺγ|θ©±γ|γγ’γ|ιθ«|δ½ζ₯­|ζΌε₯|talk)|(γ―γͺγ|θ©±γ|γγ’γ|ιθ«|δ½ζ₯­|ζΌε₯|talk)\s?(γͺγγ|γ¨|οΌ|\+|οΌ|and)\s?(γγ|ζ­|sing)|sing a song|sing songs|(sing|song|ζ­|piano).{0,8}stream|(γγ|ζ­).{0,3}(γγγγγ|η·΄)|ζζ­|(ζ­|ζΌε₯).{0,2}ζ |γγγγ|(ζ­|γγ)(οΌ|γ|ιθ«)|ζ­ε|ζ­ε|γγγ|ζ­γ|(γγ|ζ­)γ.?γΎ.?γ|ζ­.{0,4}ιδΏ‘|εΌΎγθͺγ(?!\s?ver)|γζ­|γγγ|(γγ|ζ­).{0,7}(ζι|live)|γLIVE #\d{1,3}γ|ζ­γ£γγ|.araoke|γ«γ©γͺγ±|(γγ|ζ­|sing).*(guerilla|γγγ|γ²γͺγ©)|(guerilla|γγγ|γ²γͺγ©).{0,5}(γγ|ζ­|sing)|η.{0,6}(Live|γ©γ€γ|ζΎι|ζ­γ³γ©γ)|mini live|γγγ©γ€γ|(mini|live) concert|virtual.+?concert|θδΉ.{0,12}(ζ²|γγ|ζ­|sing)|(ζ²|γγ|ζ­|sing).{0,12}θδΉ|cover live|γ«γγΌγ©γ€γ`, regexp2.IgnoreCase | regexp2.Multiline)

var regexClips = regexp2.MustCompile(`(?<!video)clip|εγζγ|translat(e|ion)|η€θ|εͺθΌ―|εη|θ­―|η²Ύθ―|εͺθΎ|θ―|η²Ύε`, regexp2.IgnoreCase | regexp2.Multiline)


var regexBadForOriginal = regexp2.MustCompile(`cover|γ¦γΏγ|(original:|γͺγͺγΈγγ«)\s?mv.+(γ¦γΏγ|cover)|(γ¦γΏγ|cover).+original:|γͺγͺγΈγγ«\s?mv|ε¬ι|full.ver|(?!^)officialι«­η·dism`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForCover = regexp2.MustCompile(`live(?!\s?video)|singing stream|ζ­ζ `, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForStream = regexp2.MustCompile(`debut|birthday stream|reaction`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForAll = regexp2.MustCompile(`REMIXζ |γ(ιθ«|δ½ζ₯­|apex|asmr)γ|θΏ·ε |trailer|XFD`, regexp2.IgnoreCase | regexp2.Multiline)

var regexCoverDesc = regexp2.MustCompile(`(ζ¬ε?Ά|εε±|original:)([^\n\r]{1,4}\n?)http[^\n\r]+?(?![^\n\r]+?ιδΏ‘)[^\n\r]+?$|(ζ¬ε?Ά|original:|εε±)([^\n\r]{1,4}\n?)http[^\n\r]+?(?![^\n\r]+?ιδΏ‘)[^\n\r]+?$|#ζ­γ£γ¦γΏγ`, regexp2.IgnoreCase | regexp2.Multiline)
var regexOriginalDesc = regexp2.MustCompile(`buy[^\n\r]{1,8}\n?http.+?$|#γγ«γ­γͺγͺγΈγγ«ζ²|#γͺγͺγΈγγ«ζ²`, regexp2.IgnoreCase | regexp2.Multiline)

var regexBadDescForOriginal = regexp2.MustCompile(`(ζ¬ε?Ά|original)(.+?\n?)http.+?$|from original`, regexp2.IgnoreCase | regexp2.Multiline)

var regexNamedOriginal = regexp2.MustCompile(`- ι·η¬ζθ± \((Official|Original)\)`, regexp2.IgnoreCase | regexp2.Multiline)
var regexNamedCover = regexp2.MustCompile(`ver.EMA`, regexp2.IgnoreCase | regexp2.Multiline)
var regexNamedStream = regexp2.MustCompile(`γYouTube Liveγζ³’ηΎγι¬Ό - Harano oni|ιΏζ­γ·γγ?γγγ«γ»`, regexp2.IgnoreCase | regexp2.Multiline)

var regexMention = regexp2.MustCompile(`^<@!(\d+)>`, regexp2.IgnoreCase | regexp2.Multiline)

var regexLinks = regexp2.MustCompile(`http(?!.+?discord\.com\/channels)`, regexp2.IgnoreCase | regexp2.Multiline)
var regexNotSkipLinks = regexp2.MustCompile(`(?<!\.)http(?!.+?discord\.com\/channels)`, regexp2.IgnoreCase | regexp2.Multiline)