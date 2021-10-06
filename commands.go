package main

import (
	"No3371.github.com/song_librarian.bot/locale"
	"No3371.github.com/song_librarian.bot/logger"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

type CommandDefinition int

const (
	DeleteRedirectedMessage CommandDefinition = iota
)

var commandIdMap map[discord.CommandID]CommandDefinition

func assureCommands (s *state.State) (err error) {
	if commandIdMap == nil {
		commandIdMap = make(map[discord.CommandID]CommandDefinition)
	}
	var commands []discord.Command
	_appid := discord.AppID(*globalFlags.appid)
	commands, err = s.Commands(_appid)
	if err != nil {
		return err
	}

	var required map[string]discord.Command = make(map[string]discord.Command)

	for _, c := range commands {
		logger.Logger.Infof("Online command: %s", c.Name)
		required[c.Name] = c
	}

	// if _, exist := required[commandNameChannel]; !exist {
	// 	logger.Logger.Infof("[LOG] Creating command: %s", commandNameChannel)
	// 	var cmd *discord.Command
	// 	cmd, err = s.CreateCommand(_appid, api.CreateCommandData{
	// 		Name:        commandNameChannel,
	// 		Description: locale.C_DESC,
	// 		Options:     []discord.CommandOption{
	// 			{
	// 				Type:        discord.ChannelOption,
	// 				Name:        "c_original",
	// 				Description: locale.C_ORIGINAL_DESC,
	// 				Required:    true,
	// 			},
	// 			{
	// 				Type:        discord.ChannelOption,
	// 				Name:        "c_cover",
	// 				Description: locale.C_COVER_DESC,
	// 				Required:    true,
	// 			},
	// 		},
	// 		NoDefaultPermission: false,
	// 		Type: discord.ChatInputCommand,
	// 	})
	// 	if err != nil {
	// 		logger.Logger.Errorf("Command creation error: %v", err)
	// 	}
	// 	cmd.NoDefaultPermission = true

	// 	var _cmd *discord.Command
	// 	_cmd, err = s.Command(discord.AppID(*globalFlags.appid), cmd.ID)
	// 	if _cmd == nil {
	// 		logger.Logger.Errorf("Failed to fetch command created: %v", cmd)
	// 	} else {
	// 		logger.Logger.Infof("Command created: %v", _cmd)
	// 		commandIdMap[namec] = DeleteRedirectedMessage
	// 	}
	// }

	// if _, exist := required[commandNameDelete]; !exist {
	// 	logger.Logger.Infof("[LOG] Creating command: %s", commandNameChannel)
	// 	var cmd *discord.Command
	// 	cmd, err = s.CreateCommand(_appid, api.CreateCommandData{
	// 		Name:        commandNameDelete,
	// 		Description: locale.C_DESC,
	// 		Options:     []discord.CommandOption{
	// 			{
	// 				Type:        discord.NumberOption,
	// 				Name:        "id",
	// 				Description: locale.C_DELETE_ID_DESC,
	// 				Required:    true,
	// 			},
	// 		},
	// 		NoDefaultPermission: false,
	// 		Type: discord.ChatInputCommand,
	// 	})
	// 	cmd.Type
	// 	if err != nil {
	// 		logger.Logger.Errorf("Command creation error: %v", err)
	// 	}

	if cmd, exist := required[commandNameDelete]; !exist {
		logger.Logger.Infof("[LOG] Creating command: %s", commandNameDelete)
		cmdR, err := s.CreateCommand(_appid, api.CreateCommandData{
			Type: discord.ChatInputCommand,
			Name:        commandNameDelete,
			Description: locale.C_DESC,
			Options:     []discord.CommandOption{
				{
					Type:        discord.ChannelOption,
					Name:        "cid",
					Required:    true,
				},
				{
					Type:        discord.IntegerOption,
					Name:        "msgid",
					Required:    true,
				},
			},
		})

		if err != nil {
			logger.Logger.Errorf("Command creation error: %v", err)
		} else {
			logger.Logger.Infof("DELETE command created: %d", cmdR.ID)
			sv.SaveCommandId(int(DeleteRedirectedMessage), uint64(cmdR.ID), 0)
			commandIdMap[cmdR.ID] = DeleteRedirectedMessage
		}
	} else {
		var savedCmdId uint64
		var savedCmdVersion uint32
		savedCmdId, savedCmdVersion, err = sv.LoadCommandId(int(DeleteRedirectedMessage))
		if savedCmdId != uint64(cmd.ID) {
			logger.Logger.Errorf("Command ID Mismatch! Overriding with online ID!")
			sv.SaveCommandId(int(DeleteRedirectedMessage), uint64(cmd.ID), savedCmdVersion)
			commandIdMap[discord.CommandID(savedCmdId)] = DeleteRedirectedMessage
		} else {
			commandIdMap[discord.CommandID(savedCmdId)] = DeleteRedirectedMessage
		}
	}

	// if cmd, exist := required[commandNameDelete]; !exist {
	// 	logger.Logger.Infof("[LOG] Creating command: %s", commandNameChannel)
	// 	cmdR, err := s.CreateCommand(_appid, api.CreateCommandData{
	// 		Name:        commandNameDelete,
	// 		NoDefaultPermission: false,
	// 		Type: discord.MessageCommand,
	// 	})
	// 	if err != nil {
	// 		logger.Logger.Errorf("Command creation error: %v", err)
	// 	} else {
	// 		logger.Logger.Infof("DELETE command created: %d", cmdR.ID)
	// 		sv.SaveCommandId(int(DeleteRedirectedMessage), uint64(cmdR.ID))
	// 		commandIdMap[cmdR.ID] = DeleteRedirectedMessage
	// 	}
	// } else {
	// 	var savedCmdId uint64
	// 	savedCmdId, err = sv.LoadCommandId(int(DeleteRedirectedMessage))
	// 	if savedCmdId != uint64(cmd.ID) {
	// 		logger.Logger.Errorf("Command ID Mismatch! Overriding with online ID!")
	// 		sv.SaveCommandId(int(DeleteRedirectedMessage), uint64(cmd.ID))
	// 		commandIdMap[discord.CommandID(savedCmdId)] = DeleteRedirectedMessage
	// 	} else {
	// 		commandIdMap[discord.CommandID(savedCmdId)] = DeleteRedirectedMessage
	// 	}
	// }

	// 	var _cmd *discord.Command
	// 	_cmd, err = s.Command(discord.AppID(*globalFlags.appid), cmd.ID)
	// 	if _cmd == nil {
	// 		logger.Logger.Errorf("Failed to fetch command created: %v", cmd)
	// 	} else {
	// 		logger.Logger.Infof("Command created: %v", _cmd)
	// 	}
	// }

	return nil
}


func resetAllCommands (s *state.State) (err error) {
	
	var commands []discord.Command
	commands, err = s.Commands(discord.AppID(*globalFlags.appid))
	if err != nil {
		logger.Logger.Errorf("Failed to get owned commands: %v", err)
		return err
	}

	for _, c := range  commands{
		err = s.DeleteCommand(c.AppID, c.ID)
		logger.Logger.Infof("Deleting command: %s", c.Name)
		if err != nil {
			logger.Logger.Errorf("Failed to delete command: %v", err)
			return err
		}
	}

	err = assureCommands(s)
	if err != nil {
		logger.Logger.Fatalf("Failed to assure commands: %v", err)
		return err
	}

	return nil
}