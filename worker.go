package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Worker struct {
	url   string
	inbox conn
}

func (w *Worker) Run() {
	pool <- true
	go func() {
		SendStats(StatsCmdNewWorker)
		defer func() {
			workerHub.Unregister(w.inbox)
			SendStats(StatsCmdCloseWorker)
			<-pool
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
			w.handleCommand(cmd)
		case <-c:
			if times >= conf.RetryTimes {
				Logf(LogLevelWarning, "reach max times. throw it away. detail: url=%s,times=%d", w.url, times)
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
	resp, err := http.Get(w.url)
	if err != nil {
		Logf(LogLevelWarning, "failed to send request. detail: url=%s,times=%d", w.url, times)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logf(LogLevelWarning, "failed to read response. detail: url=%s,times=%d", w.url, times)
		return false
	}
	bodyStr := string(body)
	if strings.Index(bodyStr, "success=true;") > -1 {
		return true
	}
	Logf(LogLevelWarning, "failed to send request2. detail: response=%s,url=%s,times=%d", bodyStr, w.url, times)
	return false
}

func (w *Worker) handleCommand(cmd int) {
	switch cmd {
	case WorkerCmdShutdown:

	}
}

func NewWorker(url string) *Worker {
	w := &Worker{
		url:   url,
		inbox: make(conn),
	}
	workerHub.Register(w.inbox)
	return w
}
