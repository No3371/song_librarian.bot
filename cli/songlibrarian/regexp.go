package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1

var regexYoutubeUrl_u0 = regexp2.MustCompile(`^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)(?!c|post)([\w\-]+)(\S+)?$`, 0)
var regexCover_s0 = regexp2.MustCompile(`cover|うたっ?てみた|歌ってみた|歌みた|踊ってみた|翻唱|翻唱|カバ|試著唱了|试着唱了|\sver\.|arrange ver`, regexp2.IgnoreCase)
var regexOriginal_s1 = regexp2.MustCompile(`original(?!\s?mv)|オリジナル|原創|music video|mv|official|feat\.|ft\.|new single`, regexp2.IgnoreCase)
var regexStream_s2 = regexp2.MustCompile(`【(sing|歌)】|歌＆|歌(と|＆)(ピアノ|雑談|作業|演奏)|(ピアノ|雑談|作業|演奏)と歌|sing a song|singing|stream|(うた|歌).{0,3}(れんしゅう|練)|朝歌|(歌|演奏)枠(?!切り抜き)|うたわく|】(sing|歌|うた)(！|。|!|\.)|^(sing|歌|うた)(！|。|!|\.)|歌回|歌回|うたう|歌う|(うた|歌)い.?ま.?す|歌配信|弾き語り(?!\s?ver)|お歌|歌ったり|.araoke|(うた|歌|sing).*(guerilla|げりら|ゲリラ)|(guerilla|げりら|ゲリラ).*(うた|歌|sing)|生.{0,6}(Live|ライブ|放送)\s`, regexp2.IgnoreCase)

var regexClips = regexp2.MustCompile(`clip|切り抜き|translate|translation|烤肉|剪輯|切片|譯|精華|剪辑|译|精华`, regexp2.IgnoreCase)

var regexUrlMapping = []*regexp2.Regexp{
	regexYoutubeUrl_u0,
}

var regexSubjectMapping = []*regexp2.Regexp{
	regexCover_s0,
	regexOriginal_s1,
	regexStream_s2,
}

var regexBadForOriginal = regexp2.MustCompile(`cover|てみた|(original|オリジナル)\s?mv.+(てみた|cover)|(てみた|cover).+(original|オリジナル)\s?mv|公開します`, regexp2.IgnoreCase)
var regexBadForCover = regexp2.MustCompile(`live`, regexp2.IgnoreCase)
var regexBadForStream = regexp2.MustCompile(`debut|birthday stream|食べ`, regexp2.IgnoreCase)
var regexBadForAll = regexp2.MustCompile(`trailer`, regexp2.IgnoreCase)

var regexNamedCover = regexp2.MustCompile(`ver.EMA`, regexp2.IgnoreCase)
var regexNamedStream = regexp2.MustCompile(`【YouTube Live】波羅ノ鬼 - Harano oni`, regexp2.IgnoreCase)
