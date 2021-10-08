package main

import (
	"time"

	flag "github.com/spf13/pflag"
)

type flags struct {
	dev *bool
	debugRegex *bool
	appid *int64
	token *string
	locale *string
	delay *time.Duration
	controlPort *uint16
	memSize *int
}

var globalFlags *flags

func resolveFlags () {
	if globalFlags == nil {
		globalFlags = &flags{}
	}
	globalFlags.dev = flag.Bool("dev", false, "development mode")
	globalFlags.debugRegex = flag.Bool("regex", false, "")
	globalFlags.appid = flag.Int64("appid", 0, "discord app id")
	globalFlags.token = flag.String("token", "", "discord bot token")
	globalFlags.locale = flag.String("locale", "TW", "locale code")
	globalFlags.delay = flag.Duration("delay", time.Minute, "")
	globalFlags.controlPort = flag.Uint16("cport", 11813, "")
	globalFlags.memSize = flag.Int("mem", 32, "")

	flag.Parse()
}