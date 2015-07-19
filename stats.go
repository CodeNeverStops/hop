package main

import (
	"errors"
	"fmt"
	"time"
)

type ServerStats struct {
	TaskTotal  uint64
	TaskSucc   uint64
	TaskFail   uint64
	WorkerCurr uint16
	WorkerMax  uint16
	Version    string
	StartTime  time.Time
}

const (
	StatsCmdSuccTask = iota
	StatsCmdFailedTask
	StatsCmdNewWorker
	StatsCmdCloseWorker
	StatsCmdReport
)

func (stats *ServerStats) HandleCommand(cmd statsCmd) error {
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
		if stats.WorkerCurr == 0 && isShutdown {
			shutdownChan <- true
		}
	case StatsCmdReport:
		if cmd.replyChan != nil {
			cmd.replyChan <- stats.Report()
		}
	default:
		Log(LogLevelWarning, "[stat server] command not found. command: %s", cmd)
		return errors.New("command not found")
	}
	return nil
}

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
	if isShutdown {
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

var statsChannel chan statsCmd
var stats *ServerStats

type statsCmd struct {
	cmd       int
	replyChan chan string
}

func statsStart() {
	// init server status
	stats = &ServerStats{
		Version:   Version,
		StartTime: time.Now(),
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

func SendStats(cmdCode int) (replyChan chan string) {
	var cmd statsCmd
	if cmdCode == StatsCmdReport {
		replyChan = make(chan string)
	} else {
		replyChan = nil
	}
	cmd = statsCmd{cmdCode, replyChan}
	go func() {
		statsChannel <- cmd
	}()
	return
}

func StatsReport() string {
	replyChan := SendStats(StatsCmdReport)
	return <-replyChan
}
