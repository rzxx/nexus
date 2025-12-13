package logger

import (
	"fmt"
	"time"
)

const (
	LevelError = 0
	LevelInfo  = 1
	LevelDebug = 2
)

type Logger struct {
	Level int
}

func New(level int) *Logger {
	return &Logger{Level: level}
}

func (l *Logger) Log(level int, format string, args ...any) {
	if l.Level >= level {
		prefix := "[INFO]"
		if level == LevelDebug {
			prefix = "[DEBUG]"
		} else if level == LevelError {
			prefix = "[ERROR]"
		}

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("%s %s %s\n", timestamp, prefix, fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Info(format string, args ...any)  { l.Log(LevelInfo, format, args...) }
func (l *Logger) Debug(format string, args ...any) { l.Log(LevelDebug, format, args...) }
func (l *Logger) Error(format string, args ...any) { l.Log(LevelError, format, args...) }
