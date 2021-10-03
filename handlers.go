package main

import (
	"fmt"
	"log"
	"time"

	"No3371.github.com/song_librarian.bot/binding"
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/redirect"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func addHandlers (s *state.State) {
	s.AddHandler(func (e *gateway.MessageCreateEvent) {
		if err := onMessageCreated(s, &e.Message); err != nil {
			logger.Logger.Errorf("[HANDLER] OnMessageCreated: %v", err)
		}
	})
}

func onMessageCreated (s *state.State, m *discord.Message) (err error) {
	defer func () {
		if p := recover(); p != nil {
			logger.Logger.Errorf("%s", p)
		}
	}()

	if bIds := binding.GetMappedBindingIDs(uint64(m.ChannelID)); bIds != nil {
		for bId := range bIds {
			b := binding.QueryBinding(bId)
			if b == nil {
				logger.Logger.Errorf("A binding Id is pointing to nil Binding: %d", bId)
				continue
			}

			// Find all bindings bound
			// For each binding, for each redirection, if the regex match...
			for ei, e := range m.Embeds {
				logger.Logger.Infof("Analyzing embed: %s (%s)", e.Title, e.URL)
				urlMatching: for i := 0; i < urlRegexCount; i++ {
					if b.UrlRegexEnabled(i) {
						logger.Logger.Infof("Url regex #%d is enabled:", i)
						if isMatch, _ := regexUrlMapping[i].MatchString(e.URL); isMatch {
							logger.Logger.Infof("Binding#%d - UrlRegex#%d match!", bId, i)
							if err := pendEmbed(s, m, ei, bId); err != nil {
								logger.Logger.Errorf("%s", err)
								continue
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

func pendEmbed (s *state.State, om *discord.Message, eIndex int, bId int) error {
	embed := om.Embeds[eIndex]
	var sendMessageData api.SendMessageData = api.SendMessageData{
		Reference: &discord.MessageReference{ MessageID: om.ID},
	}
	logger.Logger.Infof("  Guessing...")
	rType, err := guess(embed)
	switch rType {
	case redirect.Original:
		sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.ORIGINAL, delay)
		break
	case redirect.Cover:
		sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.COVER, delay)
		break
	case redirect.Stream:
		sendMessageData.Content = fmt.Sprintf(locale.DETECTED, embed.Title, locale.STREAM, delay)
		break
	case redirect.None:
		sendMessageData.Content = fmt.Sprintf(locale.DETECTED_MATCH_NONE, embed.Title, delay)
		break
	case redirect.Unknown:
		break
	}
	logger.Logger.Infof("  Sending...")
	botM, err := s.SendMessageComplex(om.ChannelID, sendMessageData)
	if err != nil {
		log.Printf("[Error] %s", fmt.Errorf("%w", err))
		return err
	}

	var reaction string
	switch rType {
	case redirect.Original:
		reaction = reactionOriginal
		break
	case redirect.Cover:
		reaction = reactionCover
		break
	case redirect.Stream:
		reaction = reactionStream
		break
	case redirect.None:
		reaction = reactionNone
		break
	case redirect.Unknown:
		break
	}
	
	logger.Logger.Infof("  Reacting...")
	err = s.React(botM.ChannelID, botM.ID, discord.APIEmoji(reaction))
	if err != nil {
		log.Printf("[Error] %s", fmt.Errorf("%w", err))
		return err
	}

	logger.Logger.Infof("  Pending...")
	pendingEmbeds<-&pendingEmbed{
		cId: botM.ChannelID,
		msgID: botM.ID,
		embedIndex: eIndex,
		bindingId: bId,
		pendedTime: time.Now(),
	}

	return nil
}


func guess (embed discord.Embed) (redirectType redirect.RedirectType, err error) {
	var countOriginalKeywords = 0
	var countCoverKeywords = 0
	var countStreamKeywords = 0

	m, err := regexCover_s0.FindStringMatch(embed.Title)
	if err != nil {
		logger.Logger.Errorf("%s", err)
		return redirect.Unknown, err
	}

	if m != nil {
		countCoverKeywords ++
		for m != nil {
			m, err = regexCover_s0.FindNextMatch(m)
			if err != nil {
				logger.Logger.Errorf("%s", err)
				return redirect.Unknown, err
			}
			countCoverKeywords++
			logger.Logger.Info("Cover+1")
		}
	}

	m, err = regexOriginal_s1.FindStringMatch(embed.Title)
	if err != nil {
		logger.Logger.Errorf("%s", err)
		return redirect.Unknown, err
	}
	if m != nil {
		countOriginalKeywords ++
		for m != nil {
			m, err = regexOriginal_s1.FindNextMatch(m)
			if err != nil {
				logger.Logger.Errorf("%s", err)
				return redirect.Unknown, err
			}
			countOriginalKeywords++
			logger.Logger.Info("Original+1")
		}
	}

	m, err = regexStream_s2.FindStringMatch(embed.Title)
	if err != nil {
		logger.Logger.Errorf("%s", err)
		return redirect.Unknown, err
	}
	if m != nil {
		countStreamKeywords ++
		for m != nil {
			m, err = regexStream_s2.FindNextMatch(m)
			if err != nil {
				logger.Logger.Errorf("%s", err)
				return redirect.Unknown, err
			}
			countStreamKeywords++
			logger.Logger.Info("Stream+1")
		}
	}

	if countCoverKeywords + countOriginalKeywords + countStreamKeywords == 0 {
		return redirect.None, nil
	}

	if countCoverKeywords == countOriginalKeywords && countOriginalKeywords == countStreamKeywords {
		return redirect.Unknown, nil
	}

	if countCoverKeywords > countOriginalKeywords && countCoverKeywords > countStreamKeywords {
		return redirect.Cover, nil
	}

	if countOriginalKeywords > countCoverKeywords && countOriginalKeywords > countStreamKeywords {
		return redirect.Original, nil
	}

	if countStreamKeywords > countOriginalKeywords && countStreamKeywords > countCoverKeywords {
		return redirect.Stream, nil
	}

	return redirect.Unknown, nil
}