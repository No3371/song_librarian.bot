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
	Unsubscribe
	Resubscribe
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

	if cmd, exist := required[commandNameUnsub]; !exist {
		logger.Logger.Infof("[LOG] Creating command: %s", commandNameUnsub)
		cmdR, err := s.CreateCommand(_appid, api.CreateCommandData{
			Type: discord.ChatInputCommand,
			Name:        commandNameUnsub,
			Description: locale.C_DESC_UNSUB,
			NoDefaultPermission: false,
		})

		if err != nil {
			logger.Logger.Errorf("Command creation error: %v", err)
		} else {
			logger.Logger.Infof("DELETE command created: %d", cmdR.ID)
			sv.SaveCommandId(int(Unsubscribe), uint64(cmdR.ID), 0)
			commandIdMap[cmdR.ID] = Unsubscribe
		}
	} else {
		var savedCmdId uint64
		var savedCmdVersion uint32
		savedCmdId, savedCmdVersion, err = sv.LoadCommandId(int(Unsubscribe))
		if err != nil {
			logger.Logger.Errorf("Command Unsubscribe loading error: %v", err)
		} else if savedCmdId != uint64(cmd.ID) {
			logger.Logger.Errorf("Command ID Mismatch! Overriding with online ID!")
			sv.SaveCommandId(int(Unsubscribe), uint64(cmd.ID), savedCmdVersion)
			commandIdMap[discord.CommandID(savedCmdId)] = Unsubscribe
		} else {
			logger.Logger.Infof("Unsubscribe command loaded: %d", savedCmdId)
			commandIdMap[discord.CommandID(savedCmdId)] = Unsubscribe
		}
	}

	
	if cmd, exist := required[commandNameResub]; !exist {
		logger.Logger.Infof("[LOG] Creating command: %s", commandNameResub)
		cmdR, err := s.CreateCommand(_appid, api.CreateCommandData{
			Type: discord.ChatInputCommand,
			Name:        commandNameResub,
			Description: locale.C_DESC_RESUB,
			NoDefaultPermission: false,
		})

		if err != nil {
			logger.Logger.Errorf("Command creation error: %v", err)
		} else {
			logger.Logger.Infof("DELETE command created: %d", cmdR.ID)
			sv.SaveCommandId(int(Resubscribe), uint64(cmdR.ID), 0)
			commandIdMap[cmdR.ID] = Resubscribe
		}
	} else {
		var savedCmdId uint64
		var savedCmdVersion uint32
		savedCmdId, savedCmdVersion, err = sv.LoadCommandId(int(Resubscribe))
		if err != nil {
			logger.Logger.Errorf("Command Resubscribe loading error: %v", err)
		} else if savedCmdId != uint64(cmd.ID) {
			logger.Logger.Errorf("Command ID Mismatch! Overriding with online ID!")
			sv.SaveCommandId(int(Resubscribe), uint64(cmd.ID), savedCmdVersion)
			commandIdMap[discord.CommandID(savedCmdId)] = Resubscribe
		} else {
			logger.Logger.Infof("Resubscribe command loaded: %d", savedCmdId)
			commandIdMap[discord.CommandID(savedCmdId)] = Resubscribe
		}
	}

	if cmd, exist := required[commandNameDelete]; !exist {
		logger.Logger.Infof("[LOG] Creating command: %s", commandNameDelete)
		cmdR, err := s.CreateCommand(_appid, api.CreateCommandData{
			Type: discord.ChatInputCommand,
			Name:        commandNameDelete,
			Description: locale.C_DESC,
			NoDefaultPermission: false,
			Options:     []discord.CommandOption{
				{
					Type:        discord.StringOption,
					Name:        "c",
					Description: "jibberish",
					Required:    true,
				},
				{
					Type:        discord.StringOption,
					Name:        "m1",
					Description: "jibberish",
					Required:    true,
				},
				{
					Type:        discord.StringOption,
					Name:        "m2",
					Description: "jibberish",
					Required:    false,
				},
				{
					Type:        discord.StringOption,
					Name:        "m3",
					Description: "jibberish",
					Required:    false,
				},
			},
			// Options:     []discord.CommandOption{
			// 	{
			// 		Type:        discord.ChannelOption,
			// 		Name:        "cid",
			// 		Required:    true,
			// 	},
			// 	{
			// 		Type:        discord.IntegerOption,
			// 		Name:        "msgid",
			// 		Required:    true,
			// 	},
			// },
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
		if err != nil {
			logger.Logger.Errorf("Command DeleteRedirectedMessage loading error: %v", err)
		} else if savedCmdId != uint64(cmd.ID) {
			logger.Logger.Errorf("Command ID Mismatch! Overriding with online ID!")
			sv.SaveCommandId(int(DeleteRedirectedMessage), uint64(cmd.ID), savedCmdVersion)
			commandIdMap[discord.CommandID(savedCmdId)] = DeleteRedirectedMessage
		} else {
			logger.Logger.Infof("DeleteRedirectedMessage command loaded: %d", savedCmdId)
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