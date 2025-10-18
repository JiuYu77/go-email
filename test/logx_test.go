package test

import (
	"fmt"
	"os"
	"testing"

	logx "github.com/JiuYu77/go-email/log"
)

func TestLogx(t *testing.T) {
	logger := logx.NewLogger(os.Stdout, logx.DebugLevel)
	// logger := logx.NewLogger(os.Stdout, logx.ErrorLevel)
	logger.EnableColor()
	logger.EnableTime()
	// logger.SetLocation(-1)

	logger.Debug("123\n", 334)

	logger.Debugf("%s \033[32m%s, %d\033[0m\n", "[goemail]", "123", 334)
	logger.Debugf("\033[32m%s, %d\033[0m\n", "123", 334)

	fmt.Println("------------------------")

	logger.Debugln(123, 334)
	logger.Infoln(123, 334)
	logger.Warnln(123, 334)
	logger.Errorln(123, 334)
	// logger.Panicln(123, 334)
	// logger.Fatalln(123, 334)

	logger.SetCallDepth(1)
	logger.SetPrettyPrint(true)

	logger.DebugJson("This is a json data.", nil)
	logger.DebugJson(
		"This is a json data!",
		logx.Fields{
			"0": 0,
			"1": map[string]any{"1.1": 1.1, "1.2": "1.2"},
		},
	)
}

func TestLogxFile(t *testing.T) {
	name := "../tmp/app.log.jsonl"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	// file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		file, err = os.Create(name)
		if err != nil {
			panic(err)
		}
	}

	// logger := logx.NewLogger(file, logx.DebugLevel)
	// logger.SetCallDepth(1)

	logger := logx.NewJsonLogger(file, logx.DebugLevel)
	// logger := logx.NewJsonLogger(file, logx.InfoLevel)

	// logger.SetPrettyPrint(true)

	logger.DebugJson("This is a json data.", nil)
	logger.DebugJson(
		"This is a json data!",
		logx.Fields{
			"0": 0,
			"1": map[string]any{"1.1": 1.1, "1.2": "1.2"},
		},
	)
}

func TestLogxRotation(t *testing.T) {
	logger, err := logx.NewJsonRotation(logx.DebugLevel, "../tmp/logx", 1024, 3)
	if err != nil {
		t.Errorf("NewJsonRotation error: %v", err)
	}
	logger.DebugJson("This is a json Rotation data", nil)
	logger.DebugJson(
		"This is a json Rotation data!",
		logx.Fields{
			"0": 0,
			"1": map[string]any{"1.1": 1.1, "1.2": "1.2"},
		},
	)
}
