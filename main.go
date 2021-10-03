package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"No3371.github.com/song_librarian.bot/binding"
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/redirect"
	"No3371.github.com/song_librarian.bot/storage"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)


var pendingEmbeds chan *pendingEmbed
var sv storage.StorageProvider
var delay time.Duration

type pendingEmbed struct {
	cId discord.ChannelID
	msgID discord.MessageID
	embedIndex int
	bindingId int
	pendedTime time.Time
}


func main() {
	resolveFlags()
	locale.SetLanguage(locale.FromString(*globalFlags.locale))
	delay = time.Minute * time.Duration(*globalFlags.delay)
	var err error

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

	go func () {
		var err error
		t := time.NewTimer(time.Minute)
		var nextPending *pendingEmbed
		for {
			if nextPending == nil {
				nextPending = <-pendingEmbeds // block until new pending
			}
			
			passed := time.Now().Sub(nextPending.pendedTime)
			if (passed < delay) {
				logger.Logger.Infof("Proceed in %s...", delay - passed)
				t.Reset(delay - passed)
				<-t.C
				continue
			}

			var botMsg *discord.Message
			botMsg, err = s.Message(nextPending.cId, nextPending.msgID)
			if err != nil {
				// Failed to access the bot message, deleted?
				logger.Logger.Errorf("Bot message inaccessible: %d", nextPending.msgID)
				nextPending = nil
				continue // cancel
			}

			// Check if the original message is not deleted
			var originalMsg *discord.Message
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

			log.Printf("[LOG] O: %d, C: %d, S: %d, N: %d", c_o, c_c, c_s, c_n)
			sum := c_o + c_c + c_s
			if sum == 0 || (c_n > c_o - 1 && c_n > c_c - 1 && c_n > c_s - 1) {
				logger.Logger.Infof("[REDIRECT] Result: Cancel.")
				nextPending = nil
				err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				if err != nil {
					// Failed to remove the bot message...?
					logger.Logger.Errorf("Failed to remove the bot message: %d", err)
				}
				continue // cancel
			}

			var rType redirect.RedirectType = redirect.None

			if c_o > c_c && c_o > c_s {
				rType = redirect.Original
			}

			if c_s > c_c && c_s > c_o {
				rType = redirect.Stream
			}

			if c_c > c_s && c_c > c_o {
				rType = redirect.Cover
			}
			
			destCId, bound := binding.QueryBinding(nextPending.bindingId).DestChannelId(rType)
			if !bound {
				nextPending = nil
				err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
				if err != nil {
					// Failed to remove the bot message...?
					logger.Logger.Errorf("Failed to remove the bot message: %d", err)
				}
				continue // cancel
			}
			logger.Logger.Infof("[REDIRECT] Redirecting to channel %d", destCId)

			_, err = s.SendMessage(
				discord.ChannelID(destCId),
			 	fmt.Sprintf(locale.REDIRECT_FORMAT,	originalMsg.Author.Username, originalMsg.URL()),
				originalMsg.Embeds[nextPending.embedIndex],
			)

			if err != nil {
				logger.Logger.Errorf("Failed to redirect the embed: %s", err)
			}

			err = s.DeleteMessage(botMsg.ChannelID, botMsg.ID, "Temporary bot message")
			if err != nil {
				// Failed to remove the bot message...?
				logger.Logger.Errorf("Failed to remove the bot message: %d", err)
			}

			nextPending = nil
		}
	} ()

	err = assureCommands(s)
	if err != nil {
		logger.Logger.Fatalf("[MAIN] %v", err)
	}

	addHandlers(s)

	if err := s.Open(context.Background()); err != nil {
		logger.Logger.Errorf("[MAIN] %v", err)
	}
	defer s.Close()

	u, err := s.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}

	log.Println("Started as", u.Username)
	var promptContext context.Context
	var cancelPromptFunc context.CancelFunc 
	promptContext, cancelPromptFunc = context.WithCancel(context.Background())
	go promptLoop(s, promptContext)

	// Block forever.
	select {}
	cancelPromptFunc()
}

