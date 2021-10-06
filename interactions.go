package main

import (
	// "No3371.github.com/song_librarian.bot/logger"
	// "github.com/diamondburned/arikawa/v3/discord"
	// "github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func assureInteractions (s *state.State) {
	// s.InteractionResponse(appID discord.AppID, token string)
}

func addInteractionHandlers(s *state.State) {
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