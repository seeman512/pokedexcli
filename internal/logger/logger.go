package logger

import (
	"io"
	"log"
)

type Level int
const (
    TRACE Level = iota
    DEBUG
    INFO
    WARN
    ERROR
)

type Logger struct {
    Level Level
    Logger *log.Logger
}

func New(level Level, out io.Writer, prefix string) Logger {
    logger := Logger{
        Level: level,
        Logger: log.New(out, prefix, log.LstdFlags),
    }

    return logger
}

func (l Logger) log(minLevel Level, s string) {
    if l.Level < minLevel {
        return
    }

    l.Logger.Println(s)
}

func (l Logger) Trace(s string) {
    l.log(TRACE, s) 
}

func (l Logger) Debug(s string) {
    l.log(DEBUG, s) 
}

func (l Logger) Info(s string) {
    l.log(INFO, s) 
}

func (l Logger) WARN(s string) {
    l.log(WARN, s) 
}

func (l Logger) Error(s string) {
    l.log(ERROR, s) 
}
