package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1
var regexYoutubeUrl_u0 = regexp2.MustCompile(`^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)([\w\-]+)(\S+)?$`, 0)
var regexCover_s0 = regexp2.MustCompile(`cover|うた.?てみた|歌ってみた|歌みた|踊ってみた|翻唱|翻唱|カバ|試著唱了|试着唱了`, regexp2.IgnoreCase)
var regexOriginal_s1 = regexp2.MustCompile(`original(?!\s?mv)|オリジナル|原創|music video|mv|official|feat\.|new single`, regexp2.IgnoreCase)
var regexStream_s2 = regexp2.MustCompile(`【SING】|singing|stream|(うた|歌).{0,3}(れんしゅう|練)|歌枠(?!切り抜き)|うたわく|うた(！|。)|歌回|歌回|うたう|歌う|歌い.?ま.?す|歌配信|弾き語り|お歌|歌ったり|karaoke|(うた|歌|sing).*(guerilla|ゲリラ)|(guerilla|ゲリラ).*(うた|歌|sing)|生.{0,6}(Live|ライブ)`, regexp2.IgnoreCase)


var regexClips = regexp2.MustCompile(`clip|切り抜き|translate|translation|烤肉|剪輯|切片|譯|精華|剪辑|译|精华`, regexp2.IgnoreCase)
// var regexBadForCover = regexp2.MustCompile(`original\s?mv`, regexp2.IgnoreCase)
// var regexBadForOriginal = regexp2.MustCompile(`original\s?mv`, regexp2.IgnoreCase)
// var regexBadForStream = regexp2.MustCompile(`original\s?mv`, regexp2.IgnoreCase)

var regexUrlMapping = []*regexp2.Regexp {
	regexYoutubeUrl_u0,
}

var regexSubjectMapping = []*regexp2.Regexp {
	regexCover_s0,
	regexOriginal_s1,
	regexStream_s2,
}

var regexBadForOriginal = regexp2.MustCompile(`cover`, regexp2.IgnoreCase)
var regexBadForCover = regexp2.MustCompile(`live`, regexp2.IgnoreCase)
var regexBadForStream = regexp2.MustCompile(`debut|birthday stream|食べ`, regexp2.IgnoreCase)

var regexNamedCover = regexp2.MustCompile(`ver.EMA`, regexp2.IgnoreCase)
var regexNamedStream = regexp2.MustCompile(`【YouTube Live】波羅ノ鬼 - Harano oni`, regexp2.IgnoreCase)
