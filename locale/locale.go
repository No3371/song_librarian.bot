package locale

import (
	"strings"

	"No3371.github.com/song_librarian.bot/logger"
)

type Locale int

const (
	TW Locale = iota
	EN
)

var HELLO string
var DETECTED string
var DETECTED_SPOILER string
var DETECTED_SPOILER_DUPLICATE string
var DETECTED_MATCH_NONE string
var DETECTED_MATCH_NONE_AND_SHARED string
var DETECTED_MATCH_NONE_AND_CANCELLED string
var DETECTED_UNKNOWN string
var DETECTED_UNKNOWN_DUPLICATE string
var DETECTED_CLIPS string
var DETECTED_CLIPS_AND_CANCELLED string
var DETECTED_CLIPS_DUPLICATE string
var DETECTED_DUPLICATE string
var DETECTED_DUPLICATE_NONE string
var DETECTED_PRE_TYPED string
var DETECTED_PRE_TYPED_AGREED string
var DETECTED_PRE_TYPED_DUPLICATE string
var DETECTED_PRE_TYPED_AGREED_DUPLICATE string
var DETECTED_PRE_TYPED_NONE_DUPLICATE string
var DETECTED_PRE_TYPED_NONE_AGREED_DUPLICATE string
var BUTTON_ORIGINAL string
var BUTTON_COVER string
var BUTTON_NOT_SONG string
var C_DESC string
var C_ORIGINAL_DESC string
var C_COVER_DESC string
var C_DELETE_ID_DESC string
var C_DESC_UNSUB string
var C_DESC_RESUB string

var ORIGINAL string
var COVER    string
var STREAM    string
var ORIGINAL_UNSIGNED string
var COVER_UNSIGNED    string
var STREAM_UNSIGNED    string
var DO_NOT_REDIRECT string

var SHARER string
var DECISION_TYPE string
var DECISION_BOT string
var DECISION_SHARER string
var DECISION_SHARER_AND_BOT string
var DECISION_SHARER_AND_COMMUNITY string
var DECISION_SHARER_AND_BOT_AND_COMMUNITY string
var DECISION_COMMUNITY_AGREE string
var DECISION_COMMUNITY_FIX string
var DECISION_COMMUNITY_HELP string
var SMSG string

var EXPLAIN_EMBED_RESOLVE string

var ACTIVITY string

var USAGE string

var HOT string

func FromString (code string) Locale {
	code = strings.ToLower(code)
	switch code {
	case "tw", "zh", "zh_tw", "zh-tw":
		return TW
	case "en", "en-us", "en_us":
		return EN
	default:
		return EN
	}
}

func ToString (locale Locale) string {
	switch locale {
	case TW:
		return "TW"
	case EN:
		return "EN"
	default:
		return "UNKNOWN"
	}
}

