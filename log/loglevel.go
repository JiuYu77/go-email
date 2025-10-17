package logx

import (
	"fmt"
)

type LogLevel int8

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel // panic
	FatalLevel // os.Exit
)

// String 返回 LogLevel 的 首字母大写 ASCII 表示形式。
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "Debug"
	case InfoLevel:
		return "Info"
	case WarnLevel:
		return "Warn"
	case ErrorLevel:
		return "Error"
	case PanicLevel:
		return "Panic"
	case FatalLevel:
		return "Fatal"
	default:
		return fmt.Sprintf("LogLevel(%d)", l)
	}
}

// LowerString 返回 LogLevel 的 小写 ASCII 表示形式。
func (l LogLevel) LowerString() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("loglevel(%d)", l)
	}
}

// UpperString 返回 LogLevel 的 大写 ASCII 表示形式。
func (l LogLevel) UpperString() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("LOGLEVEL(%d)", l)
	}
}

var (
	_levelToColor = map[LogLevel]Style{
		DebugLevel: Cyan,    // 青色
		InfoLevel:  Green,   // 绿色
		WarnLevel:  Yellow,  // 黄色
		ErrorLevel: Red,     // 红色
		PanicLevel: Red,     // 红色
		FatalLevel: Magenta, // 紫色
	}
)
