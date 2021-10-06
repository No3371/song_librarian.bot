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
)


var pendingEmbeds chan *pendingEmbed
var sv storage.StorageProvider

type pendingEmbed struct {
	cId discord.ChannelID
	msgID discord.MessageID
	embedIndex int
	bindingId int
	pendedTime time.Time
	guess redirect.RedirectType
}


var processCloser chan struct{}
func main() {
	var err error
	resolveFlags()
	locale.SetLanguage(locale.FromString(*globalFlags.locale))
	statSession = &stats{}

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
	for restart {
		if err != nil {
			logger.Logger.Errorf("[MAIN] Previous session closed with error: %v", err)
			logger.Logger.Infof("[MAIN] Restarting...")
		}
		err = nil
		err = session(sessionCloser)
	}
	statSession.Print()
	logger.Logger.Infof("SAVE ALL running")
	binding.SaveAll()
	logger.Logger.Infof("SAVE ALL finished")
	os.Exit(0)
}

func session (sCloser chan struct{}) (err error) {
	var selfCloser chan struct{} = make(chan struct{})
	logger.Logger.Infof("[MAIN] Session is starting...")
	sv, err = storage.Sqlite()
	if err != nil {
		log.Fatalf("Failed to get storage: %s", err)
	}

	s, err := state.New("Bot " + *globalFlags.token)
	if err != nil {
		log.Fatalln("Session failed:", err)
	}

	s.AddIntents(gateway.IntentGuildMessages)
	s.AddIntents(gateway.IntentGuildMessageReactions)

	pendingEmbeds = make(chan *pendingEmbed, 1024)

	redirectorClosed := redirectorLoop(s, sCloser)

	// err = assureCommands(s)
	// if err != nil {
	// 	logger.Logger.Fatalf("[MAIN] %v", err)
	// }

	addEventHandlers(s)
	// addInteractionHandlers(s)

	s.FatalErrorCallback = func(innerErr error) {
		logger.Logger.Errorf("[MAIN] Fatal gateway error: %s", err)
	}

	s.AfterClose = func(innerErr error) {
		logger.Logger.Errorf("[MAIN] After gateway closed: %s", err)
		err = innerErr
		close(selfCloser)
	}

	if err := s.Open(context.Background()); err != nil {
		logger.Logger.Fatalf("[MAIN] %v", err)
	}
	defer s.Close()

	u, err := s.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}

	logger.Logger.Infof("====== %s at your service ======", u.Username)

	promptClosed := startPromptLoop(s, sCloser)

	select {
	case <-sCloser:
	case <-selfCloser:
	}
	<-promptClosed
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
		var err error
		t := time.NewTimer(time.Minute)
		var nextPending *pendingEmbed
		loopBody: for {
			if nextPending == nil {
				select {
					case nextPending = <-pendingEmbeds: // block until new pending
				case <-loopCloser:
					break loopBody					
				}
			}
			logger.Logger.Infof("Redirector: new task.")
			
			
			var botMsg *discord.Message
			var originalMsg *discord.Message

			passed := time.Now().Sub(nextPending.pendedTime)
			delay := *globalFlags.delay
			if *globalFlags.dev {
				delay = time.Second * 5
			}

			for passed < delay && nextPending != nil {
				t.Reset(time.Second * 10)
				select {
					case <-t.C:
				case <-loopCloser:
					break loopBody
				}
				passed = time.Now().Sub(nextPending.pendedTime)
				
				botMsg, err = s.Message(nextPending.cId, nextPending.msgID)
				if err != nil || botMsg == nil {
					// Failed to access the bot message, deleted?
					logger.Logger.Errorf("Bot message inaccessible: %d", nextPending.msgID)
					nextPending = nil
					continue // cancel
				} else {
					if originalMsg, err = s.Message(botMsg.Reference.ChannelID, botMsg.Reference.MessageID); originalMsg == nil || err != nil {
						logger.Logger.Errorf("Original message inaccessible: %d (error? %s)", botMsg.Reference.MessageID, err)
						nextPending = nil
						err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
						if err != nil {
							// Failed to remove the bot message...?
							logger.Logger.Errorf("Failed to remove the bot message: %d", err)
						}
						continue // error skip
					}
				}
			}

			if nextPending == nil {
				nextPending = nil
				continue // cancel
			}
	
			botMsg, err = s.Message(nextPending.cId, nextPending.msgID)
			if err != nil || botMsg == nil {
				// Failed to access the bot message, deleted?
				logger.Logger.Errorf("Bot message inaccessible: %d", nextPending.msgID)
				nextPending = nil
				continue // cancel
			} else {
				if originalMsg, err = s.Message(botMsg.Reference.ChannelID, botMsg.Reference.MessageID); originalMsg == nil || err != nil {
					logger.Logger.Errorf("Original message inaccessible: %d (error? %s)", botMsg.Reference.MessageID, err)
					nextPending = nil
					err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
					if err != nil {
						// Failed to remove the bot message...?
						logger.Logger.Errorf("Failed to remove the bot message: %d", err)
					}
					continue // error skip
				}
			}
	
			// Check if the original message is not deleted
			var isAuto bool
			var rType redirect.RedirectType
			rType, isAuto, err = decideType(nextPending, botMsg)
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to decide type: %v", err)
			}
			
			if rType == redirect.None {
				nextPending = nil
				err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				if err != nil {
					// Failed to remove the bot message...?
					logger.Logger.Errorf("Failed to remove the bot message: %v", err)
				}
				continue
			}
	
			destCId, bound := binding.QueryBinding(nextPending.bindingId).DestChannelId(rType)
			if !bound {
				logger.Logger.Infof("[MAIN] No destination bound to %v", rType)
				nextPending = nil
				err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				if err != nil {
					// Failed to remove the bot message...?
					logger.Logger.Errorf("Failed to remove the bot message: %d", err)
				}
				continue // cancel
			}

			logger.Logger.Infof("[REDIRECT] Redirecting to channel %d", destCId)
	
			data := api.SendMessageData{
				Content:    "ｷﾀ━━ｷﾀ━━ｷﾀ──────────==========≡≡≡≡≡Σ≡Σ(((つ•̀ㅂ•́)و \\*✧\\*✧\\*✧ ",
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
								Value: originalMsg.Author.Username,
								Inline: true,
							},
							{
								Name: locale.SMSG,
								Value: originalMsg.URL(),
								Inline: true,
							},
							{
								Name: locale.SDTYPE,
								Value: func () string{
									if isAuto {
										return locale.SDTYPE_AUTO
									} else {
										return locale.SDTYPE_MANUAL
									}
								} (),
								Inline: true,
							},
						},
					},
				},
			}
	
			_, err := s.SendMessageComplex(
				discord.ChannelID(destCId), data,
			)
	
			if err != nil {
				logger.Logger.Errorf("F1: %s", err)
			}
	
			_, err = s.SendMessage(discord.ChannelID(destCId), fmt.Sprintf("%s %s", originalMsg.Embeds[nextPending.embedIndex].URL, locale.EXPLAIN_EMBED_RESOLVE))
	
			if err != nil {
				logger.Logger.Errorf("F2: %s", err)
			}
	
			atomic.AddUint64(&statSession.redirected, 1)

			err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to remove the bot message: %d", err)
			}
	
			nextPending = nil
		}
		close(loopDone)
		logger.Logger.Infof("[MAIN] Redirector is ended.")
	} ()
	return loopDone
}

func decideType (pending *pendingEmbed,botMsg *discord.Message) (rType redirect.RedirectType, auto bool, err error) {
	
	var c_o, c_c, c_s, c_n int
	var isAuto bool = false

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

	if c_o + c_c + c_s + c_n == 0 { // If no user vote, apply guess
		isAuto = true
		switch pending.guess {
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


	logger.Logger.Infof("[LOG] O: %d, C: %d, S: %d, N: %d", c_o, c_c, c_s, c_n)

	sum := c_o + c_c + c_s
	if sum == 0 || (c_n > c_o - 1 && c_n > c_c - 1 && c_n > c_s - 1) {
		logger.Logger.Infof("[REDIRECT] Result: Cancel.")
		return redirect.None, isAuto, nil
	}

	if c_o > c_c && c_o > c_s {
		rType = redirect.Original
	}

	if c_s > c_c && c_s > c_o {
		rType = redirect.Stream
	}

	if c_c > c_s && c_c > c_o {
		rType = redirect.Cover
	}

	return rType, isAuto, nil
}