package main

import (
	// "No3371.github.com/song_librarian.bot/logger"
	// "github.com/diamondburned/arikawa/v3/discord"
	// "github.com/diamondburned/arikawa/v3/gateway"
	"strconv"

	"No3371.github.com/song_librarian.bot/logger"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func assureInteractions (s *state.State) {
	// s.InteractionResponse(appID discord.AppID, token string)
}

func addInteractionHandlers(s *state.State) {
	s.AddHandler(func (e *gateway.InteractionCreateEvent) {
		var err error
		var me *discord.User
		me, err = s.Me()
		if err != nil {
			logger.Logger.Errorf("*state.State.Me err: %v", err)
			return
		}
		switch e.Type {
		case discord.CommandInteraction:
			logger.Logger.Infof("CommandInteraction")
			var data = e.Data.(*discord.CommandInteractionData)
			switch commandIdMap[data.ID] {
			case DeleteRedirectedMessage:
				if len(data.Options) < 2 {
					logger.Logger.Infof("Need 2 option values")
					return
				}
				logger.Logger.Infof("  DELETE (%s): %s - %s", data.Name, data.Options[0], data.Options[1])
				var cId uint64
				var mId uint64
				mId, err = strconv.ParseUint(data.Options[1].String(), 10, 64)
				cId, err = strconv.ParseUint(data.Options[0].String(), 10, 64)
				if err != nil {
					logger.Logger.Infof("Failed to parse message Id: %s", data.Options[0])
					return
				}
				var m *discord.Message
				if m, err = s.Message(discord.ChannelID(cId), discord.MessageID(mId)); err == nil && m != nil && m.Author.ID == me.ID {
					logger.Logger.Infof("  Message fetched and it's mine")
					err = s.DeleteMessage(e.Message.ChannelID, e.Message.ID, "Requested")
					if err != nil {
						logger.Logger.Errorf("  DeleteMessage err: %v", err)
						return
					}
				}
				break
			}
			break
		}
	})
	
	// s.AddHandler(func (e *gateway.InteractionCreateEvent) {
	// 	var err error
	// 	var me *discord.User
	// 	me, err = s.Me()
	// 	if err != nil {
	// 		logger.Logger.Errorf("*state.State.Me err: %v", err)
	// 		return
	// 	}
	// 	switch e.Type {
	// 	case discord.CommandInteraction:
	// 		logger.Logger.Infof("CommandInteraction")
	// 		var data = e.Data.(*discord.CommandInteractionData)
	// 		switch commandIdMap[data.ID] {
	// 		case DeleteRedirectedMessage:
	// 			logger.Logger.Infof("  DELETE (%s)", data.Name)
	// 			if e. == me.ID {
	// 				err = s.DeleteMessage(e.Message.ChannelID, e.Message.ID, "Requested")
	// 				if err != nil {
	// 					logger.Logger.Errorf("  DeleteMessage err: %v", err)
	// 					return
	// 				}
	// 			}
	// 			break
	// 		}
	// 		break
	// 	}
	// })
}