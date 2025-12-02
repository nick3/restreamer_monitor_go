package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LoggerWrapper 兼容标准库log的包装器
type LoggerWrapper struct {
	entry *logrus.Entry
}

// NewWrapper 创建包装器
func NewWrapper(fields map[string]interface{}) *LoggerWrapper {
	return &LoggerWrapper{
		entry: GetLogger(fields),
	}
}

// Printf 兼容标准库log.Printf
func (w *LoggerWrapper) Printf(format string, v ...interface{}) {
	w.entry.Infof(format, v...)
}

// Print 兼容标准库log.Print
func (w *LoggerWrapper) Print(v ...interface{}) {
	w.entry.Info(fmt.Sprint(v...))
}

// Println 兼容标准库log.Println
func (w *LoggerWrapper) Println(v ...interface{}) {
	w.entry.Info(fmt.Sprintln(v...))
}

// Fatal 兼容标准库log.Printf
func (w *LoggerWrapper) Fatal(v ...interface{}) {
	w.entry.Fatal(fmt.Sprint(v...))
}

// Fatalf 兼容标准库log.Fatalf
func (w *LoggerWrapper) Fatalf(format string, v ...interface{}) {
	w.entry.Fatalf(format, v...)
}

// Panic 兼容标准库log.Panic
func (w *LoggerWrapper) Panic(v ...interface{}) {
	w.entry.Panic(fmt.Sprint(v...))
}

// Panicf 兼容标准库log.Panicf
func (w *LoggerWrapper) Panicf(format string, v ...interface{}) {
	w.entry.Panicf(format, v...)
}

// DefaultWrapper 全局默认包装器实例
// 使用方式：logger.DefaultWrapper.Printf("format", args...)
var DefaultWrapper *LoggerWrapper

func init() {
	DefaultWrapper = NewWrapper(nil)
}
