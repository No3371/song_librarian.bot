package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync/atomic"

	// "os/signal"
	// "syscall"

	// "fmt"
	"log"
	"time"

	"No3371.github.com/song_librarian.bot/binding"
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/redirect"
	"No3371.github.com/song_librarian.bot/storage"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/pkg/errors"
)


var pendingEmbeds chan *pendingEmbed
var sv storage.StorageProvider

type pendingEmbed struct {
	cId discord.ChannelID
	msgID discord.MessageID
	embedIndex int
	urlValidation string
	taskId string
	bindingId int
	estimatedRTime time.Time
	autoType, preType redirect.RedirectType
	isDup bool
}

type resultType int

const (
	RESULT_BOT_GUESS resultType = iota
	RESULT_BOT_GUESS_AGREED
	RESULT_BOT_GUESS_FIXED
	RESULT_COMMUNITY
	RESULT_SHARER
	RESULT_SHARER_FIXED
	RESULT_SHARER_AND_COMMUNITY
	RESULT_SHARER_AND_BOT
	RESULT_SHARER_AND_BOT_AND_COMMUNITY
	FALLBACK
)


var processCloser chan struct{}
func main() {
	var err error
	resolveFlags()
	logger.SetupLogger(!*globalFlags.dev)
	locale.SetLanguage(locale.FromString(*globalFlags.locale))
	statSession = &stats{
		StartAt: time.Now(),
	}

	if *globalFlags.dev {
        f, err := os.Create("./cpuprof")
        if err != nil {
            logger.Logger.Fatalf("could not create CPU profile: ", err)
        }
        defer f.Close() // error handling omitted for example
		logger.Logger.Infof("Starting cpu profiling...")
		runtime.SetCPUProfileRate(200)
        if err := pprof.StartCPUProfile(f); err != nil {
            logger.Logger.Fatalf("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }
	// sigs := make(chan os.Signal, 1)

    // signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	
	

	processCloser = make(chan struct{})

	var restart bool = true
	var sessionCloser chan struct{} = make(chan struct{})
    go func() {
        // sig := <-sigs
        <-processCloser
		// logger.Logger.Infof("Captured EXIT signal: %s", sig)
		restart = false
		go func () {
			timer := time.NewTimer(time.Second * 10)
			<-timer.C
			logger.Logger.Infof("[MAIN] === Force bailing out ===")
			os.Exit(1)
		} ()
        close(sessionCloser)
    }()
	loopTimer := time.NewTimer(time.Second * 10)
	for restart {
		if err != nil {
			logger.Logger.Errorf("[MAIN] Previous session closed with error: %v", err)
			logger.Logger.Infof("[MAIN] Restarting...")
		}
		err = nil
		err = session(sessionCloser)
		<-loopTimer.C
		loopTimer.Reset(time.Second * 10)
	}
	statSession.Print()
	logger.Logger.Infof("SAVE ALL running")
	binding.SaveAll()
	logger.Logger.Infof("SAVE ALL finished")
	os.Exit(0)
}

func session (sCloser chan struct{}) (err error) {
	var sessionSelfCloser chan struct{} = make(chan struct{})
	logger.Logger.Infof("[MAIN] Session is starting...")
	sv, err = storage.Sqlite()
	if err != nil {
		return errors.Wrap(err, "Failed to get storage")
	}

	binding.Setup(sv)

	s, err := state.New("Bot " + *globalFlags.token)
	if err != nil {
		return errors.Wrap(err, "Failed to get new bot state")
	}

	s.AddIntents(gateway.IntentDirectMessages)
	s.AddIntents(gateway.IntentGuildMessages)
	s.AddIntents(gateway.IntentGuildMessageReactions)

	pendingEmbeds = make(chan *pendingEmbed, 512)

	redirectorClosed := redirectorLoop(s, sessionSelfCloser)

	err = assureCommands(s)
	if err != nil {
		logger.Logger.Fatalf("[MAIN] %v", err)
	}

	addEventHandlers(s)
	addInteractionHandlers(s)

	s.ErrorLog = func(innerErr error) {
		logger.Logger.Errorf("[MAIN] Gateway error: %v", innerErr)
		err = innerErr
		select {
		case <-sessionSelfCloser:
		default:
			close(sessionSelfCloser)
		}
	}

	s.FatalErrorCallback = func(innerErr error) {
		logger.Logger.Errorf("[MAIN] Fatal gateway error: %v", err)
		err = innerErr
		select {
		case <-sessionSelfCloser:
		default:
			close(sessionSelfCloser)
		}
	}

	s.AfterClose = func(innerErr error) {
		logger.Logger.Errorf("[MAIN] After gateway closed: %v", err)
		err = innerErr
		select {
		case <-sessionSelfCloser:
		default:
			close(sessionSelfCloser)
		}
	}
	
	if err := s.Open(context.Background()); err != nil {
		logger.Logger.Errorf("[MAIN] %v", err)
		select {
		case <-sessionSelfCloser:
		default:
			close(sessionSelfCloser)
		}
	}
	defer s.Close()

	u, err := s.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}
	
	s.UpdateStatus(gateway.UpdateStatusData{
		Since:      0,
		Activities: [] discord.Activity {
			{
				Name: locale.ACTIVITY,
				Type: discord.WatchingActivity,
			},
		},
		Status:     discord.OnlineStatus,
		AFK:        false,
	})

	logger.Logger.Infof("====== %s at your service ======", u.Username)

	promptClosed := startPromptLoop(s, sessionSelfCloser)

	select {
	case <-sCloser:
		close(sessionSelfCloser)
	case <-sessionSelfCloser:
	}
	s.ErrorLog = nil
	os.Stdin.WriteString("o")
	select {
	case <-promptClosed:
	case <-time.NewTimer(time.Second*5).C:
	}
	<-redirectorClosed
	err = sv.Close()
	if err != nil {
		return err
	}
	logger.Logger.Infof("[MAIN] Session is closed.")
	return nil
}

func redirectorLoop (s *state.State, loopCloser chan struct{}) (loopDone chan struct{}){
	loopDone = make(chan struct{})
	go func () {
		logger.Logger.Infof("[MAIN] Redirector is starting...")
		t := time.NewTimer(time.Minute)

		processRedirect := func (p *pendingEmbed) (err error){

			defer func () {
				if err == nil {
					if pErr := recover(); pErr != nil {
						err = pErr.(error)
					}
				}
			} ()
			logger.Logger.Infof("Redirector: %s-%d", p.taskId, p.embedIndex)
			
			var botMsg *discord.Message
			var originalMsg *discord.Message

			if p == nil {
				panic("nil pended?")
			}

			for time.Now().Before(p.estimatedRTime) {
				t.Reset(time.Second * 7)
				<-t.C
				
				botMsg, originalMsg, err = validate(s, p.cId, p.msgID)
				if botMsg == nil || originalMsg == nil {
					if err != nil {
						logger.Logger.Errorf("Failed to validate: %v", err)
					}
					memorizeResult(originalMsg.Embeds[p.embedIndex].URL, CancelledWithError)
					return
				}
			}
			

			// Do it again because passed < delay may not always be true
			botMsg, originalMsg, err = validate(s, p.cId, p.msgID)
			if botMsg == nil || originalMsg == nil {
				if err != nil {
					logger.Logger.Errorf("Failed to validate: %v", err)
				}
				memorizeResult(originalMsg.Embeds[p.embedIndex].URL, CancelledWithError)
				return
			}
	
			var communityVotes int
			var finalType redirect.RedirectType
			var result resultType = FALLBACK
			finalType, communityVotes, err = decideType(p, botMsg)
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to decide type: %v", err)
			}

			if p.autoType != finalType {
				noteGuessedWrong(badGuessRecord {
					title: originalMsg.Embeds[p.embedIndex].Title,
					guess: p.autoType,
					result: finalType,
				})
			}

			if finalType == redirect.None {
				err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				if err != nil {
					// Failed to remove the bot message...?
					logger.Logger.Errorf("Failed to remove the bot message: %v", err)
				}
				memorizeResult(originalMsg.Embeds[p.embedIndex].URL, Cancelled)
				return
			}
	
			destCId, bound := binding.QueryBinding(p.bindingId).DestChannelId(finalType)
			if !bound {
				logger.Logger.Infof("[MAIN] No destination bound to %v", finalType)
				err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				if err != nil {
					// Failed to remove the bot message...?
					logger.Logger.Errorf("Failed to remove the bot message: %d", err)
				}
				memorizeResult(originalMsg.Embeds[p.embedIndex].URL, CancelledWithError)
				return
			}

			if communityVotes == 0 { // No one voted
				if p.preType == redirect.Unknown { // It's NOT pre-typed
					if finalType == p.autoType {
						atomic.AddUint64(&statSession.GuessRight, 1)
						result = RESULT_BOT_GUESS
					} else {
						panic("WHO CHANGED MY ANSWER!?")
					}
				} else {
					if p.preType != finalType {
						panic("WHO CHANGED HIS ANSWER!!??")
					}

					if p.autoType == finalType { // Bot guessed it right
						atomic.AddUint64(&statSession.GuessRight, 1)
						result = RESULT_SHARER_AND_BOT
					} else {
						result = RESULT_SHARER
					}
				}
			} else { // Some one voted
				if p.preType == redirect.Unknown { // It's NOT pre-typed
					if p.autoType == redirect.Unknown {
						result = RESULT_COMMUNITY
					} else if finalType == p.autoType {
						atomic.AddUint64(&statSession.GuessRight, 1)
						result = RESULT_BOT_GUESS_AGREED
					} else {
						result = RESULT_BOT_GUESS_FIXED
					}
				} else {
					if p.autoType == finalType && p.preType == finalType {
						atomic.AddUint64(&statSession.GuessRight, 1)
						result = RESULT_SHARER_AND_BOT_AND_COMMUNITY
					} else if p.autoType == finalType {
						atomic.AddUint64(&statSession.GuessRight, 1)
						result = RESULT_SHARER_AND_BOT
					} else if p.preType == finalType {
						result = RESULT_SHARER_AND_COMMUNITY
					} else {
						result = RESULT_BOT_GUESS_FIXED
					}
				}
			}

			
			var data *api.SendMessageData
			data, err = prepareRedirectionMessage(originalMsg, p, result)
			if err != nil {
				logger.Logger.Errorf("Failed to prepare redirection message!\n%v", err)
				memorizeResult(originalMsg.Embeds[p.embedIndex].URL, CancelledWithError)
				return
			}
	
			var rm *discord.Message
			rm, err = s.SendMessageComplex(
				discord.ChannelID(destCId), *data,
			)
	
			if err != nil {
				logger.Logger.Errorf("F1: %s", err)
			}
	
			_, err = s.SendMessage(discord.ChannelID(destCId), fmt.Sprintf("%s %s", originalMsg.Embeds[p.embedIndex].URL, locale.EXPLAIN_EMBED_RESOLVE))
			if err != nil {
				logger.Logger.Errorf("F2: %s", err)
			}

			if communityVotes > 3 {	
				_, err = s.SendMessage(discord.ChannelID(destCId), locale.HOT)
				if err != nil {
					logger.Logger.Errorf("F3: %s", err)
				}
			}

			memorizeResult(originalMsg.Embeds[p.embedIndex].URL, Redirected)

			logger.Logger.Infof("  Redirected     c%d - m%d", destCId, rm.ID)	
			atomic.AddUint64(&statSession.Redirected, 1)

			err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to remove the bot message: %d", err)
			}
	
			return nil
		}
		rLoop: for {
			select {
			case p := <-pendingEmbeds: // block until new pending
				if rErr := processRedirect(p); rErr != nil {
					logger.Logger.Infof("Redirector task failed with: %v", rErr)
				}
				break
			case <-loopCloser:
				endingLoop: for {
					select {
					case p := <-pendingEmbeds: // block until new pending
						if rErr := processRedirect(p); rErr != nil {
							logger.Logger.Infof("Redirector task failed with: %v", rErr)
						}
						break
					default:
						break endingLoop
					}
				}
				break rLoop
			}

		}
		close(loopDone)
		logger.Logger.Infof("[MAIN] Redirector is ended.")
	} ()
	return loopDone
}

