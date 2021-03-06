package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
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
	"No3371.github.com/song_librarian.bot/memory"
	"No3371.github.com/song_librarian.bot/redirect"
	"No3371.github.com/song_librarian.bot/storage"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/pkg/errors"
)


var pendingEmbeds chan *pendingEmbed
var sp storage.StorageProvider

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
	spoiler bool
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
		runtime.SetCPUProfileRate(200)
        f, err := os.Create("./cpuprof")
        if err != nil {
            logger.Logger.Fatalf("could not create CPU profile: ", err)
        }
        defer f.Close() // error handling omitted for example
		logger.Logger.Infof("Starting cpu profiling...")
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
	os.Exit(0)
}

func session (sCloser chan struct{}) (err error) {
	var sessionSelfCloser chan struct{} = make(chan struct{})
	logger.Logger.Infof("[MAIN] Session is starting...")
	sp, err = storage.Sqlite(*globalFlags.printSqlStmt)
	if err != nil {
		return errors.Wrap(err, "Failed to get storage")
	}

	binding.Setup(sp)
	memory.Setup(sp, *globalFlags.memsize)

	s, err := state.New("Bot " + *globalFlags.token)
	if err != nil {
		return errors.Wrap(err, "Failed to get new bot state")
	}

	s.AddIntents(gateway.IntentDirectMessages)
	s.AddIntents(gateway.IntentGuildMessages)
	s.AddIntents(gateway.IntentGuildMessageReactions)

	pendingEmbeds = make(chan *pendingEmbed, 512)

	err = assureCommands(s)
	if err != nil {
		logger.Logger.Fatalf("[MAIN] %v", err)
	}

	addEventHandlers(s)
	addInteractionHandlers(s)

	redirectorClosed := redirectorLoop(s, sessionSelfCloser)

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
	logger.Logger.Infof("Session: %d", u.ID)
	
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
	logger.Logger.Infof("SAVE ALL running")
	binding.SaveAll()
	logger.Logger.Infof("SAVE ALL finished")
	err = sp.Close()
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

			var botMsg *discord.Message
			var originalMsg *discord.Message

			defer func () {
				if err == nil {
					if pErr := recover(); pErr != nil {
						logger.Logger.Errorf("TARCING PANIC:\n", debug.Stack())
						err = fmt.Errorf("PANIC: %v", pErr)
					}
				}

				if botMsg != nil {
					err2 := s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
					if err2 != nil {
						// Failed to remove the bot message...?
						logger.Logger.Errorf("Failed to remove the bot message: %v", err2)
					} else {
						botMsg = nil
					}
				}
			} ()
			logger.Logger.Infof("Redirector [%s-%d]", p.taskId, p.embedIndex)

			var _binding = binding.QueryBinding(p.bindingId)
			if _binding == nil {
				panic("nil binding")
			}
			if p == nil {
				panic("nil pended?")
			}

			for time.Now().Before(p.estimatedRTime.Add(time.Second * 2)) {
				t.Reset(time.Second * 16)
				<-t.C
				if *globalFlags.novote {
					break
				}
				botMsg, originalMsg, err = validate(s, p)
				if botMsg == nil || originalMsg == nil {
					break // Do not cancel here, we still have one more chance down there (Somehow sometimes Discord randomly return 404 for message?)
				}
			}

			// Do it again because passed < delay may not always be true
			botMsg, originalMsg, err = validate(s, p)
			if botMsg == nil || originalMsg == nil {
				if originalMsg != nil {
					err = _binding.Memorize(originalMsg.Embeds[p.embedIndex].URL, memory.CancelledWithError)
					if err != nil {
						logger.Logger.Errorf("[%s-%d] Failed to memorize.", p.taskId, p.embedIndex)
					}
				}
				if !*globalFlags.novote {
					return
				}
			}

			if *globalFlags.novote && originalMsg == nil {
				return
			}
	
			lastRedirectTime := _binding.GetLastTime(originalMsg.Embeds[p.embedIndex].URL, memory.Redirected)
			if time.Since(lastRedirectTime) < time.Minute { // Handles user racing
				logger.Logger.Infof("  [%s-%d] Seems like it just got redirected, abort.", p.taskId, p.embedIndex)
				// err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				// if err != nil {
				// 	// Failed to remove the bot message...?
				// 	logger.Logger.Errorf("Failed to remove the bot message: %d", err)
				// }
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

			destCId, bound := binding.QueryBinding(p.bindingId).DestChannelId(finalType)
			if !bound {
				logger.Logger.Infof("[MAIN] No destination bound to %v", finalType)
				// err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				// if err != nil {
				// 	// Failed to remove the bot message...?
				// 	logger.Logger.Errorf("Failed to remove the bot message: %d", err)
				// }
				if originalMsg != nil {
					err = _binding.Memorize(originalMsg.Embeds[p.embedIndex].URL, memory.CancelledWithError)
					if err != nil {
						logger.Logger.Errorf("[%s-%d] Failed to memorize.", p.taskId, p.embedIndex)
					}
				}
				return
			}

			if finalType == redirect.None {
				if p.autoType == redirect.None {
					atomic.AddUint64(&statSession.GuessRight, 1)
				}
				// err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				// if err != nil {
				// 	// Failed to remove the bot message...?
				// 	logger.Logger.Errorf("Failed to remove the bot message: %v", err)
				// }
				if *globalFlags.redirNone {
					var rm *discord.Message
					rm, err = send(s, destCId, p, originalMsg, _binding, result)
					if err != nil {
						return
					}
		
					err = _binding.Memorize(originalMsg.Embeds[p.embedIndex].URL, memory.Redirected)
					if err != nil {
						logger.Logger.Errorf("  [%s-%d] Failed to memorize.", p.taskId, p.embedIndex)
					}
		
					if rm != nil {
						logger.Logger.Infof("  [%s-%d] Redirected     c%d - m%d", p.taskId, p.embedIndex, destCId, rm.ID)	
					}
					atomic.AddUint64(&statSession.Redirected, 1)
					return
				}
				err = _binding.Memorize(originalMsg.Embeds[p.embedIndex].URL, memory.Cancelled)
				if err != nil {
					logger.Logger.Errorf("[%s-%d] Failed to memorize.", p.taskId, p.embedIndex)
				}
				return
			}

			if p.autoType != finalType { // finalType is not None
				atomic.AddUint64(&statSession.GueseWrongType, 1)
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
	
			var rm *discord.Message
			rm, err = send(s, destCId, p, originalMsg, _binding, result)
			if err != nil {
				return
			}

			err = _binding.Memorize(originalMsg.Embeds[p.embedIndex].URL, memory.Redirected)
			if err != nil {
				logger.Logger.Errorf("  [%s-%d] Failed to memorize.", p.taskId, p.embedIndex)
			}

			if rm != nil {
				logger.Logger.Infof("  [%s-%d] Redirected     c%d - m%d", p.taskId, p.embedIndex, destCId, rm.ID)	
			}
			atomic.AddUint64(&statSession.Redirected, 1)
	
			return nil
		}

		rLoop: for {
			select {
			case p := <-pendingEmbeds: // block until new pending
				if rErr := processRedirect(p); rErr != nil {
					logger.Logger.Errorf("Redirector task failed with: %v", rErr)
				}
			case <-loopCloser:
				endingLoop: for {
					select {
					case p := <-pendingEmbeds: // block until new pending
						if rErr := processRedirect(p); rErr != nil {
							logger.Logger.Errorf("Redirector task failed with: %v", rErr)
						}
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
		Content:    "????????????????????????=====?????????????((???*???????)?? \\*???\\*???\\*??? ",
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

	if *globalFlags.novote {
		if pending.isDup { // Skip duplicates
			logger.Logger.Infof("  [%s-%d] DUPLICATE", pending.taskId, pending.embedIndex)
			atomic.AddUint64(&statSession.SkippedDuplicate, 1)
			return redirect.None, 0, nil
		}

		return pending.autoType, 0, nil
	}

	if pending.urlValidation != botMsg.ReferencedMessage.Embeds[pending.embedIndex].URL {
		logger.Logger.Infof("  [!] Url modified, ABORT!")
		return redirect.None, 0, nil

	}

	var cO, cC, cS, cX int

	for _, r := range botMsg.Reactions {
		switch r.Emoji.Name {
		case reactionCover:
			cC = r.Count
			break
		case reactionOriginal:
			cO = r.Count
			break
		case reactionStream:
			cS = r.Count
			break
		case reactionNone:
			cX = r.Count
			break
		}
	}

	communityVotes = cO + cC + cS + cX

	// If no user vote
	if communityVotes == 0  {
		if pending.isDup { // Skip duplicates
			logger.Logger.Infof("  [%s-%d] DUPLICATE", pending.taskId, pending.embedIndex)
			atomic.AddUint64(&statSession.SkippedDuplicate, 1)
			return redirect.None, 0, nil
		}

		if pending.preType == redirect.Unknown { //  and no pre_type, apply guess
			switch pending.autoType {
				case redirect.Original:
					cO++
				case redirect.Cover:
					cC++
				case redirect.Stream:
					cS++
			}
		} else { // accept the pre_type
			switch pending.preType {
			case redirect.Original:
				cO++
			case redirect.Cover:
				cC++
			case redirect.Stream:
				cS++
			}
		}
	} else { // users voted
		if pending.autoType == pending.preType {
			switch pending.autoType {
				case redirect.Original:
					cO++
				case redirect.Cover:
					cC++
				case redirect.Stream:
					cS++
			}
		}
	}

	rType = decideByFinalVote(cO, cC, cS, cX, pending)

	return rType, communityVotes, nil
}

func decideByFinalVote (o, c, s, x int, pending *pendingEmbed) (rType redirect.RedirectType) {
	sum := o + c + s

	if sum == 0 || (x > o - 1 && x > c - 1 && x > s - 1) { // X has higher power
		if pending.autoType == redirect.None {
			logger.Logger.Infof("  [%s-%d] CANCEL   | o%d c%d s%d x%d ??????", pending.taskId, pending.embedIndex, o, c, s, x)
		} else {
			logger.Logger.Infof("  [%s-%d] CANCEL   | o%d c%d s%d x%d ???", pending.taskId, pending.embedIndex, o, c, s, x)
		}
		return redirect.None
	}

	if o > c && o > s {
		if pending.autoType == redirect.Original {
			logger.Logger.Infof("  [%s-%d] ORIGINAL | o%d c%d s%d x%d ??????", pending.taskId, pending.embedIndex, o, c, s, x)
		} else {
			logger.Logger.Infof("  [%s-%d] ORIGINAL | o%d c%d s%d x%d ???", pending.taskId, pending.embedIndex, o, c, s, x)
		}
		rType = redirect.Original
	} else if s > c && s > o {
		if pending.autoType == redirect.Stream {
			logger.Logger.Infof("  [%s-%d] STREAM   | o%d c%d s%d x%d ??????", pending.taskId, pending.embedIndex, o, c, s, x)
		} else {
			logger.Logger.Infof("  [%s-%d] STREAM   | o%d c%d s%d x%d ???", pending.taskId, pending.embedIndex, o, c, s, x)
		}
		rType = redirect.Stream
	} else if c > s && c > o {
		if pending.autoType == redirect.Cover {
			logger.Logger.Infof("  [%s-%d] COVER    | o%d c%d s%d x%d ??????", pending.taskId, pending.embedIndex, o, c, s, x)
		} else {
			logger.Logger.Infof("  [%s-%d] COVER    | o%d c%d s%d x%d ???", pending.taskId, pending.embedIndex, o, c, s, x)
		}
		rType = redirect.Cover
	}

	return
}

func validate (s *state.State, task *pendingEmbed) (botMsg *discord.Message, originalMsg *discord.Message, err error) {
	if *globalFlags.novote {
		if originalMsg, err = s.Message(task.cId, task.msgID); originalMsg == nil || err != nil {
			logger.Logger.Infof("[%s-%d] ! Original message inaccessible (error? %v", task.taskId, task.embedIndex, err)
		}
		return nil, originalMsg, err
	}

	botMsg, err = s.Message(task.cId, task.msgID)
	if err != nil || botMsg == nil {
		// Failed to access the bot message, deleted?
		logger.Logger.Errorf("Bot message inaccessible: %d", task.cId)
		return botMsg, originalMsg, err
	} else {
		if originalMsg, err = s.Message(botMsg.Reference.ChannelID, botMsg.Reference.MessageID); originalMsg == nil || err != nil {
			logger.Logger.Errorf("[%s-%d] Original message inaccessible (error? %v", task.taskId, task.embedIndex, err)
			err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to remove the bot message: %s", err)
			} else {
				botMsg = nil
			}
			return botMsg, originalMsg, err
		}
	}
	
	return botMsg, originalMsg, nil
}

func send (s *state.State, destCId uint64, p *pendingEmbed, originalMsg *discord.Message, binding *binding.ChannelBinding, result resultType) (msg *discord.Message, err error) {
	
	var data *api.SendMessageData
	data, err = prepareRedirectionMessage(originalMsg, p, result)
	if err != nil {
		logger.Logger.Errorf("Failed to prepare redirection message!\n%v", err)
		if originalMsg != nil {
			err = binding.Memorize(originalMsg.Embeds[p.embedIndex].URL, memory.CancelledWithError)
			if err != nil {
				logger.Logger.Errorf("[%s-%d] Failed to memorize.", p.taskId, p.embedIndex)
			}
		}
		return
	}

	msg, err = s.SendMessageComplex(
		discord.ChannelID(destCId), *data,
	)

	if err != nil {
		logger.Logger.Errorf("Failed to send message1: %s", err)
		msg, err = s.SendMessageComplex(
			discord.ChannelID(destCId), *data,
		)
	}

	_, err = s.SendMessage(discord.ChannelID(destCId), fmt.Sprintf("%s %s", originalMsg.Embeds[p.embedIndex].URL, locale.EXPLAIN_EMBED_RESOLVE))
	if err != nil {
		logger.Logger.Errorf("Failed to send message2: %s", err)
	}

	return
}