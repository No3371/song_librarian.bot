package main

import (
	"bufio"
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"No3371.github.com/song_librarian.bot/binding"
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/memory"
	"No3371.github.com/song_librarian.bot/redirect"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/dlclark/regexp2"
)

var fetchDelayTimer *time.Timer = time.NewTimer(time.Second * 3)
var buffer chan *mHandleSession

func init () {
	buffer = make(chan *mHandleSession, 128)
}

type mHandleSession struct {
	notUnlinked bool
	randomId string
	bindingId int
	cId discord.ChannelID
	mId discord.MessageID
	msg *discord.Message
	setTypes []redirect.RedirectType
	spolierLines []string
}

var meIdString string
var meIDCahce discord.UserID

func addEventHandlers (s *state.State) (err error){
	me, err := s.Me()
	if err != nil {
		return err
	}

	meIDCahce = me.ID
	meIdString = fmt.Sprintf("%d", meIDCahce)

	s.AddHandler(func (e *gateway.MessageCreateEvent) {
		if e.Author.Bot {
			return
		}
		atomic.AddUint64(&statSession.MessageEvents, 1)
		
		var mentioned bool
		
		for _, m := range e.Mentions {
			if m.ID == meIDCahce {
				mentioned = true
			}
		}
		if !mentioned {
			if sub, iErr := getSubState(e.Author.ID); iErr != nil {
				logger.Logger.Errorf("Error when getSubState(): %v", iErr)
			} else if !sub {
				atomic.AddUint64(&statSession.UnSubbedSkips, 1)
				return;
			}
		}


		buffer<-&mHandleSession{
			randomId: getRandomID(3),
			cId: e.Message.ChannelID,
			mId: e.Message.ID,
			msg: &e.Message,
		}
	})

	go handlerLoop(s)
	return nil
}

func handlerLoop (s *state.State) {
	var err error
	for {
		item := <-buffer
		atomic.AddUint64(&statSession.MessageBuffered, 1)
		
		if match, err := regexLinks.MatchString(item.msg.Content); err != nil {
			logger.Logger.Errorf("Failed to pre-filter: %v", err)
			item.notUnlinked = true
		} else {
			item.notUnlinked = match
		}

		if item.notUnlinked {
			item.msg, err = s.Message(item.cId, item.mId)
			if item.msg == nil || err != nil {
				logger.Logger.Infof("[HANDLER] [%s] Buffered item invalid. error? %v", item.randomId, err)
				continue
			}
			if len(item.msg.Embeds) == 0 {
				atomic.AddUint64(&statSession.FirstFetchEmbeds0, 1)
			}
		} else {
			atomic.AddUint64(&statSession.Unlinked, 1)
		}


		if item.notUnlinked && len(item.msg.Embeds) == 0 {
			<-fetchDelayTimer.C
			fetchDelayTimer.Reset(time.Second * 3)
			item.msg, err = s.Message(item.cId, item.mId)
			if err != nil || item.msg == nil {
				logger.Logger.Errorf("[%s] Retrying because failed to fetch message: %v", item.randomId, err)
				item.msg, err = s.Message(item.cId, item.mId)
				if err != nil || item.msg == nil {
					logger.Logger.Errorf("[%s] Retry failed too, ABORT: %v", item.randomId, err)
					continue
				}
			}
			if len(item.msg.Embeds) == 0 {
				atomic.AddUint64(&statSession.SecondFetchEmbeds0, 1)
			}
		}

		err = onMessageCreated(s, item)
		if err != nil {
			logger.Logger.Errorf("[HANDLER] [%s] OnMessageCreated error: %v", item.randomId, err)
		}
	}

}

