package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// defined log levels
const (
	LogLevelDebug   = "DEBUG"
	LogLevelInfo    = "INFO"
	LogLevelNotice  = "NOTICE"
	LogLevelWarning = "WARNING"
	LogLevelError   = "ERROR"
)

type logData struct {
	message string
	level   string
}

var (
	// a channel use to transfer log
	logChannel chan *logData
	logger     *log.Logger

	// a buffer use to write logs
	buf            bytes.Buffer
	currBufferSize uint16 = 0
)

// Start the log server
func logStart() {
	logChannel = make(chan *logData, conf.LogQueueSize)
	logger = log.New(&buf, "", log.LstdFlags)
	go func() {
		for {
			select {
			case aLog := <-logChannel:
				writeLog(aLog)
			}
		}
	}()
}

func Logf(level string, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	aLog := &logData{message, level}
	logChannel <- aLog
}

func Log(level string, a ...interface{}) {
	message := fmt.Sprint(a...)
	aLog := &logData{message, level}
	logChannel <- aLog
}

// Write logs to buffer. Flush logs to stdout if buffer size reach the configured size.
func writeLog(aLog *logData) {
	logger.Printf("[%s] %s\n", aLog.level, aLog.message)
	if aLog.level == LogLevelError {
		defer func() { os.Exit(1) }()
		logger.Printf("SHUTDOWN\n")
	}
	currBufferSize++
	if currBufferSize >= conf.LogBufferSize {
		FlushLog()
	}
}

func FlushLog() {
	log.Print(&buf)
	buf.Reset()
	currBufferSize = 0
}