func prepareRedirectionMessage (originalMsg *discord.Message, nextPending *pendingEmbed, result resultType) (data *api.SendMessageData, err error) {
	defer func () {
		if err == nil {
			if pErr := recover(); pErr != nil {
				err = pErr.(error)
				logger.Logger.Errorf("Panic! When preparing message: %v", err)
			}
		}
	} ()
	return  &api.SendMessageData{
		Content:    "ｷﾀｷﾀ────=====≡≡Σ≡Σ((つ*°∀°)و \\*✧\\*✧\\*✧ ",
		Embeds:     []discord.Embed{
			{
				Type: discord.NormalEmbed,
				Title: originalMsg.Embeds[nextPending.embedIndex].Title,
				// Description: originalMsg.Embeds[nextPending.embedIndex].Description,
				URL: originalMsg.Embeds[nextPending.embedIndex].URL,
				// Timestamp: originalMsg.Embeds[nextPending.embedIndex].Timestamp,
				// Footer: originalMsg.Embeds[nextPending.embedIndex].Footer,
				Color: originalMsg.Embeds[nextPending.embedIndex].Color,
				Provider: &discord.EmbedProvider{
					Name: originalMsg.Embeds[nextPending.embedIndex].Provider.Name,
					URL: originalMsg.Embeds[nextPending.embedIndex].Provider.URL,
				},
				Author: originalMsg.Embeds[nextPending.embedIndex].Author,
				Fields: []discord.EmbedField {
					{
						Name: locale.SHARER,
						Value: originalMsg.Author.Username + "#" + originalMsg.Author.Discriminator,
						Inline: true,
					},
					{
						Name: locale.SMSG,
						Value: originalMsg.URL(),
						Inline: true,
					},
					{
						Name: locale.DECISION_TYPE,
						Inline: true,
						Value: func () string{
							switch result{
							case RESULT_BOT_GUESS:
								return locale.DECISION_BOT
							case RESULT_BOT_GUESS_AGREED:
								return locale.DECISION_COMMUNITY_AGREE
							case RESULT_BOT_GUESS_FIXED:
								return locale.DECISION_COMMUNITY_FIX
							case RESULT_COMMUNITY:
								return locale.DECISION_COMMUNITY_HELP
							case RESULT_SHARER:
								return locale.DECISION_SHARER
							case RESULT_SHARER_AND_BOT:
								return locale.DECISION_SHARER_AND_BOT
							case RESULT_SHARER_AND_BOT_AND_COMMUNITY:
								return locale.DECISION_SHARER_AND_BOT_AND_COMMUNITY
							case RESULT_SHARER_AND_COMMUNITY:
								return locale.DECISION_SHARER_AND_COMMUNITY
							default:
								return "?"
							}
						} (),
					},
				},
			},
		},
	}, nil
}

