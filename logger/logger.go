package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init () {
	SetupLogger(false)
}

func SetupLogger (prod bool) {
	if prod {
		nl, _ := zap.NewProduction()
		Logger = nl.Sugar()
	} else {
		nl, _ := zap.NewDevelopment()
		Logger = nl.Sugar()
	}
}
