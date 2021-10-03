package main

import (
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

func assureCommands (s *state.State) (err error) {
	var commands []discord.Command
	_appid := discord.AppID(*globalFlags.appid)
	commands, err = s.Commands(_appid)
	if err != nil {
		return err
	}

	var required map[string]discord.Command = make(map[string]discord.Command)

	for _, c := range commands {
		required[c.Name] = c
	}

	if _, exist := required[commandNameChannel]; !exist {
		logger.Logger.Infof("[LOG] Creating command: %s", commandNameChannel)
		var cmd *discord.Command
		cmd, err = s.CreateCommand(_appid, api.CreateCommandData{
			Name:        commandNameChannel,
			Description: locale.C_DESC,
			Options:     []discord.CommandOption{
				{
					Type:        discord.ChannelOption,
					Name:        "c_original",
					Description: locale.C_ORIGINAL_DESC,
					Required:    true,
				},
				{
					Type:        discord.ChannelOption,
					Name:        "c_cover",
					Description: locale.C_COVER_DESC,
					Required:    true,
				},
			},
			NoDefaultPermission: false,
			Type: discord.ChatInputCommand,
		})
		if err != nil {
			logger.Logger.Errorf("Command creation error: %v", err)
		}
		cmd.NoDefaultPermission = true

		var _cmd *discord.Command
		_cmd, err = s.Command(discord.AppID(*globalFlags.appid), cmd.ID)
		if _cmd == nil {
			logger.Logger.Errorf("Failed to fetch command created: %v", cmd)
		} else {
			logger.Logger.Infof("Command created: %v", _cmd)
		}
	}

	return nil
}


func resetAllCommands (s *state.State) (err error) {
	
	var commands []discord.Command
	commands, err = s.Commands(discord.AppID(*globalFlags.appid))
	if err != nil {
		logger.Logger.Errorf("[MAIN] %v", err)
		return err
	}
	for _, c := range  commands{
		err = s.DeleteCommand(c.AppID, c.ID)
		if err != nil {
			logger.Logger.Errorf("[MAIN] %v", err)
			return err
		}
	}

	err = assureCommands(s)
	if err != nil {
		logger.Logger.Fatalf("[MAIN] %v", err)
		return err
	}

	return nil
}