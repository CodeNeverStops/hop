package main

import (
	"encoding/json"
	"errors"
	"strings"
)

type Conversion struct {
	TransId string `json:"trans_id"`
	UserId  string `json:"uid"`
}

type Poster interface {
	Tpl() string
	Url() string
	Json() (string, error)
}

func (this *Conversion) Tpl() string {
	return "http://arcgames.go2cloud.org/aff_lsr?transaction_id=[trans_id]&adv_sub=[user_id]"
}

func (this *Conversion) Url() string {
	tpl := this.Tpl()
	s := strings.Replace(tpl, "[trans_id]", this.TransId, 1)
	s = strings.Replace(s, "[user_id]", this.UserId, 1)
	return s
}

func (this *Conversion) Json() (string, error) {
	ret, err := json.Marshal(*this)
	if err != nil {
		Logf("encode conversion failed. detail: %s", err.Error())
		return "", errEncodeConversionFailed
	}
	return string(ret), nil
}

type Goal struct {
	TransId string `json:"trans_id"`
	GoalId  string `json:"goal_id"`
}

func (this *Goal) Tpl() string {
	return "http://arcgames.go2cloud.org/aff_goal?a=lsr&transaction_id=[trans_id]&goal_id=[goal_id]"
}

func (this *Goal) Url() string {
	tpl := this.Tpl()
	s := strings.Replace(tpl, "[trans_id]", this.TransId, 1)
	s = strings.Replace(s, "[goal_id]", this.GoalId, 1)
	return s
}

func (this *Goal) Json() (string, error) {
	ret, err := json.Marshal(*this)
	if err != nil {
		Logf("encode goal failed. detail: %s", err.Error())
		return "", errEncodeGoalFailed
	}
	return string(ret), nil
}

func NewWorkerPool() chan bool {
	return make(chan bool, conf.WorkerPoolSize)
}

type Task struct {
	TaskPoster Poster
	TaskType   uint8
}

const (
	TaskTypeConversion = 1
	TaskTypeGoal       = 2
)

var (
	errEncodeGoalFailed       = errors.New("task: encode goal failed")
	errDecodeGoalFailed       = errors.New("task: decode goal failed")
	errEncodeConversionFailed = errors.New("task: encode conversion failed")
	errDecodeConversionFailed = errors.New("task: decode conversion failed")
)

func NewTask() (*Task, error) {
	var (
		poster   Poster
		taskType uint8
	)
	data, err := taskQueue.LPop()
	if err != nil {
		return nil, err
	}
	tmpByte := []byte(data)
	if strings.Index(data, "\"goal_id\":") > -1 {
		goal := &Goal{}
		if decodeErr := json.Unmarshal(tmpByte, &goal); decodeErr != nil {
			Log("decode goal failed")
			return nil, errDecodeGoalFailed
		}
		poster = goal
		taskType = TaskTypeGoal
	} else {
		conversion := &Conversion{}
		if decodeErr := json.Unmarshal(tmpByte, &conversion); decodeErr != nil {
			Log("decode conversion failed")
			return nil, errDecodeConversionFailed
		}
		poster = conversion
		taskType = TaskTypeConversion
	}
	return &Task{poster, taskType}, nil
}

func (t *Task) SaveTask() error {
	data, err := t.TaskPoster.Json()
	if err != nil {
		return err
	}
	err = taskQueue.LPush(data)
	return err
}