func onMessageCreated (s *state.State, task *mHandleSession) (err error) {
	defer func () {
		if p := recover(); p != nil {
			logger.Logger.Errorf("%s", p)
		}
	}()

	if bIds := binding.GetMappedBindingIDs(uint64(task.msg.ChannelID)); bIds != nil {
		for bId := range bIds {
			b := binding.QueryBinding(bId)
			if b == nil {
				logger.Logger.Errorf("[%s] A binding Id is pointing to nil Binding: %d", task.randomId, bId)
				continue
			}

			if task.notUnlinked && len(task.msg.Embeds) == 0 {
				<-fetchDelayTimer.C
				fetchDelayTimer.Reset(time.Second * 3)
				task.msg, err = s.Message(task.msg.ChannelID, task.msg.ID)
				if task.msg == nil || err != nil {
					logger.Logger.Errorf("[%s] The message is gone!? Abort!\n%v (%s)", task.randomId, err, )
					return
				}
				if len(task.msg.Embeds) == 0 {
					atomic.AddUint64(&statSession.ThirdFetchEmbeds0, 1)
				}
			}
			
			atomic.AddUint64(&statSession.BoundChannelMessage, 1)
			if len(task.msg.Embeds) > 0 {
				atomic.AddUint64(&statSession.FetchedAndAnalyzed, 1)
			}

			scanner := bufio.NewScanner(strings.NewReader(task.msg.Content))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "/spoiler") || strings.Count(line, "||") > 1 {
					task.spolierLines = append(task.spolierLines, line)
				}

				var mentionMatch *regexp2.Match
				mentionMatch, err = regexMention.FindStringMatch(line)
				if err != nil || mentionMatch == nil {
					continue
				}
				if mentionMatch.GroupCount() < 2 {
					continue
				}
				idMentioned := mentionMatch.Groups()[1].String()
				if idMentioned == meIdString {
					for _, flag := range strings.Fields(line)[1:] {
						switch flag {
						case "o":
							task.setTypes = append(task.setTypes, redirect.Original)
							break
						case "c":
							task.setTypes = append(task.setTypes, redirect.Cover)
							break
						case "s":
							task.setTypes = append(task.setTypes, redirect.Stream)
							break
						case "x":
							task.setTypes = append(task.setTypes, redirect.None)
							break
						default:
							task.setTypes = append(task.setTypes, redirect.Unknown)
							break
						}
					}
				}
			}
	
			if notSkipLinks, err := regexNotSkipLinks.MatchString(task.msg.Content); err != nil {
				logger.Logger.Errorf("Failed to check skip link: %v", err)
			} else if notSkipLinks {
				// Find all bindings bound
				// For each binding, for each redirection, if the regex match...
				for ei, e := range task.msg.Embeds {
					logger.Logger.Infof("💬 %s-%d / %s (%s)", task.randomId, ei, e.Title, e.URL)
					atomic.AddUint64(&statSession.AnalyzedEmbeds, 1)
					urlMatching: for i := 0; i < urlRegexCount; i++ {
						if b.UrlRegexEnabled(i) {
							if isMatch, _ := regexUrlMapping[i].MatchString(e.URL); isMatch {
								atomic.AddUint64(&statSession.UrlRegexMatched, 1)
								task.bindingId = bId
								if err := pendEmbed(s, task, ei, bId); err != nil {
									logger.Logger.Errorf("%s", err)
									continue
								} else {
									atomic.AddUint64(&statSession.Pended, 1)
								}
								break urlMatching
							}
						}
					}
				}
			} else if task.notUnlinked {
				atomic.AddUint64(&statSession.SkippedLinks, 1)
			}
		}
	}

	return nil
}

