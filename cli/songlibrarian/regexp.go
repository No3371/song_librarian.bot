package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1

var regexYoutubeUrl_u0 = regexp2.MustCompile(`^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)(?!c|post|playlist)([\w\-]+)(\S+)?$`, 0)
var regexCover_s0 = regexp2.MustCompile(`(cover|𝑐𝑜𝑣𝑒𝑟)(?!\s?live)|うたっ?てみた|歌いました|歌ってみた|歌みた|踊ってみた|翻唱|翻唱|カバ|試著唱了|试着唱了|\sver\.|(arrange\s|弾き語り)\s?ver`, regexp2.IgnoreCase | regexp2.Multiline)
var regexOriginal_s1 = regexp2.MustCompile(`original(?!\s?mv)|オリジナル|原創|music video|mv|official|feat\.|ft\.|new single|M.?V.+from.+album`, regexp2.IgnoreCase | regexp2.Multiline)
var regexStream_s2 = regexp2.MustCompile(`【(sing(ing)?|song|歌)】|歌(と|＆|＆)(ピアノ|雑談|作業|演奏)|(ピアノ|雑談|作業|演奏)(と|＆)歌|sing a song|(sing|song|歌).+stream|(うた|歌).{0,3}(れんしゅう|練)|朝歌|(歌|演奏).{0,2}枠|うたわく|(歌|うた)(！|。)|歌回|歌回|うたう|歌う|(うた|歌)い.?ま.?す|歌.{0,4}配信|弾き語り(?!\s?ver)|お歌|歌ったり|.araoke|カラオケ|(うた|歌|sing).*(guerilla|げりら|ゲリラ)|(guerilla|げりら|ゲリラ).*(うた|歌|sing)|生.{0,6}(Live|ライブ|放送)\s|mini concert|耐久.{0,12}(曲|うた|歌|sing)|(曲|うた|歌|sing).{0,12}耐久|cover Live`, regexp2.IgnoreCase | regexp2.Multiline)

var regexClips = regexp2.MustCompile(`clip|切り抜き|translate|translation|烤肉|剪輯|切片|譯|精華|剪辑|译|精华`, regexp2.IgnoreCase | regexp2.Multiline)

var regexUrlMapping = []*regexp2.Regexp{
	regexYoutubeUrl_u0,
}

var regexBadForOriginal = regexp2.MustCompile(`cover|てみた|(original|オリジナル)\s?mv.+(てみた|cover)|(てみた|cover).+(original|オリジナル)\s?mv|公開`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForCover = regexp2.MustCompile(`live`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForStream = regexp2.MustCompile(`debut|birthday stream|食べ`, regexp2.IgnoreCase | regexp2.Multiline)
var regexBadForAll = regexp2.MustCompile(`trailer`, regexp2.IgnoreCase | regexp2.Multiline)

var regexCoverDesc = regexp2.MustCompile(`(本家|original)(.+?\n?)http.+?$`, regexp2.IgnoreCase | regexp2.Multiline)

var regexBadDescForOriginal = regexp2.MustCompile(`(本家|original)(.+?\n?)http.+?$|from original`, regexp2.IgnoreCase | regexp2.Multiline)

var regexNamedCover = regexp2.MustCompile(`ver.EMA`, regexp2.IgnoreCase | regexp2.Multiline)
var regexNamedStream = regexp2.MustCompile(`【YouTube Live】波羅ノ鬼 - Harano oni|響歌シノのヒビカセ`, regexp2.IgnoreCase | regexp2.Multiline)

var regexMention = regexp2.MustCompile(`^<@!(\d+)>`, regexp2.IgnoreCase | regexp2.Multiline)