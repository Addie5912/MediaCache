package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	stdLogger *log.Logger
	level     = "INFO"
)

// Init 初始化日志系统
func Init(logFile, logLevel string) {
	level = strings.ToUpper(logLevel)
	var output *os.File
	var err error
	if logFile != "" {
		output, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			output = os.Stdout
		}
	} else {
		output = os.Stdout
	}
	stdLogger = log.New(output, "", log.LstdFlags)
}

func logf(prefix, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if stdLogger == nil {
		fmt.Printf("%s %s\n", prefix, msg)
		return
	}
	stdLogger.Printf("%s %s", prefix, msg)
}

// Infof 记录 INFO 级别日志
func Infof(format string, args ...interface{}) {
	logf("[INFO]", format, args...)
}

// Info 记录 INFO 级别日志（非格式化兼容）
func Info(format string, args ...interface{}) {
	logf("[INFO]", format, args...)
}

// Warnf 记录 WARN 级别日志
func Warnf(format string, args ...interface{}) {
	logf("[WARN]", format, args...)
}

// Warn 记录 WARN 级别日志
func Warn(format string, args ...interface{}) {
	logf("[WARN]", format, args...)
}

// Debugf 记录 DEBUG 级别日志（仅 DEBUG 模式下输出）
func Debugf(format string, args ...interface{}) {
	if level != "DEBUG" {
		return
	}
	logf("[DEBUG]", format, args...)
}

// Debug 记录 DEBUG 级别日志
func Debug(format string, args ...interface{}) {
	if level != "DEBUG" {
		return
	}
	logf("[DEBUG]", format, args...)
}

// Errorf 记录 ERROR 级别日志
func Errorf(format string, args ...interface{}) {
	logf("[ERROR]", format, args...)
}

// Error 记录 ERROR 级别日志
func Error(format string, args ...interface{}) {
	logf("[ERROR]", format, args...)
}

// TeeErrorf 记录 ERROR 级别日志并返回 error 对象
func TeeErrorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	logf("[ERROR]", "%s", msg)
	return fmt.Errorf("%s", msg)
}

// Fatalf 记录 FATAL 级别日志并退出程序
func Fatalf(format string, args ...interface{}) {
	logf("[FATAL]", format, args...)
	os.Exit(1)
}
