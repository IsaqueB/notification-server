package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	debug *log.Logger
}

func New() *Logger {
	flags := log.LstdFlags | log.Lshortfile

	return &Logger{
		info:  log.New(os.Stdout, fmt.Sprintf("%s - [INFO] ", (time.Now()).Format("02-01-2006 15:04:05")), flags),
		warn:  log.New(os.Stdout, fmt.Sprintf("%s - [WARN] ", (time.Now()).Format("02-01-2006 15:04:05")), flags),
		error: log.New(os.Stderr, fmt.Sprintf("%s - [ERROR] ", (time.Now()).Format("02-01-2006 15:04:05")), flags),
		debug: log.New(os.Stdout, fmt.Sprintf("%s - [DEBUG] ", (time.Now()).Format("02-01-2006 15:04:05")), flags),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.error.Println(v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.debug.Println(v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.error.Fatal(v...)
}
