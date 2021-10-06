package main

import (
	// "No3371.github.com/song_librarian.bot/logger"
	// "github.com/diamondburned/arikawa/v3/discord"
	// "github.com/diamondburned/arikawa/v3/gateway"
	"fmt"
	"strconv"
	"strings"

	"No3371.github.com/song_librarian.bot/logger"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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
					logger.Logger.Infof("Need atleast 2 options provided")
					return
				}
				logger.Logger.Infof("  DELETE")
				var cId uint64
				cId, err = strconv.ParseUint(data.Options[0].String(), 10, 64)
				if err != nil {
					logger.Logger.Infof("Failed to parse channel Id: %s", data.Options[0])
					return
				}
				deleted := make([]string, 0)
				for i := 1; i < len(data.Options); i++ {
					logger.Logger.Infof("  DELETE (%s): %s - %s", data.Name, data.Options[0], data.Options[i])
					var mId uint64
					mId, err = strconv.ParseUint(data.Options[i].String(), 10, 64)
					if err != nil {
						logger.Logger.Infof("Failed to parse message Id: %s", data.Options[i].String())
						return
					}
					var m *discord.Message
					if m, err = s.Message(discord.ChannelID(cId), discord.MessageID(mId)); err == nil && m != nil && m.Author.ID == me.ID {
						logger.Logger.Infof("  #%d Message fetched and it's mine.", i)
						err = s.DeleteMessage(m.ChannelID, m.ID, "Requested")
						if err != nil {
							logger.Logger.Errorf("  DeleteMessage err: %v", err)
							continue
						} else {
							deleted = append(deleted, m.URL())
							logger.Logger.Infof("  Deleted.")
						}
					}
				}
				sb := strings.Builder{}
				_, err = sb.WriteString(fmt.Sprintf("Deleted %d/%d \n", len(deleted), len(data.Options) - 1))
				for _, url := range deleted {
					sb.WriteString("Deleted ")
					sb.WriteString(url)
					sb.WriteRune('\n')
				}
				if err != nil {
					logger.Logger.Errorf("  Failed to prepare response: %v", err)
					return
				}
				err = s.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Content: option.NewNullableString(sb.String()),
					},
				})
				if err != nil {
					logger.Logger.Errorf("  Failed to respond: %v", err)
					return
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