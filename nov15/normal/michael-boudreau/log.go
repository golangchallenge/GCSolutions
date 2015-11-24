package main

import "log"

type LogLevel int

const (
	SilentLogLevel LogLevel = iota
	ErrorLogLevel
	DebugLogLevel
	TraceLogLevel
)

type LevelLogger struct {
	Level LogLevel
}

func NewLevelLogger(level LogLevel) *LevelLogger {
	return &LevelLogger{level}
}
func NewDebugLevelLogger() *LevelLogger {
	return &LevelLogger{DebugLogLevel}
}
func NewTraceLevelLogger() *LevelLogger {
	return &LevelLogger{TraceLogLevel}
}
func NewStandardLevelLogger() *LevelLogger {
	return &LevelLogger{ErrorLogLevel}
}
func NewSilentLevelLogger() *LevelLogger {
	return &LevelLogger{SilentLogLevel}
}

func (l *LevelLogger) Trace(v ...interface{}) {
	if l.Level >= TraceLogLevel {
		log.Print(v...)
	}
}

func (l *LevelLogger) Tracef(format string, v ...interface{}) {
	if l.Level >= TraceLogLevel {
		log.Printf(format, v...)
	}
}

func (l *LevelLogger) Traceln(v ...interface{}) {
	if l.Level >= TraceLogLevel {
		log.Println(v...)
	}
}
func (l *LevelLogger) Debug(v ...interface{}) {
	if l.Level >= DebugLogLevel {
		log.Print(v...)
	}
}

func (l *LevelLogger) Debugf(format string, v ...interface{}) {
	if l.Level >= DebugLogLevel {
		log.Printf(format, v...)
	}
}

func (l *LevelLogger) Debugln(v ...interface{}) {
	if l.Level >= DebugLogLevel {
		log.Println(v...)
	}
}

func (l *LevelLogger) Print(v ...interface{}) {
	if l.Level >= ErrorLogLevel {
		log.Print(v...)
	}
}

func (l *LevelLogger) Printf(format string, v ...interface{}) {
	if l.Level >= ErrorLogLevel {
		log.Printf(format, v...)
	}
}

func (l *LevelLogger) Println(v ...interface{}) {
	if l.Level >= ErrorLogLevel {
		log.Println(v...)
	}
}
