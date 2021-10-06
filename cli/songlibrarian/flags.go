package main

import (
	"time"

	"No3371.github.com/song_librarian.bot/logger"
	flag "github.com/spf13/pflag"
)

type flags struct {
	dev *bool
	appid *int64
	token *string
	locale *string
	delay *time.Duration
	controlPort *uint16
}

var globalFlags *flags

func resolveFlags () {
	if globalFlags == nil {
		globalFlags = &flags{}
	}
	globalFlags.dev = flag.Bool("dev", false, "development mode")
	globalFlags.appid = flag.Int64("appid", 0, "discord app id")
	globalFlags.token = flag.String("token", "", "discord bot token")
	globalFlags.locale = flag.String("locale", "TW", "locale code")
	globalFlags.delay = flag.Duration("delay", time.Minute, "")
	globalFlags.controlPort = flag.Uint16("cport", 11813, "")

	flag.Parse()
	
	logger.Logger.Infof("Flags resolved: %+v", globalFlags)
}