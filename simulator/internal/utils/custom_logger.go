package utils

import (
	"fmt"
	"log"
	"os"
)

// Logger 日志记录器类型
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	level       string
}

// 日志级别常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelNone  = "none"
)

// NewLogger 创建新的日志记录器
func NewLogger(level string) *Logger {
	// 设置默认日志级别
	if level == "" {
		level = LogLevelInfo
	}

	// 创建日志前缀
	debugPrefix := "\033[36m[DEBUG]\033[0m "
	infoPrefix := "\033[32m[INFO]\033[0m "
	warnPrefix := "\033[33m[WARN]\033[0m "
	errorPrefix := "\033[31m[ERROR]\033[0m "

	// 在Windows命令行上可能不支持颜色代码，可以使用简单前缀
	if os.Getenv("NO_COLOR") != "" {
		debugPrefix = "[DEBUG] "
		infoPrefix = "[INFO] "
		warnPrefix = "[WARN] "
		errorPrefix = "[ERROR] "
	}

	// 创建日志输出
	flags := log.Ltime

	return &Logger{
		debugLogger: log.New(os.Stdout, debugPrefix, flags),
		infoLogger:  log.New(os.Stdout, infoPrefix, flags),
		warnLogger:  log.New(os.Stdout, warnPrefix, flags),
		errorLogger: log.New(os.Stderr, errorPrefix, flags),
		level:       level,
	}
}

// Debug 输出调试日志
func (l *Logger) Debug(format string, v ...any) {
	if l.level == LogLevelDebug {
		l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info 输出信息日志
func (l *Logger) Info(format string, v ...any) {
	if l.level == LogLevelDebug || l.level == LogLevelInfo {
		l.infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warning 输出警告日志
func (l *Logger) Warning(format string, v ...any) {
	if l.level == LogLevelDebug || l.level == LogLevelInfo || l.level == LogLevelWarn {
		l.warnLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error 输出错误日志
func (l *Logger) Error(format string, v ...any) {
	if l.level != LogLevelNone {
		l.errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level string) {
	l.level = level
}

// LogToFile 将日志输出到文件
func (l *Logger) LogToFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// 设置日志输出
	flags := log.Ldate | log.Ltime
	l.debugLogger = log.New(file, "[DEBUG] ", flags)
	l.infoLogger = log.New(file, "[INFO] ", flags)
	l.warnLogger = log.New(file, "[WARN] ", flags)
	l.errorLogger = log.New(file, "[ERROR] ", flags)

	return nil
}
