package utils

import (
	"os"

	logx "github.com/JiuYu77/go-email/log"
)

var (
	Logger    *logx.Logger
	LogPrefix = "[go-email] "
)

func init() {
	if os.Getenv("GO_EMAIL_MODE") == "" {
		Logger = logx.NewLogger(os.Stdout, logx.DebugLevel)
	} else {
		Logger = logx.NewLogger(os.Stdout, logx.InfoLevel)
	}
	Logger.EnableColor()
	Logger.SetLocation(1)
}

func SetMode(mode logx.LogLevel) {
	Logger.SetLevel(mode)
}
