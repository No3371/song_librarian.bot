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
	cId discord.ChannelID
	mId discord.MessageID
	msg *discord.Message
	setTypes []redirect.RedirectType
}

var meMentionCache string
var meIDCahce discord.UserID

func addEventHandlers (s *state.State) (err error){
	me, err := s.Me()
	if err != nil {
		return err
	}

	meIDCahce = me.ID
	meMentionCache = fmt.Sprintf("<@!%d>", meIDCahce)

	s.AddHandler(func (e *gateway.MessageCreateEvent) {
		if e.Author.Bot {
			return
		}

		atomic.AddUint64(&statSession.MessageEvents, 1)

		buffer<-&mHandleSession{
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
			logger.Logger.Infof("[HANDLER] Buffered item invalid. error? %v", err)
			continue
		}

		if len(item.msg.Embeds) == 0 {
			atomic.AddUint64(&statSession.FirstFetchEmbeds0, 1)
		}

		if time.Now().Sub(item.msg.Timestamp.Time().Local()) < time.Second*2 {
			<-fetchDelayTimer.C
			fetchDelayTimer.Reset(time.Second * 2)
			item.msg, err = s.Message(item.cId, item.mId)
			if len(item.msg.Embeds) == 0 {
				atomic.AddUint64(&statSession.SecondFetchEmbeds0, 1)
			}
		}

		err = onMessageCreated(s, item)
		if err != nil {
			logger.Logger.Errorf("[HANDLER] OnMessageCreated error: %v", err)
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
				logger.Logger.Errorf("A binding Id is pointing to nil Binding: %d", bId)
				continue
			}

			if len(task.msg.Embeds) == 0 {
				<-fetchDelayTimer.C
				fetchDelayTimer.Reset(time.Second * 2)
				task.msg, err = s.Message(task.msg.ChannelID, task.msg.ID)
				if task == nil || err != nil {
					logger.Logger.Errorf("The message is gone!? Abort!\n%v", err)
					return
				}
				if len(task.msg.Embeds) == 0 {
					atomic.AddUint64(&statSession.ThirdFetchEmbeds0, 1)
				}
			}
			
			if len(task.msg.Embeds) > 0 {
				atomic.AddUint64(&statSession.FetchedAndAnalyzed, 1)
			}

			for _, mentioned := range task.msg.Mentions {
				if mentioned.ID == meIDCahce {
					scanner := bufio.NewScanner(strings.NewReader(task.msg.Content))
					for scanner.Scan() {
						line := scanner.Text()
						if strings.HasPrefix(line, meMentionCache) {
							for _, flag := range strings.Fields(line) {
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
								case "_":
									task.setTypes = append(task.setTypes, redirect.Unknown)
									break
								}
							}
						}
					}
				}
			}
	
			// Find all bindings bound
			// For each binding, for each redirection, if the regex match...
			for ei, e := range task.msg.Embeds {
				logger.Logger.Infof("💬 %s (%s) / %d #%d", e.Title, e.URL, task.msg.ID, ei)
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

func pendEmbed (s *state.State, task *mHandleSession, eIndex int, bId int) error {
	embed := task.msg.Embeds[eIndex]
	var sendMessageData api.SendMessageData = api.SendMessageData{
		Reference: &discord.MessageReference{ MessageID: task.msg.ID},
	}
	myGuess, err := guess(embed)
	if err != nil {
		logger.Logger.Errorf("  Guessing error: %v", err)
		myGuess = redirect.Unknown
	}
	var delay time.Duration = *globalFlags.delay

	preType := redirect.Unknown
	if len(task.setTypes) > eIndex {
		preType = task.setTypes[eIndex]
	}

	if preType != redirect.Unknown { // If not UNKNOWN, accept the preype
		var format, typeLocale string
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

		if myGuess == preType {
			if preType == redirect.None {
				delay = delay * 4 / 10			
			} else {
				delay = delay * 5 / 10
			}
			format = locale.DETECTED_PRE_TYPED_AGREED
		} else {
			if preType == redirect.None {
				delay = delay * 4 / 10			
			} else {
				delay = delay * 7 / 10
			}
			format = locale.DETECTED_PRE_TYPED
		}

		sendMessageData.Content = fmt.Sprintf(format, embed.Title, typeLocale, delay.Seconds())

	} else {
		switch myGuess {
		case redirect.Original:
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.ORIGINAL, delay.Seconds())
			break
		case redirect.Cover:
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.COVER, delay.Seconds())
			break
		case redirect.Stream:
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.STREAM, delay.Seconds())
			break
		case redirect.None:
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_MATCH_NONE, embed.Title, delay.Seconds())
			break
		case redirect.Unknown:
			delay = delay * 3 / 2
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_UNKNOWN, embed.Title, delay.Seconds())
			break
		case redirect.Clip:
			sendMessageData.Content = fmt.Sprintf(locale.DETECTED_CLIPS, embed.Title, delay.Seconds())
			myGuess = redirect.None
			break
		}
	}
	botM, err := s.SendMessageComplex(task.msg.ChannelID, sendMessageData)
	if err != nil {
		logger.Logger.Errorf("%v", fmt.Errorf("%w", err))
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
	logger.Logger.Infof("  Pending %d #%d...", task.msg.ID, eIndex)
	pendingEmbeds<-&pendingEmbed{
		cId: botM.ChannelID,
		msgID: botM.ID,
		embedIndex: eIndex,
		urlValidation: task.msg.Embeds[eIndex].URL,
		bindingId: bId,
		estimatedRTime: time.Now().Add(delay),
		guess: myGuess,
		preType: preType,
	}

	return nil
}


func guess (embed discord.Embed) (redirectType redirect.RedirectType, err error) {
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
	var countO = 0
	var countNotO = 0
	var countC = 0
	var countNotC = 0
	var countS = 0

	sb := &strings.Builder{}

	countC, err = countMatch(sb, "Cover", regexCover_s0, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Cover keywords: %v", err)
		return redirect.Unknown, err
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

	countS, err = countMatch(sb,"Stream", regexStream_s2, embed.Title)
	if err != nil {
		logger.Logger.Errorf("Failed to match for Stream keywords: %v", err)
		return redirect.Unknown, err
	}


	if countC + countO + countS == 0 {
		logger.Logger.Infof("  [GUESS] o%d c%d s%d", countO, countC, countS)
		return redirect.None, nil
	} else {
		logger.Logger.Infof("  [GUESS] %s", sb.String())
		logger.Logger.Infof("  [GUESS] o%d(-%d) c%d(-%d) s%d", countO, countNotO, countC, countNotC, countS)
	}


	if countC == countO && countO == countS {
		return redirect.Unknown, nil
	}

	countC -= countNotC
	countO -= countNotO

	if countC > countO && countC > countS {
		return redirect.Cover, nil
	}

	if countO > countC && countO > countS {
		if countC > 0 && countO < 3 {
			return redirect.Cover, nil
		}
		return redirect.Original, nil
	}

	if countS > countO && countS > countC {
		return redirect.Stream, nil
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