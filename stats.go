package main

import (
	"errors"
	"fmt"
	"time"
)

// A type use to store server status
type ServerStats struct {
	TaskTotal  uint64
	TaskSucc   uint64
	TaskFail   uint64
	WorkerCurr uint16
	WorkerMax  uint16
	IsShutdown bool
	Version    string
	StartTime  time.Time
}

// Define status command
const (
	StatsCmdSuccTask = iota
	StatsCmdFailedTask
	StatsCmdNewWorker
	StatsCmdCloseWorker
	StatsCmdReport
	StatsCmdIsShutdown
	StatsCmdShutdown
)

const (
	ShutdownYes = "1"
	ShutdownNo  = "0"
)

// Status server handle status commands
func (stats *ServerStats) HandleCommand(cmd statsCmd) error {
	Logf(LogLevelInfo, "stats command: %d", cmd.cmd)
	switch cmd.cmd {
	case StatsCmdSuccTask:
		stats.TaskTotal++
		stats.TaskSucc++
	case StatsCmdFailedTask:
		stats.TaskTotal++
		stats.TaskFail++
	case StatsCmdNewWorker:
		stats.WorkerCurr++
		if stats.WorkerCurr > stats.WorkerMax {
			stats.WorkerMax = stats.WorkerCurr
		}
	case StatsCmdCloseWorker:
		stats.WorkerCurr--
		Logf(LogLevelInfo, "worker curr: %d", stats.WorkerCurr)
		if stats.WorkerCurr == 0 && stats.IsShutdown {
			shutdownCompChan <- true
		}
	case StatsCmdReport:
		cmd.replyChan <- stats.Report()
	case StatsCmdShutdown:
		Log(LogLevelInfo, "set shutdown to YES")
		stats.IsShutdown = true
	case StatsCmdIsShutdown:
		var reply string
		if stats.IsShutdown {
			reply = ShutdownYes
		} else {
			reply = ShutdownNo
		}
		Logf(LogLevelInfo, "is shutdown: %s", reply)
		cmd.replyChan <- reply
	default:
		Log(LogLevelWarning, "[stats server] command not found. command: %s", cmd)
		return errors.New("command not found")
	}
	return nil
}

// Show status report
func (stats *ServerStats) Report() string {
	uptime := UptimeFormat(uint32(time.Now().Sub(stats.StartTime)/time.Second), 2)
	var (
		taskSuccRatio   float64 = 0
		taskFailRatio   float64 = 0
		workerCurrRatio float32 = 0
		workerMaxRatio  float32 = 0
	)
	if stats.TaskTotal > 0 {
		taskSuccRatio = float64(stats.TaskSucc) / float64(stats.TaskTotal) * 100
		taskFailRatio = float64(stats.TaskFail) / float64(stats.TaskTotal) * 100
	}

	if conf.WorkerPoolSize > 0 {
		workerCurrRatio = float32(stats.WorkerCurr) / float32(conf.WorkerPoolSize) * 100
		workerMaxRatio = float32(stats.WorkerMax) / float32(conf.WorkerPoolSize) * 100
	}

	status := "online"
	if stats.IsShutdown {
		status = "closing..."
	}

	return fmt.Sprintf(`===============================
Version: %s 
Uptime: %s
Status: %s
Copyright (c) 2015 PerfectWorld
*******************************
Task Total:     %d
  Task Success:   %d (%0.2f%%)
  Task Failed:    %d (%0.2f%%)
Worker Config:  %d
  Worker Current: %d (%0.2f%%)
  Worker Max:     %d (%0.2f%%)
===============================`,
		Version,
		uptime,
		status,
		// task report
		stats.TaskTotal,
		stats.TaskSucc, taskSuccRatio,
		stats.TaskFail, taskFailRatio,
		// worker report
		conf.WorkerPoolSize,
		stats.WorkerCurr, workerCurrRatio,
		stats.WorkerMax, workerMaxRatio,
	)
}

// A channel use to transfer status command
var statsChannel chan statsCmd

// the global status server
var stats *ServerStats

// A type use to store status command
type statsCmd struct {
	cmd       int
	replyChan chan string
}

// Start status server to collect status of the server
func statsStart() {
	// init server status
	stats = &ServerStats{
		IsShutdown: false,
		Version:    Version,
		StartTime:  time.Now(),
	}
	poolSize := conf.WorkerPoolSize / 10
	if poolSize < 1 {
		poolSize = 1
	}
	statsChannel = make(chan statsCmd, poolSize)
	go func(stats *ServerStats) {
		for {
			select {
			case cmd := <-statsChannel:
				stats.HandleCommand(cmd)
			}
		}
	}(stats)
}

// Send status command to status server
func SendStats(cmdCode int) (replyChan chan string) {
	var cmd statsCmd
	// only report command returns result
	// others commands have no result, so they don't need reply channel
	switch cmdCode {
	case StatsCmdReport, StatsCmdIsShutdown:
		replyChan = make(chan string)
	default:
		replyChan = nil
	}
	cmd = statsCmd{cmdCode, replyChan}
	statsChannel <- cmd
	return
}

func StatsReport() string {
	replyChan := SendStats(StatsCmdReport)
	return <-replyChan
}

func IsShutdown() bool {
	replyChan := SendStats(StatsCmdIsShutdown)
	if isShutdown := <-replyChan; isShutdown == ShutdownYes {
		return true
	}
	return false
}