func SetLanguage (lang Locale) {
	logger.Logger.Infof("Setting language to %s", ToString(lang))
	switch lang {
	case EN:
	case TW:
		HELLO = "[啟動]"
		BUTTON_NOT_SONG = "非歌曲"
		BUTTON_ORIGINAL = "原創"
		BUTTON_COVER = "翻唱"
		DETECTED =                                 "▶️ **%s**\n**(๑•̀ㅂ•́)و✧** 猜測為 *%s*，%.0f 秒後自動轉發。可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_SPOILER =                         "▶️ ||**%s**||\n** ( ͡° ͜ʖ ͡°)** 這一坨是啥？預設 ❌*不轉發*。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_SPOILER_DUPLICATE =               "▶️ ||**%s**||\n** ( ͡° ͜ʖ ͡°)** 這一坨是啥？預設 ❌*不轉發*，但 **%s 前轉發過**...%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_CLIPS =                           "▶️ **%s**\n**/ᐠ｡ꞈ｡ᐟ\\\\** 疑似剪輯，預設 ❌*不轉發*。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_MATCH_NONE =                      "▶️ **%s**\n**( ˘•ω•˘ )** 標題不含關鍵字，預設 ❌*不轉發*。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_MATCH_NONE_AND_SHARED =           "▶️ **%s**\n**( ˘•ω•˘ )** 標題不含關鍵字，預設 ❌*不轉發*（ %s 前已被分享過）。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_MATCH_NONE_AND_CANCELLED =        "▶️ **%s**\n**( ˘•ω•˘ )** 標題不含關鍵字，預設 ❌*不轉發*（ %s 前曾經不轉發）。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_UNKNOWN =                         "▶️ **%s**\n**(ﾟ∀。)** 瓦卡拉奈，預設 ❌*不轉發*。多拉 A 夢幫幫我！%.0f 秒內可對此訊息投票決定分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_PRE_TYPED =                       "▶️ **%s**\n**(\\*゜ω゜)ゞ** 分享者建議為 *%s*，%.0f 秒後自動執行。可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_PRE_TYPED_AGREED =                "▶️ **%s**\n**(\\*゜ω゜)ゞ** 分享者建議為 *%s*（我也這麼認為！）%.0f 秒後自動執行。可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_PRE_TYPED_DUPLICATE =             "▶️ **%s**\n**(･ω´･ )** 分享者建議了 *%s*，但 **%s 前轉發過**，預設 ❌*不轉發*。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_CLIPS_DUPLICATE =                 "▶️ **%s**\n**(･ω´･ )** 疑似剪輯，預設 ❌*不轉發*，不過**%s 前轉發過**...%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_CLIPS_AND_CANCELLED =                 "▶️ **%s**\n**(･ω´･ )** 疑似剪輯，預設 ❌*不轉發*（ %s 前曾經不轉發）。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_DUPLICATE =                       "▶️ **%s**\n**(･ω´･ )** 猜測為 *%s*，但 **%s 前轉發過**，預設 ❌*不轉發*。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_UNKNOWN_DUPLICATE =               "▶️ **%s**\n**(･ω´･ )** 瓦卡拉奈，預設 ❌*不轉發*，不過**%s 前轉發過**...多拉 A 夢幫幫我！%.0f 秒內可對此訊息投票決定分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_DUPLICATE_NONE =                  "▶️ **%s**\n**(･ω´･ )** 猜測為 ❌*不轉發*，但 **%s 前轉發過**...%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_PRE_TYPED_AGREED_DUPLICATE =      "▶️ **%s**\n**(･ω´･ )ゞ** 分享者建議了 *%s*（我也這麼認為！），但 **%s 前轉發過**，預設 ❌*不轉發*。%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_PRE_TYPED_NONE_DUPLICATE =        "▶️ **%s**\n**(･ω´･ )ゞ** 分享者建議 ❌*不轉發*，且 **%s 前轉發過**...%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		DETECTED_PRE_TYPED_NONE_AGREED_DUPLICATE = "▶️ **%s**\n**(･ω´･ )ゞ** 分享者建議 ❌*不轉發*（我也這麼認為！），且 **%s 前轉發過**...%.0f 秒內可對此訊息投票修改分類：🇴 原創 / 🇨 翻唱 / 🇸 歌回 / ❌ 不轉發"
		C_DESC = "channel"
		C_COVER_DESC = "翻唱歌曲頻道 ID"
		C_ORIGINAL_DESC = "原創歌曲頻道 ID"
		C_DESC_UNSUB = "忽略我的分享"
		C_DESC_RESUB = "重新開始轉發我的分享"
		ORIGINAL = "🇴 原創"
		COVER = "🇨 翻唱"
		STREAM = "🇸 歌回"
		ORIGINAL_UNSIGNED = "原創"
		COVER_UNSIGNED = "翻唱"
		STREAM_UNSIGNED = "歌回"
		DO_NOT_REDIRECT = "❌不轉發"
		SHARER = "分享者"
		DECISION_TYPE = "判定"
		DECISION_BOT = "EXE🤖"
		DECISION_SHARER = "分享者🦸"
		DECISION_SHARER_AND_COMMUNITY = "分享者🦸🗳️"
		DECISION_SHARER_AND_BOT = "分享者🦸🤖"
		DECISION_SHARER_AND_BOT_AND_COMMUNITY = "全場通過🦸🤖🗳️"
		DECISION_COMMUNITY_AGREE = "社群確認🗳️🤖"
		DECISION_COMMUNITY_FIX = "社群修正🗳️"
		DECISION_COMMUNITY_HELP = "社群🗳️"
		SMSG = "原文"
		EXPLAIN_EMBED_RESOLVE = "（內嵌播放）"
		ACTIVITY = "私訊 `/dm [頻道ID] [訊息ID]` 刪除我的訊息！"
		HOT = "⬆️⬆️⬆️ 🔥大🔥熱🔥門🔥⎝༼ ◕д ◕ ༽⎠"
		USAGE = `
		
		`
		break
	// case EN:
	// 	HELLO = "*wake up*"
	// 	BUTTON_NOT_SONG = "Non-Song"
	// 	BUTTON_ORIGINAL = "Original"
	// 	BUTTON_COVER = "Cover"
	// 	DETECTED = "Embed detected: %s\nAccording to the title, assuming it's **%s**, redirecting in %d minutes.\nReact to suggest: 🇴 Original / 🇨 Cover / 🇸 Stream / ❌ Non-Song"
	// 	DETECTED_MATCH_NONE = "Embed detected: %s\nKeyword not found in the title, React to suggest: 🇴 Original / 🇨 Cover / 🇸 Stream / ❌ Non-Song in %d minutes"
	// 	DETECTED_UNKNOWN = "Embed detected: %s\nFailed to guess. React to suggest: 🇴 Original / 🇨 Cover / 🇸 Stream / ❌ Non-Song in %d minutes"
	// 	FAILED_TO_GUESS = "[Failed to guess]]"
	// 	REDIRECT_FORMAT = "Sharer：%s\nSource：%s"
	// 	C_DESC = "Setup channels"
	// 	C_COVER_DESC = "ID of channel for cover songs"
	// 	C_ORIGINAL_DESC = "ID of channel for original songs"
	// 	ORIGINAL = "Original"
	// 	COVER = "Cover"
	// 	STREAM = "Stream"
	// 	SHARER = "Sharer"
	// 	SMSG = "Origin"
	// 	EXPLAIN_EMBED_RESOLVE = "(Playable Embed)"
	// 	break
	}
}
