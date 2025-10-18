package logx

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

type Fields map[string]any

type LogEntry struct {
	Timestamp string `json:"time_stamp"`
	Level     string `json:"log_level"`
	Message   string `json:"message,omitempty"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Fields    Fields `json:"fields,omitempty"`
}

// NewJsonLogger 创建json日志器
func NewJsonLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		out:           out,
		level:         level,
		withTime:      true,
		withLocation:  0,
		callDepth:     4,
		marshalIndent: false,
	}
}

// 美化输出（开发环境）
func (l *Logger) SetPrettyPrint(enable bool) {
	l.mtx2.Lock()
	defer l.mtx2.Unlock()
	l.marshalIndent = enable
}

func (l *Logger) Marshal(v any) ([]byte, error) {
	if !l.marshalIndent {
		return json.Marshal(v)
	} else {
		return json.MarshalIndent(v, "", "  ")
	}
}

func (l *Logger) outputJSON(level LogLevel, message string, fields Fields) {
	if level < l.level {
		return
	}

	l.mtx2.Lock()
	defer l.mtx2.Unlock()

	// 构建日志条目
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Fields:    fields,
	}

	// 添加位置信息
	if l.withLocation >= 0 {
		entry.File, entry.Line = l.getLocation()
	}

	// 序列化为 JSON
	jsonData, err := l.Marshal(entry)
	if err != nil {
		// 如果 JSON 序列化失败，回退到普通文本输出
		l.callDepth += 1 // callDepth 调用深度增加1
		l.output(level, "JSON marshal error: %v\n", err)
		l.callDepth -= 1 // 恢复
		return
	}

	if l.withRotation { // Log Rotation
		if l.currentSize+int64(len(jsonData)) > l.maxSize {
			if err := l.rotate(); err != nil {
				l.callDepth += 1 // callDepth 调用深度增加1
				l.output(level, "logger rotate error: %v\n", err)
				l.callDepth -= 1 // 恢复
				return
			}
		}
	}

	// 输出 JSON
	jsonData = append(jsonData, '\n')
	n, err := l.out.Write(jsonData)

	if l.withRotation { // Log Rotation
		l.currentSize += int64(n)
		if err != nil {
			l.callDepth += 1 // callDepth 调用深度增加1
			l.output(level, "logger write error: %v\n", err)
			l.callDepth -= 1 // 恢复
		}
	}
}

// WithFields 记录带字段的日志
//
// logger.WithFields(level, "msg", nil)
// logger.WithFields(level, "msg", map[string]any{"0":0, "1":"1"})
func (l *Logger) WithFields(level LogLevel, message string, fields Fields) {
	l.outputJSON(level, message, fields)
}

// DebugWithFields Debug 级别带字段
func (l *Logger) DebugJson(message string, fields Fields) {
	l.WithFields(DebugLevel, message, fields)
}

// InfoWithFields Info 级别带字段
func (l *Logger) InfoJson(message string, fields Fields) {
	l.WithFields(InfoLevel, message, fields)
}

// WarnWithFields Warn 级别带字段
func (l *Logger) WarnJson(message string, fields Fields) {
	l.WithFields(WarnLevel, message, fields)
}

// ErrorWithFields Error 级别带字段
func (l *Logger) ErrorJson(message string, fields Fields) {
	l.WithFields(ErrorLevel, message, fields)
}

// PanicWithFields Panic 级别带字段
func (l *Logger) PanicJson(message string, fields Fields) {
	l.WithFields(PanicLevel, message, fields)
	panic(message)
}

// FatalWithFields Fatal 级别带字段
func (l *Logger) FatalWithFields(message string, fields Fields) {
	l.WithFields(FatalLevel, message, fields)
	os.Exit(1)
}
