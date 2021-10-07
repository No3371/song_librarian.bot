package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func SetupLogger (prod bool) {
	if prod {
		nl, _ := zap.NewDevelopment(zap.WithCaller(false))
		Logger = nl.Sugar()
	} else {
		nl, _ := zap.NewDevelopment()
		Logger = nl.Sugar()
	}
}