// decideType only gives Original, Cover, Stream, None
func decideType (pending *pendingEmbed, botMsg *discord.Message) (rType redirect.RedirectType, communityVotes int, err error) {
	
	if pending.urlValidation != botMsg.ReferencedMessage.Embeds[pending.embedIndex].URL {
		logger.Logger.Infof("  [!] Url modified, ABORT!")
		return redirect.None, 0, nil

	}

	var c_o, c_c, c_s, c_n int

	for _, r := range botMsg.Reactions {
		switch r.Emoji.Name {
		case reactionCover:
			c_c = r.Count
			break
		case reactionOriginal:
			c_o = r.Count
			break
		case reactionStream:
			c_s = r.Count
			break
		case reactionNone:
			c_n = r.Count
			break
		}
	}

	communityVotes = c_o + c_c + c_s + c_n
	if communityVotes == 0  {
		if pending.isDup { // Skip duplicates
			logger.Logger.Infof("  [%s-%d] DUPLICATE", pending.taskId, pending.embedIndex)
			atomic.AddUint64(&statSession.SkippedDuplicate, 1)
			return redirect.None, 0, nil
		}
		// If no user vote
		if pending.preType == redirect.Unknown { //  and no pre_type, apply guess
			switch pending.autoType {
				case redirect.Original:
					c_o++
					break
				case redirect.Cover:
					c_c++
					break
				case redirect.Stream:
					c_s++
					break
			}
		} else { // accept the pre_type
			switch pending.preType {
			case redirect.Original:
				c_o++
				break
			case redirect.Cover:
				c_c++
				break
			case redirect.Stream:
				c_s++
				break
			}
		}
	} else {
		if pending.autoType == pending.preType {
			switch pending.autoType {
				case redirect.Original:
					c_o++
					break
				case redirect.Cover:
					c_c++
					break
				case redirect.Stream:
					c_s++
					break
			}
		}
	}

	sum := c_o + c_c + c_s
	if sum == 0 || (c_n > c_o - 1 && c_n > c_c - 1 && c_n > c_s - 1) {
		if pending.autoType == redirect.None {
			logger.Logger.Infof("  [%s-%d] CANCEL   | o%d c%d s%d n%d ✔️", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		} else {
			logger.Logger.Infof("  [%s-%d] CANCEL   | o%d c%d s%d n%d ❓", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		}
		return redirect.None, communityVotes, nil
	}

	if c_o > c_c && c_o > c_s {
		if pending.autoType == redirect.Original {
			logger.Logger.Infof("  [%s-%d] ORIGINAL | o%d c%d s%d n%d ✔️", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		} else {
			logger.Logger.Infof("  [%s-%d] ORIGINAL | o%d c%d s%d n%d ❓", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		}
		rType = redirect.Original
	}

	if c_s > c_c && c_s > c_o {
		if pending.autoType == redirect.Stream {
			logger.Logger.Infof("  [%s-%d] STREAM   | o%d c%d s%d n%d ✔️", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		} else {
			logger.Logger.Infof("  [%s-%d] STREAM   | o%d c%d s%d n%d ❓", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		}
		rType = redirect.Stream
	}

	if c_c > c_s && c_c > c_o {
		if pending.autoType == redirect.Cover {
			logger.Logger.Infof("  [%s-%d] COVER    | o%d c%d s%d n%d ✔️", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		} else {
			logger.Logger.Infof("  [%s-%d] COVER    | o%d c%d s%d n%d ❓", pending.taskId, pending.embedIndex, c_o, c_c, c_s, c_n)
		}
		rType = redirect.Cover
	}

	return rType, communityVotes, nil
}

func validate (s *state.State, botMsgCId discord.ChannelID, botMsgID discord.MessageID) (botMsg *discord.Message, originalMsg *discord.Message, err error) {
	botMsg, err = s.Message(botMsgCId, botMsgID)
	if err != nil || botMsg == nil {
		// Failed to access the bot message, deleted?
		logger.Logger.Errorf("Bot message inaccessible: %d", botMsgCId)
		return botMsg, originalMsg, err
	} else {
		if originalMsg, err = s.Message(botMsg.Reference.ChannelID, botMsg.Reference.MessageID); originalMsg == nil || err != nil {
			logger.Logger.Errorf("Original message inaccessible (error? %v", err)
			err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to remove the bot message: %d", err)
			} else {
				botMsg = nil
			}
			return botMsg, originalMsg, err
		}
	}
	return botMsg, originalMsg, nil
}