package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	level   Level
	prefix  string
	buf     []byte
}

func New(out io.Writer, level Level, prefix string) *Logger {
	return &Logger{
		out:    out,
		level:  level,
		prefix: prefix,
		buf:    make([]byte, 0, 256),
	}
}

func Default() *Logger {
	return New(os.Stderr, LevelInfo, "")
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.buf = l.buf[:0]

	l.buf = append(l.buf, '[')
	l.buf = append(l.buf, time.Now().Format(time.RFC3339)...)
	l.buf = append(l.buf, "] ["...)
	l.buf = append(l.buf, level.String()...)
	l.buf = append(l.buf, "] "...)

	if l.prefix != "" {
		l.buf = append(l.buf, l.prefix...)
		l.buf = append(l.buf, ' ')
	}

	l.buf = append(l.buf, fmt.Sprintf(format, args...)...)
	l.buf = append(l.buf, '\n')

	l.out.Write(l.buf)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

func (l *Logger) WithPrefix(prefix string) *Logger {
	newPrefix := l.prefix
	if newPrefix != "" {
		newPrefix = newPrefix + " " + prefix
	} else {
		newPrefix = prefix
	}
	return New(l.out, l.level, newPrefix)
}
