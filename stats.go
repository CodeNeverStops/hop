package main

import (
	"errors"
	"fmt"
	"time"
)

type ServerStats struct {
	TaskTotal uint64
	TaskSucc  uint64
	TaskFail  uint64
	//TaskSuccReport map[uint8]uint64
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

func (stats *ServerStats) HandleCommand(cmd int) error {
	switch cmd {
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
	default:
		Log(LevelWarning, "[stat server] command not found. command: %s", cmd)
		return errors.New("command not found")
	}
	return nil
}

func (stats *ServerStats) Report() string {
	uptime := UptimeFormat(uint32(time.Now().Sub(stats.StartTime) / time.Second))
	var (
		taskSuccRatio   float32 = 0
		taskFailRatio   float32 = 0
		workerCurrRatio float32 = 0
		workerMaxRatio  float32 = 0
	)
	if stats.TaskTotal > 0 {
		taskSuccRatio = float32(stats.TaskSucc / stats.TaskTotal * 100)
		taskFailRatio = float32(stats.TaskFail / stats.TaskTotal * 100)
	}

	if conf.WorkerPoolSize > 0 {
		workerCurrRatio = float32(stats.WorkerCurr / conf.WorkerPoolSize * 100)
		workerMaxRatio = float32(stats.WorkerMax / conf.WorkerPoolSize * 100)
	}

	return fmt.Sprintf(`
Version: %s 
Uptime: %s
Copyright (c) 2015 PerfectWorld
===============================
Task Total:     %d
  Task Success:   %d (%0.2f%%)
  Task Failed:    %d (%0.2f%%)
Worker Config:  %d
  Worker Current: %d (%0.2f%%)
  Worker Max:     %d (%0.2f%%)
	`,
		Version,
		uptime,
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

var statsChannel chan int
var stats *ServerStats

func statsStart() {
	stats = &ServerStats{
		Version:   Version,
		StartTime: time.Now(),
	}
	statsChannel = make(chan int, 10)
	go func(stats *ServerStats) {
		for {
			select {
			case cmd := <-statsChannel:
				stats.HandleCommand(cmd)
			}
		}
	}(stats)
}

func SendStats(cmd int) {
	go func() {
		statsChannel <- cmd
	}()
}

func StatsReport() string {
	return stats.Report()
}