func pendEmbed (s *state.State, task *mHandleSession, eIndex int, bId int) (err error) {
	embed := task.msg.Embeds[eIndex]

	var sendMessageData api.SendMessageData = api.SendMessageData{
		Reference: &discord.MessageReference{ MessageID: task.msg.ID},
	}
	var guessType redirect.RedirectType

	guessType, err = guess(task, eIndex)
	if err != nil {
		logger.Logger.Errorf("  Guessing error: %v", err)
		guessType = redirect.Unknown
	}

	_binding := binding.QueryBinding(task.bindingId)
	last, lastTouched := _binding.GetLastState(embed.URL)
	var lastSharedTime, lastRedirectTime time.Time
	lastSharedTime = _binding.GetLastTime(embed.URL, memory.Shared)
	lastRedirectTime = _binding.GetLastTime(embed.URL, memory.Redirected)
	logger.Logger.Infof("  [%s-%d] Memory: %s (%s) | %d", task.randomId, eIndex, memory.MemStateToString(last), lastTouched.Sub(time.Now()), _binding.MemPointer)

	var redirectedRecently bool = false
	if time.Now().Sub(lastRedirectTime) < *globalFlags.cooldown {
		redirectedRecently = true
	}

	var delay time.Duration = *globalFlags.delay

	preType := redirect.Unknown
	if len(task.setTypes) > eIndex {
		preType = task.setTypes[eIndex]
	}

	var hasSpoilerTag bool = len(task.spolierLines) > 0 // ! Embed links are resolved by Discord, the link may be different from the original shared one

	if hasSpoilerTag {
		logger.Logger.Infof("  Spoiler tag detected." )
		delay = delay * 4 / 10
		passed := time.Now().Sub(lastTouched).Round(time.Second)
		if redirectedRecently {
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_SPOILER_DUPLICATE, embed.Title, passed, delay.Seconds())
		} else {
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_SPOILER, embed.Title, delay.Seconds())
		}
		guessType = redirect.None
	} else if preType != redirect.Unknown { // If not UNKNOWN, accept the preype
		var typeLocale string

		if redirectedRecently {
			switch preType {
			case redirect.Original:
				typeLocale = locale.ORIGINAL_UNSIGNED // UNSIGNED
				logger.Logger.Infof("  pre_typed: o" )
				break
			case redirect.Cover:
				typeLocale = locale.COVER_UNSIGNED // UNSIGNED
				logger.Logger.Infof("  pre_typed: c" )
				break
			case redirect.Stream:
				typeLocale = locale.STREAM_UNSIGNED // UNSIGNED
				logger.Logger.Infof("  pre_typed: s" )
				break
			case redirect.None:
				typeLocale = locale.DO_NOT_REDIRECT
				logger.Logger.Infof("  pre_typed: x" )
				break
			}
			delay = delay * 5 / 10
			passed := time.Now().Sub(lastTouched).Round(time.Second)
			if guessType == preType {
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_PRE_TYPED_AGREED_DUPLICATE, embed.Title, typeLocale, passed, delay.Seconds())
			} else {
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_PRE_TYPED_DUPLICATE, embed.Title, typeLocale, passed, delay.Seconds())
			}
		} else {
			switch preType {
			case redirect.Original:
				typeLocale = locale.ORIGINAL
				logger.Logger.Infof("  pre_typed: o" )
				break
			case redirect.Cover:
				typeLocale = locale.COVER
				logger.Logger.Infof("  pre_typed: c" )
				break
			case redirect.Stream:
				typeLocale = locale.STREAM
				logger.Logger.Infof("  pre_typed: s" )
				break
			case redirect.None:
				typeLocale = locale.DO_NOT_REDIRECT
				logger.Logger.Infof("  pre_typed: x" )
				break
			}
			if guessType == preType {
				if preType == redirect.None {
					delay = delay * 3 / 10
				} else {
					delay = delay * 5 / 10
				}
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_PRE_TYPED_AGREED, embed.Title, typeLocale, delay.Seconds())
			} else {
				if preType == redirect.None {
					delay = delay * 4 / 10
				} else {
					delay = delay * 7 / 10
				}
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_PRE_TYPED, embed.Title, typeLocale, delay.Seconds())
			}
		}


	} else {
		if redirectedRecently {
			passed := time.Now().Sub(lastRedirectTime).Round(time.Second)
			switch guessType {
			case redirect.Original:
				delay = delay / 2
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_DUPLICATE, embed.Title, locale.ORIGINAL_UNSIGNED, passed ,delay.Seconds())
				break
			case redirect.Cover:
				delay = delay / 2
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_DUPLICATE, embed.Title, locale.COVER_UNSIGNED, passed, delay.Seconds())
				break
			case redirect.Stream:
				delay = delay / 2
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_DUPLICATE, embed.Title, locale.STREAM_UNSIGNED, passed, delay.Seconds())
				break
			case redirect.None:
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_DUPLICATE_NONE, embed.Title, passed, delay.Seconds())
				break
			case redirect.Unknown:
				delay = delay * 3 / 2
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_UNKNOWN_DUPLICATE, embed.Title, passed, delay.Seconds())
				break
			case redirect.Clip:
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_CLIPS_DUPLICATE, embed.Title, passed, delay.Seconds())
				guessType = redirect.None
			}

		} else {
			passed := time.Now().Sub(lastTouched).Round(time.Second)
			switch guessType {
			case redirect.Original:
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.ORIGINAL, delay.Seconds())
			case redirect.Cover:
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.COVER, delay.Seconds())
				break
			case redirect.Stream:
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.STREAM, delay.Seconds())
				break
			case redirect.None:
				if last == memory.Cancelled {
					delay = delay / 2
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_MATCH_NONE_AND_CANCELLED, embed.Title, passed, delay.Seconds())
				} else if last > memory.None && lastTouched.Before(lastSharedTime) {
					delay = delay * 2 / 3
					passed = time.Now().Sub(lastSharedTime).Round(time.Second)
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_MATCH_NONE_AND_SHARED, embed.Title, passed, delay.Seconds())
				} else {
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_MATCH_NONE, embed.Title, delay.Seconds())
				}
				break
			case redirect.Unknown:
				sendMessageData.Content = fmt.Sprintf(locale.DETECTED_UNKNOWN, embed.Title, delay.Seconds())
				break
			case redirect.Clip:
				if last == memory.Cancelled {
					delay = delay / 3
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_CLIPS_AND_CANCELLED, embed.Title, passed, delay.Seconds())
				} else {
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_CLIPS, embed.Title, delay.Seconds())
				}
				guessType = redirect.None
			}

		}
	}

	if *globalFlags.dev {
		delay /= 4
	}
	if err = _binding.Memorize(embed.URL, memory.Pended); err != nil {
		logger.Logger.Errorf("  [%s] Failed to memorized(Pended): %v", task.randomId, err)
	}
	botM, err := s.SendMessageComplex(task.msg.ChannelID, sendMessageData)
	if err != nil {
		logger.Logger.Errorf("[%s-%d] %v", task.randomId, eIndex, err)
		return err
	}

	// var reaction string
	// switch rType {
	// case redirect.Original:
	// 	reaction = reactionOriginal
	// 	break
	// case redirect.Cover:
	// 	reaction = reactionCover
	// 	break
	// case redirect.Stream:
	// 	reaction = reactionStream
	// 	break
	// case redirect.None:
	// 	reaction = reactionNone
	// 	break
	// case redirect.Unknown:
	// 	break
	// }
	
	// logger.Logger.Infof("  Reacting...")
	// err = s.React(botM.ChannelID, botM.ID, discord.APIEmoji(reactionOriginal))
	// if err != nil {
	// 	log.Printf("[Error] %s", fmt.Errorf("%w", err))
	// 	return err
	// }
	// err = s.React(botM.ChannelID, botM.ID, discord.APIEmoji(reactionCover))
	// if err != nil {
	// 	log.Printf("[Error] %s", fmt.Errorf("%w", err))
	// 	return err
	// }
	// err = s.React(botM.ChannelID, botM.ID, discord.APIEmoji(reactionStream))
	// if err != nil {
	// 	log.Printf("[Error] %s", fmt.Errorf("%w", err))
	// 	return err
	// }
	// err = s.React(botM.ChannelID, botM.ID, discord.APIEmoji(reactionNone))
	// if err != nil {
	// 	log.Printf("[Error] %s", fmt.Errorf("%w", err))
	// 	return err
	// }
	pendingEmbeds<-&pendingEmbed{
		cId: botM.ChannelID,
		msgID: botM.ID,
		embedIndex: eIndex,
		urlValidation: task.msg.Embeds[eIndex].URL,
		taskId: task.randomId,
		bindingId: bId,
		estimatedRTime: time.Now().Add(delay),
		autoType: guessType,
		preType: preType,
		isDup: redirectedRecently,
		spoiler: hasSpoilerTag,
	}

	if hasSpoilerTag {
		atomic.AddUint64(&statSession.PendedSpoilerFlag, 1)
	}

	if err = _binding.Memorize(embed.URL, memory.Pended); err != nil {
		logger.Logger.Errorf("  [%s] Failed to memorized(Pended): %v", task.randomId, err)
	}

	return nil
}


