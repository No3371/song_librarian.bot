package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1

var regexUrlMapping = []*regexp2.Regexp{
	regexYoutubeUrl_u0,
}

var regexYoutubeUrl_u0 = regexp2.MustCompile(`^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)(?!c|post|playlist)([\w\-]+)(\S+)?$`, 0)

var regexCoverPlus = regexp2.MustCompile(`【\s?(cover|𝑐𝑜𝑣𝑒𝑟|歌いました|歌ってみた|翻唱)\s?】`, regexp2.IgnoreCase | regexp2.Multiline)
var regexOriginalPlus = regexp2.MustCompile(`【\s?(original|原創|オリジナル)\s?】`, regexp2.IgnoreCase | regexp2.Multiline)
var regexStreamPlus = regexp2.MustCompile(`【\s?(歌枠|sing(ing)?|歌回)\s?】`, regexp2.IgnoreCase | regexp2.Multiline)

var regexCover_s0 = regexp2.MustCompile(`(cover|𝑐𝑜𝑣𝑒𝑟)(?!\s?live)|うたっ?てみた(?!.{0,4}(枠|配信))|歌いました(?!.{0,4}(枠|配信))|歌ってみた(?!.{0,4}(枠|配信))|歌みた|踊ってみた|翻唱|翻唱|カバ|試著唱了|试着唱了|\sver\.|(arrange\s|弾き語り)\s?ver\s`, regexp2.IgnoreCase | regexp2.Multiline)
var regexOriginal_s1 = regexp2.MustCompile(`original(?!\s?mv)|オリジナル|原創|music\s?video|pv|mv|ＭＶ|official|feat\.|ft\.|new single|M.?V.+from.+album|(1st|2nd|3rd|4th|5th)\s?single`, regexp2.IgnoreCase | regexp2.Multiline)
var regexStream_s2 = regexp2.MustCompile(
`歌ってお(し|知)らせ|【\s?(sing(ing)?|song|歌|唱歌|歌枠)\s?】|(うた|歌|sing)\s?(ながら|と|＆|\+|＋|and)\s?(はなし|話し|ピアノ|雑談|作業|演奏|talk)|(はなし|話し|ピアノ|雑談|作業|演奏|talk)\s?(ながら|と|＆|\+|＋|and)\s?(うた|歌|sing)|sing a song|sing songs|(sing|song|歌|piano).{0,8}stream|(うた|歌).{0,3}(れんしゅう|練)|朝歌|(歌|演奏).{0,2}枠|うたわく|(歌|うた)(！|。|雑談)|歌回|歌回|うたう|歌う|(うた|歌)い.?ま.?す|歌.{0,4}配信|弾き語り(?!\s?ver)|お歌|(うた|歌).{0,7}(時間|live)|歌ったり|.araoke|カラオケ|(うた|歌|sing).*(guerilla|げりら|ゲリラ)|(guerilla|げりら|ゲリラ).{0,5}(うた|歌|sing)|生.{0,6}(Live|ライブ|放送|歌コラボ)|mini live|ミニライブ|(mini|live) concert|virtual.+?concert|耐久.{0,12}(曲|うた|歌|sing)|(曲|うた|歌|sing).{0,12}耐久|cover live|カバーライブ`, regexp2.IgnoreCase | regexp2.Multiline)

var regexClips = regexp2.MustCompile(`(?<!video)clip|切り抜き|translat(e|ion)|烤肉|剪輯|切片|譯|精華|剪辑|译|精华`, regexp2.IgnoreCase | regexp2.Multiline)


var regexBadForOriginal = regexp2.MustCompile(`cover|てみた|(original:|オリジナル)\s?mv.+(てみた|cover)|(てみた|cover).+original:|オリジナル\s?mv|公開|full.ver|(?!^)official髭男dism`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForCover = regexp2.MustCompile(`live|singing stream|歌枠`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForStream = regexp2.MustCompile(`debut|birthday stream|reaction`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForAll = regexp2.MustCompile(`REMIX枠|【(雑談|作業)】|迷因|trailer|XFD`, regexp2.IgnoreCase | regexp2.Multiline)

var regexCoverDesc = regexp2.MustCompile(`(本家|原唱|original:)([^\n\r]{1,4}\n?)http[^\n\r]+?(?![^\n\r]+?配信)[^\n\r]+?$|(本家|original:|原唱)([^\n\r]{1,4}\n?)http[^\n\r]+?(?![^\n\r]+?配信)[^\n\r]+?$`, regexp2.IgnoreCase | regexp2.Multiline)
// var regexOriginalDesc = regexp2.MustCompile(`buy[^\n\r]{1,8}\n?http.+?$`, regexp2.IgnoreCase | regexp2.Multiline)

var regexBadDescForOriginal = regexp2.MustCompile(`(本家|original)(.+?\n?)http.+?$|from original`, regexp2.IgnoreCase | regexp2.Multiline)

var regexNamedCover = regexp2.MustCompile(`ver.EMA`, regexp2.IgnoreCase | regexp2.Multiline)
var regexNamedStream = regexp2.MustCompile(`【YouTube Live】波羅ノ鬼 - Harano oni|響歌シノのヒビカセ`, regexp2.IgnoreCase | regexp2.Multiline)

var regexMention = regexp2.MustCompile(`^<@!(\d+)>`, regexp2.IgnoreCase | regexp2.Multiline)