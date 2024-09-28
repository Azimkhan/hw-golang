package logger

import "fmt"

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

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

func (l Logger) Info(msg string) {
	if l.level <= Info {
		fmt.Println(msg)
	}
}

func (l Logger) Error(msg string) {
	if l.level <= Error {
		fmt.Println(msg)
	}
}

// TODO