func guess (task *mHandleSession, eIndex int) (redirectType redirect.RedirectType, err error) {
	embed := task.msg.Embeds[eIndex]
	defer func () {
		if redirectType == redirect.None || redirectType == redirect.Unknown || redirectType == redirect.Clip {
			return
		}

		if likeAClip, _ := regexClips.MatchString(embed.Title); likeAClip {
			logger.Logger.Infof("  [GUESS] Wait! It looks like a clip!")
			redirectType = redirect.Clip
		}

		if likeAClip, _ := regexClips.MatchString(embed.Author.Name); likeAClip {
			logger.Logger.Infof("  [GUESS] Wait! The author looks like a clipping channel!")
			redirectType = redirect.Clip
		}

	} ()
	
	sb := &strings.Builder{}

	var m *regexp2.Match
	m, err = regexNamedCover.FindStringMatch(embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for NamedCover keywords: %v", err)
		return redirect.Unknown, err
	} else if m != nil {
		logger.Logger.Infof("  Matched named cover: %s", m)
		return redirect.Cover, nil
	}

	m, err = regexNamedStream.FindStringMatch(embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for NamedStream keywords: %v", err)
		return redirect.Unknown, err
	} else if m != nil {
		logger.Logger.Infof("  Matched named cover: %s", m)
		return redirect.Stream, nil
	}

	var countO = 0
	var countOPlus = 0
	var countNotO = 0
	var countDNotO = 0
	var countC = 0
	var countCPlus = 0
	var countDC = 0
	var countNotC = 0
	var countS = 0
	var countSPlus = 0
	var countNotS = 0
	var countBadForAll = 0

	// ! NOTE: Discord does not provide full description

	countC, err = countMatch(sb, "Cover", regexCover_s0, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Cover keywords: %v", err)
		return redirect.Unknown, err
	}

	countCPlus, err = countMatch(sb, "Cover+", regexCoverPlus, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Cover keywords: %v", err)
		return redirect.Unknown, err
	}

	countDC, err = countMatch(sb, "D.Cover", regexCoverDesc, embed.Description)
	if err != nil {
		logger.Logger.Errorf("Failed to match for D.Cover keywords: %v", err)
		return redirect.Unknown, err
	}
	if countDC >= 2 {
		countDC /= 2
	}

	countNotC, err = countMatch(sb, "NotCover", regexBadForCover, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for NotCover keywords: %v", err)
		return redirect.Unknown, err
	}

	countO, err = countMatch(sb,"Original", regexOriginal_s1, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Original keywords: %v", err)
		return redirect.Unknown, err
	}

	countOPlus, err = countMatch(sb,"Original+", regexOriginalPlus, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Original keywords: %v", err)
		return redirect.Unknown, err
	}

	countNotO, err = countMatch(sb, "NotOriginal", regexBadForOriginal, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for NotOriginal keywords: %v", err)
		return redirect.Unknown, err
	}

	countDNotO, err = countMatch(sb, "D.NotOriginal", regexBadDescForOriginal, embed.Description)
	if err != nil {
		logger.Logger.Errorf("Failed to match for D.NotOriginal keywords: %v", err)
		return redirect.Unknown, err
	}
	if countDNotO >= 2 {
		countDNotO /= 2
	}

	countS, err = countMatch(sb,"Stream", regexStream_s2, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Stream keywords: %v", err)
		return redirect.Unknown, err
	}

	countSPlus, err = countMatch(sb,"Stream+", regexStreamPlus, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Stream keywords: %v", err)
		return redirect.Unknown, err
	}

	countNotS, err = countMatch(sb, "NotStream", regexBadForStream, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for NotStream keywords: %v", err)
		return redirect.Unknown, err
	}

	countBadForAll, err = countMatch(sb, "BadForAll", regexBadForAll, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for BadForAll keywords: %v", err)
		return redirect.Unknown, err
	}

	scoreO := countO + countOPlus - countNotO - countDNotO - countBadForAll
	if scoreO < 0 {
		scoreO = 0
	}
	scoreC := countC + countDC + countCPlus - countNotC - countBadForAll
	if scoreC < 0 {
		scoreC = 0
	}
	scoreS := countS + countSPlus - countNotS - countBadForAll
	if scoreS < 0 {
		scoreS = 0
	}

	defer func () {
		logger.Logger.Infof("  [GUESS-%s-%d] %s", task.randomId, eIndex, sb.String())
		logger.Logger.Infof("  [GUESS-%s-%d] %s | o%d=%d+%d-%d-%d c%d=%d+%d+%d-%d s%d=%d+%d-%d (ocs-%d)", task.randomId, eIndex,
			redirect.RedirectTypetoString(redirectType),
			scoreO, countO, countOPlus, countNotO, countDNotO,
			scoreC, countC, countCPlus, countDC, countNotC,
			scoreS, countS, countSPlus, countNotS,
			countBadForAll,
		)
	} ()

	if scoreO <= 0 && scoreC <= 0 && scoreS <= 0 {
		return redirect.None, nil
	}

	if countC == countO && countO == countS {
		return redirect.Unknown, nil
	}

	if countC == 0 && countO == 0 && countS == 0 {
		countDC = 0
		countDNotO = 0
	}

	if scoreC > scoreO && scoreC > scoreS {
		return redirect.Cover, nil
	}

	if scoreS > scoreO && scoreS > scoreC {
		return redirect.Stream, nil
	}

	if scoreO >= scoreC && scoreO > scoreS {
		if scoreS > 0 && scoreO < 2 && scoreC < 2 {
			return redirect.Stream, nil // It seems like people go like 'Today I'll cover XX's original songs...'
		}

		if scoreC > 0 && scoreO < 3 { // If there's a keywork of Cover, it's very unlikely it'd be Original
			return redirect.Cover, nil
		}

		return redirect.Original, nil
	}

	if scoreC > 0 && scoreS > 0 && scoreC >= scoreS{ // If it looks like a stream or cover, Cover is prefered (Usually, a stream is not titled with keyword of Cover)
		return redirect.Cover, nil
	}

	return redirect.Unknown, nil
}


func countMatch (sb *strings.Builder, regexType string, r *regexp2.Regexp, subject string) (count int, err error) {
	var m *regexp2.Match
	m, err = r.FindStringMatch(subject)
	if err != nil || m == nil {
		return count, err
	}

	count++
	sb.WriteString(fmt.Sprintf("|   (%s) %s", regexType, m.String()))
	for m != nil {
		m, err = r.FindNextMatch(m)
		if err != nil {
			logger.Logger.Errorf("  %s", err)
			return count, err
			} else if m != nil {
				count++
				sb.WriteString(fmt.Sprintf(" / %s", m.String()))
			}
		}

	return count, err
}