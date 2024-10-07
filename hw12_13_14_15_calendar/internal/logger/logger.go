package logger

import (
	"fmt"
	"time"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

const logTemplate = "%s [%s] %s\n"

func (l LogLevel) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(level string) (LogLevel, error) {
	switch level {
	case "DEBUG":
		return Debug, nil
	case "INFO":
		return Info, nil
	case "WARN":
		return Warn, nil
	case "ERROR":
		return Error, nil
	default:
		return Error, fmt.Errorf("unknown log level: %s", level)
	}
}

type Logger struct {
	level LogLevel
}

func New(level string) (*Logger, error) {
	l, err := ParseLevel(level)
	if err != nil {
		return nil, err
	}
	return &Logger{
		level: l,
	}, nil
}

func (l Logger) Debug(msg string) {
	if l.level <= Debug {
		l.logf(Debug, msg)
	}
}

func (l Logger) Info(msg string) {
	if l.level <= Info {
		l.logf(Info, msg)
	}
}

func (l Logger) Warn(msg string) {
	if l.level <= Warn {
		l.logf(Warn, msg)
	}
}

func (l Logger) Error(msg string) {
	if l.level <= Error {
		l.logf(Error, msg)
	}
}

func (l Logger) logf(level LogLevel, message string) {
	fmt.Printf(logTemplate, time.Now().Format("2006-01-02 15:04:05.000"), level, message)
}
