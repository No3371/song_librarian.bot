package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1
var regexYoutubeUrl_u0 = regexp2.MustCompile(`^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)([\w\-]+)(\S+)?$`, 0)
var regexCover_s0 = regexp2.MustCompile(`cover|うたてみた|うたってみた|歌ってみた|歌みた|踊ってみた|翻唱|翻唱|カバ|試著唱了|试着唱了`, regexp2.IgnoreCase)
var regexOriginal_s1 = regexp2.MustCompile(`original(?!\s?mv)|オリジナル|原創|music video|full mv|mv|official|feat\.`, regexp2.IgnoreCase)
var regexStream_s2 = regexp2.MustCompile(`sing|stream|歌枠(?!切り抜き)|歌回|歌回|うたう|歌う|歌い.ます|歌配信|弾き語り|お歌|歌ったり|karaoke`, regexp2.IgnoreCase)

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