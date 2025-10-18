// 日志轮转(Log Rotation)
package logx

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const (
	m1024 = 1024
	// 单位：字节 (unit: byte)
	Byte = 1
	KB   = m1024 * Byte
	MB   = m1024 * KB
	GB   = m1024 * MB

	MB100 = 100 * MB // 100MB
)

func NewJsonRotation(level LogLevel, filename string, maxSize int64, maxBackups int) (*Logger, error) {
	var out, currentSize, newFilename, err = check(filename)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		out:           out,
		level:         level,
		withTime:      true,
		withLocation:  0,
		callDepth:     4,
		marshalIndent: false,

		withRotation: true,
		filename:     newFilename,
		maxSize:      maxSize,
		maxBackups:   maxBackups,
		currentSize:  currentSize,
	}
	return l, nil
}

func getFilename(appName string) string {
	if filepath.Ext(appName) == ".log" {
		return appName
	}
	return appName + ".log"
}
func check(appName string) (io.Writer, int64, string, error) {
	var out *os.File
	var currentSize int64 = 0

	filename := getFilename(appName)
	if _, err := os.Stat(filepath.Dir(filename)); err != nil {
		return nil, -1, "", err
	}

	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, -1, "", err
	}

	info, err := out.Stat()
	if err != nil {
		return nil, -1, "", err
	}
	currentSize = info.Size()

	return out, currentSize, filename, nil
}

func (r *Logger) backupName(i int) string {
	if i == 0 {
		return r.filename
	}
	return r.filename + "." + strconv.Itoa(i)
}
func (r *Logger) rotate() error {
	if r.out != nil {
		r.out.(*os.File).Close()
	}

	// 重命名现有文件
	for i := r.maxBackups - 1; i >= 0; i-- {
		oldName := r.backupName(i)
		newName := r.backupName(i + 1)

		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}

	// 创建新文件
	file, err := os.OpenFile(r.filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	r.out = file
	r.currentSize = 0
	return nil
}
