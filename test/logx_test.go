package test

import (
	"fmt"
	"os"
	"testing"

	logx "github.com/JiuYu77/go-email/log"
)

func TestLogx(t *testing.T) {
	logger := logx.NewLogger(os.Stdout, logx.DebugLevel)
	logger.EnableColor()
	logger.EnableTime()
	// logger.DisableLocation()

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
}
