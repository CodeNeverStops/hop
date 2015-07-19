package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Worker struct {
	task  *Task
	inbox conn
}

func (w *Worker) Run() {
	workerPool <- true
	go func() {
		SendStats(StatsCmdNewWorker)
		defer func() {
			workerHub.Unregister(w.inbox)
			SendStats(StatsCmdCloseWorker)
			<-workerPool
		}()
		w.postback()
	}()
}

func (w *Worker) postback() {
	var (
		times uint8 = 0
		ret   bool
	)
	ret = w.sendRequest(times)
	if ret {
		SendStats(StatsCmdSuccTask)
		return
	}
	c := time.Tick(time.Duration(conf.RetryInterval) * time.Second)
	for {
		select {
		case cmd := <-w.inbox:
			if shutdown := w.handleCommand(cmd); shutdown {
				break
			}
		case <-c:
			if times >= conf.RetryTimes {
				url := w.task.TaskPoster.Url()
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

func (w *Worker) sendRequest(times uint8) bool {
	url := w.task.TaskPoster.Url()
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

func NewWorker(task *Task) *Worker {
	w := &Worker{
		task,
		make(conn),
	}
	workerHub.Register(w.inbox)
	return w
}
