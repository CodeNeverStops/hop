package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// A worker type use to handle task.
type Worker struct {
	task *Task

	// A message inbox to receive command messages.
	msgInbox inbox
}

// A worker run.
func (w *Worker) Run() {
	// Add a new flag into pool.
	// It will blocked if the pool is full.
	// We can specify the pool size when the server start.
	workerPool <- true
	go func() {
		SendStats(StatsCmdNewWorker)
		defer func() {
			workerHub.Unregister(w.msgInbox)
			SendStats(StatsCmdCloseWorker)
			// Free a flag from pool. So another worker can begin to run.
			<-workerPool
		}()
		w.postback()
	}()
}

// Do postback operation
func (w *Worker) postback() {
	var (
		times uint8 = 0
		ret   bool
	)
	// send a request
	ret = w.sendRequest(times)
	if ret {
		SendStats(StatsCmdSuccTask)
		return
	}
	// We will start a time ticker to send requests if the above request is failed.
	c := time.Tick(time.Duration(conf.RetryInterval) * time.Second)
	for {
		select {
		case cmd := <-w.msgInbox:
			if shutdown := w.handleCommand(cmd); shutdown {
				break
			}
		case <-c:
			// We will give it up if retry times beyond the max.
			if times >= conf.RetryTimes {
				url := w.task.Url()
				Logf(LogLevelWarning, "reach max times. throw it away. detail: url=%s,times=%d", url, times)
				SendStats(StatsCmdFailedTask)
				break
			}
			times++
			ret = w.sendRequest(times)
			if ret {
				SendStats(StatsCmdSuccTask)
				break
			}
		}
	}
}

// Send a http request
func (w *Worker) sendRequest(times uint8) bool {
	url := w.task.Url()
	resp, err := http.Get(url)
	if err != nil {
		Logf(LogLevelWarning, "failed to send request. detail: url=%s,times=%d", url, times)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logf(LogLevelWarning, "failed to read response. detail: url=%s,times=%d", url, times)
		return false
	}
	bodyStr := string(body)
	if strings.Index(bodyStr, "success=true;") > -1 {
		return true
	}
	Logf(LogLevelWarning, "failed to send request2. detail: response=%s,url=%s,times=%d", bodyStr, url, times)
	return false
}

func (w *Worker) handleCommand(cmd int) (shutdown bool) {
	switch cmd {
	case WorkerCmdShutdown:
		w.task.SaveTask()
		shutdown = true
	default:
		Logf(LogLevelWarning, "unknown command: %d", cmd)
		shutdown = false
	}
	return
}

// Create a new worker
func NewWorker(task *Task) *Worker {
	w := &Worker{
		task,
		make(inbox),
	}
	workerHub.Register(w.msgInbox)
	return w
}

// Create global pool of workers
func NewWorkerPool() chan bool {
	return make(chan bool, conf.WorkerPoolSize)
}
