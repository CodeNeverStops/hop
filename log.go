package main

import (
	"bytes"
	"fmt"
	"log"
)

const (
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelNotice  = "NOTICE"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
)

type logData struct {
	message string
	level   string
}

var (
	channel chan *logData
	buf     bytes.Buffer
	logger  *log.Logger
)

func init() {
	channel = make(chan *logData, 1000)
	logger = log.New(&buf, "", log.LstdFlags)
	go func() {
		for {
			select {
			case aLog := <-channel:
				writeLog(aLog)
			}
		}
	}()
}

func Logf(level string, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	aLog := &logData{message, level}
	channel <- aLog
}

func Log(level string, a ...interface{}) {
	message := fmt.Sprint(a...)
	aLog := &logData{message, level}
	channel <- aLog
}

func writeLog(aLog *logData) {
	logger.Printf("[%s] %s\n", aLog.level, aLog.message)
	log.Print(&buf)
}
