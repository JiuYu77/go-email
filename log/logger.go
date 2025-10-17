package logx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Logger struct {
	mtx          sync.Mutex
	out          io.Writer
	level        LogLevel
	withColor    bool
	withTime     bool
	withLocation bool
	callDepth    int
}

// NewLogger 创建新日志器
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		out:          out,
		level:        level,
		withColor:    false,
		withTime:     false,
		withLocation: true,
		callDepth:    2,
	}
}

// 启用颜色输出
func (l *Logger) EnableColor() {
	l.withColor = true
}
func (l *Logger) EnableTime() {
	l.withTime = true
}
func (l *Logger) DisableLocation() {
	l.withLocation = false
}

// 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.level = level
}

// 设置调用深度
func (l *Logger) SetCallDepth(depth int) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.callDepth = depth
}

// 获取颜色代码
func (l *Logger) getColor(level LogLevel) Style {
	if !l.withColor {
		return Reset
	}
	level_color := _levelToColor[level]
	return level_color
}

// 输出日志
func (l *Logger) output(level LogLevel, format string, v ...any) {
	if level < l.level {
		return
	}

	l.mtx.Lock()
	defer l.mtx.Unlock()

	// 获取调用者信息
	var location string
	if l.withLocation {
		_, file, line, ok := runtime.Caller(l.callDepth)
		if !ok {
			file = "???"
			line = 0
		} else {
			file = filepath.Base(file)
		}
		location = fmt.Sprintf(" %s:%d", file, line)
	}

	// 格式化时间
	now := ""
	if l.withTime {
		now = time.Now().Format("2006-01-02 15:04:05.000") + " "
	}

	color := l.getColor(level)

	// 构建日志内容
	var msg string
	if format == "" {
		msg = fmt.Sprint(v...)
		if color != Reset {
			msg = color.Add(msg)
		}
	} else {
		msg = fmt.Sprintf(format, v...)
	}

	// 构建完整日志行
	logLine := fmt.Sprintf("%s%s%s %s",
		now, color.Add("["+level.String()+"]"), location, msg)

	// 输出到 writer
	l.out.Write([]byte(logLine))
}

func (l *Logger) Appendln(v ...any) []any {
	return append(v, "\n")
}

// debug
func (l *Logger) Debug(v ...any) {
	l.output(DebugLevel, "", v...)
}
func (l *Logger) Debugf(format string, v ...any) {
	l.output(DebugLevel, format, v...)
}
func (l *Logger) Debugln(v ...any) {
	v = l.Appendln(v...)
	l.output(DebugLevel, "", v...)
}

// info
func (l *Logger) Info(v ...any) {
	l.output(InfoLevel, "", v...)
}
func (l *Logger) Infof(format string, v ...any) {
	l.output(InfoLevel, format, v...)
}
func (l *Logger) Infoln(v ...any) {
	v = l.Appendln(v...)
	l.output(InfoLevel, "", v...)
}

// warn
func (l *Logger) Warn(v ...any) {
	l.output(WarnLevel, "", v...)
}
func (l *Logger) Warnf(format string, v ...any) {
	l.output(WarnLevel, format, v...)
}
func (l *Logger) Warnln(v ...any) {
	v = l.Appendln(v...)
	l.output(WarnLevel, "", v...)
}

// error
func (l *Logger) Error(v ...any) {
	l.output(ErrorLevel, "", v...)
}
func (l *Logger) Errorf(format string, v ...any) {
	l.output(ErrorLevel, format, v...)
}
func (l *Logger) Errorln(v ...any) {
	v = l.Appendln(v...)
	l.output(ErrorLevel, "", v...)
}

// panic
func (l *Logger) Panic(v ...any) {
	l.output(PanicLevel, "", v...)
	s := fmt.Sprint(v...)
	panic(s) // 退出程序
}
func (l *Logger) Panicf(format string, v ...any) {
	l.output(PanicLevel, format, v...)
	s := fmt.Sprint(v...)
	panic(s)
}
func (l *Logger) Panicln(v ...any) {
	v = l.Appendln(v...)
	l.output(PanicLevel, "", v...)
	s := fmt.Sprint(v...)
	panic(s)
}

// fatal
func (l *Logger) Fatal(v ...any) {
	l.output(FatalLevel, "", v...)
	os.Exit(1) // 退出程序
}
func (l *Logger) Fatalf(format string, v ...any) {
	l.output(FatalLevel, format, v...)
	os.Exit(1)
}
func (l *Logger) Fatalln(v ...any) {
	v = l.Appendln(v...)
	l.output(FatalLevel, "", v...)
	os.Exit(1) // 退出程序
}
