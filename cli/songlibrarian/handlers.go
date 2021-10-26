package main

import (
	"bufio"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"No3371.github.com/song_librarian.bot/binding"
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
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
	rId string
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
			rId: getRandomID(3),
			cId: e.Message.ChannelID,
			mId: e.Message.ID,
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
		item.msg, err = s.Message(item.cId, item.mId)
		if item.msg == nil || err != nil {
			logger.Logger.Infof("[HANDLER] [%s] Buffered item invalid. error? %v", item.rId, err)
			continue
		}

		if len(item.msg.Embeds) == 0 {
			atomic.AddUint64(&statSession.FirstFetchEmbeds0, 1)
		}

		if time.Since(item.msg.Timestamp.Time().Local()) < time.Second*2 {
			<-fetchDelayTimer.C
			fetchDelayTimer.Reset(time.Second * 3)
			item.msg, err = s.Message(item.cId, item.mId)
			if err != nil || item.msg == nil {
				logger.Logger.Errorf("[%s] Retrying because failed to fetch message: %v", item.rId, err)
				item.msg, err = s.Message(item.cId, item.mId)
				if err != nil || item.msg == nil {
					logger.Logger.Errorf("[%s] Retry failed too, ABORT: %v", item.rId, err)
					continue
				}
			}
			if len(item.msg.Embeds) == 0 {
				atomic.AddUint64(&statSession.SecondFetchEmbeds0, 1)
			}
		}

		err = onMessageCreated(s, item)
		if err != nil {
			logger.Logger.Errorf("[HANDLER] [%s] OnMessageCreated error: %v", item.rId, err)
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
				logger.Logger.Errorf("[%s] A binding Id is pointing to nil Binding: %d", task.rId, bId)
				continue
			}

			if len(task.msg.Embeds) == 0 {
				<-fetchDelayTimer.C
				fetchDelayTimer.Reset(time.Second * 3)
				task.msg, err = s.Message(task.msg.ChannelID, task.msg.ID)
				if task.msg == nil || err != nil {
					logger.Logger.Errorf("[%s] The message is gone!? Abort!\n%v (%s)", task.rId, err, )
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
	
			// Find all bindings bound
			// For each binding, for each redirection, if the regex match...
			for ei, e := range task.msg.Embeds {
				logger.Logger.Infof("💬 %s (%s) / %s-%d", e.Title, e.URL, task.rId, ei)
				atomic.AddUint64(&statSession.AnalyzedEmbeds, 1)
				urlMatching: for i := 0; i < urlRegexCount; i++ {
					if b.UrlRegexEnabled(i) {
						if isMatch, _ := regexUrlMapping[i].MatchString(e.URL); isMatch {
							atomic.AddUint64(&statSession.UrlRegexMatched, 1)
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
		}
	}

	return nil
}

func pendEmbed (s *state.State, task *mHandleSession, eIndex int, bId int) (err error) {
	embed := task.msg.Embeds[eIndex]

	var sendMessageData api.SendMessageData = api.SendMessageData{
		Reference: &discord.MessageReference{ MessageID: task.msg.ID},
	}
	var autoType redirect.RedirectType

	var lastProcessedTime, lastSharedTime, lastRedirectTime, lastResultTime time.Time
	var lastMem memState
	lastMem, lastResultTime = getLastState(embed.URL)
	_, lastSharedTime = getLastShared(embed.URL)
	_, lastRedirectTime = getLastRedirected(embed.URL)
	_, lastProcessedTime = getLastResult(embed.URL)
	logger.Logger.Infof("  [%s] Memory: %s (%s) | %d", task.rId, memStateToString(lastMem), lastResultTime.Sub(time.Now()), memPointer)
	
	var redirectedRecently bool = false
	if (lastMem == Redirected || lastMem == Cancelled) && time.Now().Sub(lastRedirectTime) < time.Hour * 24 {
		redirectedRecently = true
	}

	autoType, err = guess(task, eIndex)
	if err != nil {
		logger.Logger.Errorf("  Guessing error: %v", err)
		autoType = redirect.Unknown
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
		passed := time.Now().Sub(lastResultTime).Round(time.Second)
		if redirectedRecently {
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_SPOILER_DUPLICATE, embed.Title, passed, delay.Seconds())
		} else {
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_SPOILER, embed.Title, delay.Seconds())
		}
		autoType = redirect.None
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
			passed := time.Now().Sub(lastResultTime).Round(time.Second)
			if autoType == preType {
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
			if autoType == preType {
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
		passed := time.Now().Sub(lastResultTime).Round(time.Second)
		if redirectedRecently {
			switch autoType {
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
				autoType = redirect.None
			}

		} else {
			switch autoType {
			case redirect.Original:
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.ORIGINAL, delay.Seconds())
			case redirect.Cover:
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.COVER, delay.Seconds())
				break
			case redirect.Stream:
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.STREAM, delay.Seconds())
				break
			case redirect.None:
				if lastMem == Cancelled {
					delay = delay / 2
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_MATCH_NONE_AND_CANCELLED, embed.Title, passed, delay.Seconds())
				} else if lastMem > None && lastProcessedTime.Before(lastSharedTime) {
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
				if lastMem == Cancelled {
					delay = delay / 3
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_CLIPS_AND_CANCELLED, embed.Title, passed, delay.Seconds())
				} else {
					sendMessageData.Content = fmt.Sprintf(locale.DETECTED_CLIPS, embed.Title, delay.Seconds())
				}
				autoType = redirect.None
			}

		}
	}

	if *globalFlags.dev {
		delay /= 4
	}
	if err = memorizeShared(embed.URL); err != nil {
		logger.Logger.Errorf("  [%s] Failed to memorizedPended: %v", task.rId, err)
	}
	botM, err := s.SendMessageComplex(task.msg.ChannelID, sendMessageData)
	if err != nil {
		logger.Logger.Errorf("[%s-%d] %v", task.rId, eIndex, err)
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
		taskId: task.rId,
		bindingId: bId,
		estimatedRTime: time.Now().Add(delay),
		autoType: autoType,
		preType: preType,
		isDup: redirectedRecently,
		spoiler: hasSpoilerTag,
	}

	if hasSpoilerTag {
		atomic.AddUint64(&statSession.PendedSpoilerFlag, 1)
	}

	if err = memorizePended(embed.URL); err != nil {
		logger.Logger.Errorf("  [%s] Failed to memorizedPended: %v", task.rId, err)
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
	var countNotO = 0
	var countDNotO = 0
	var countC = 0
	var countDC = 0
	var countNotC = 0
	var countS = 0
	var countNotS = 0
	var countBadForAll = 0

	// ! NOTE: Discord does not provide full description

	countBadForAll, err = countMatch(sb, "BadForAll", regexBadForAll, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for BadForAll keywords: %v", err)
		return redirect.Unknown, err
	}

	countC, err = countMatch(sb, "Cover", regexCover_s0, embed.Title)
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

	countNotS, err = countMatch(sb, "NotStream", regexBadForStream, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for NotStream keywords: %v", err)
		return redirect.Unknown, err
	}

	if countC + countO + countS == 0 {
		logger.Logger.Infof("  [GUESS-%s-%d] o%d(-%d-%d) c%d(+%d-%d) s%d(-%d)", task.rId, eIndex, countO, countNotO, countDNotO, countC, countDC, countNotC, countS, countNotS)
		return redirect.None, nil
	} else {
		logger.Logger.Infof("  [GUESS-%s-%d] %s", task.rId, eIndex, sb.String())
		logger.Logger.Infof("  [GUESS-%s-%d] o%d(-%d-%d) c%d(+%d-%d) s%d(-%d)", task.rId, eIndex, countO, countNotO, countDNotO, countC, countDC, countNotC, countS, countNotS)
	}


	if countC == countO && countO == countS {
		return redirect.Unknown, nil
	}

	countC -= (countNotC + countBadForAll)
	countC += countDC
	countO -= (countNotO + countBadForAll + countDNotO)
	countS -= (countNotS + countBadForAll)
	if countC < 0 {
		countC = 0
	}
	if countO < 0 {
		countO = 0
	}
	if countS < 0 {
		countS = 0
	}
	
	if countC > countO && countC > countS {
		return redirect.Cover, nil
	}

	if countO > countC && countO > countS || (countO >= countC && countO > countS) {
		if countC > 0 && countO < 3 { // If there's a keywork of Cover, it's very unlikely it'd be Original
			return redirect.Cover, nil
		}
		return redirect.Original, nil
	}

	if countS > countO && countS > countC {
		return redirect.Stream, nil
	}

	if countC > 0 && countS > 0 && countC >= countS{ // If it looks like a stream or cover, Cover is prefered (Usually, a stream is not titled with keyword of Cover)
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

	count ++
	sb.WriteString(fmt.Sprintf("    (%s) %s", regexType, m.String()))
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