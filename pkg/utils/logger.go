package utils

import (
	"fmt"
	"os"
)

type Logger struct {
	prefix string
}

func NewLogger() *Logger {
	return &Logger{
		prefix: "🐍 eel-cli:",
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s Error: %s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Success(format string, args ...interface{}) {
	fmt.Printf("%s ✅ %s\n", l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Warning(format string, args ...interface{}) {
	fmt.Printf("%s ⚠️  %s\n", l.prefix, fmt.Sprintf(format, args...))
}
