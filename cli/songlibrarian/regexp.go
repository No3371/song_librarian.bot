package main

import "github.com/dlclark/regexp2"

const urlRegexCount = 1
var regexYoutubeUrl_u0 = regexp2.MustCompile(`^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)([\w\-]+)(\S+)?$`, 0)
var regexCover_s0 = regexp2.MustCompile(`cover|うたてみた|うたってみた|歌ってみた|踊ってみた|翻唱|カバ`, regexp2.IgnoreCase)
var regexOriginal_s1 = regexp2.MustCompile(`original|オリジナル|原創|music video|full mv|mv|official|feat\.`, regexp2.IgnoreCase)
var regexStream_s2 = regexp2.MustCompile(`sing|stream|歌枠|歌回|うたう|歌う|歌います|歌配信|弾き語り枠|お歌`, regexp2.IgnoreCase)

var regexClips = regexp2.MustCompile(`clip|切り抜き`, regexp2.IgnoreCase)

var regexUrlMapping = []*regexp2.Regexp {
	regexYoutubeUrl_u0,
}

var regexSubjectMapping = []*regexp2.Regexp {
	regexCover_s0,
	regexOriginal_s1,
	regexStream_s2,
}