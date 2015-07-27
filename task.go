package main

import (
	"encoding/json"
	"errors"
	"strings"
)

// A poster interface, every poster type should implement it.
type Poster interface {
	// Return the template of postback url
	Tpl() string

	// Return the real url of postback
	Url() string

	// Convert type poster to json format,
	// Save the json string into redis list once the server shutdown.
	Json() (string, error)
}

// A poster type: Conversion
type Conversion struct {
	TransId string `json:"trans_id"`
	UserId  string `json:"uid"`
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
		Logf(LogLevelWarning, errEncodeConversionFailed.Error()+" trans_id=%s,uid=%s: ", this.TransId, this.UserId)
		return "", errEncodeConversionFailed
	}
	return string(ret), nil
}

// A poster type: Goal
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
		Logf(LogLevelWarning, errEncodeGoalFailed.Error()+" trans_id=%s,goal_id=%s %s", this.GoalId, this.TransId)
		return "", errEncodeGoalFailed
	}
	return string(ret), nil
}

type Task struct {
	TaskPoster Poster
	TaskType   uint8
}

// The task has two types
const (
	TaskTypeConversion = 1
	TaskTypeGoal       = 2
)

// Defined some errors
var (
	errEncodeGoalFailed       = errors.New("task: encode goal failed.")
	errDecodeGoalFailed       = errors.New("task: decode goal failed.")
	errEncodeConversionFailed = errors.New("task: encode conversion failed.")
	errDecodeConversionFailed = errors.New("task: decode conversion failed.")
)

// Create a new task
// Fetch data from task queue
func NewTask() (*Task, error) {
	var (
		poster   Poster
		taskType uint8
	)
	Log(LogLevelInfo, "get task start")
	data, err := taskQueue.LPop()
	Log(LogLevelInfo, "get task done")
	if err != nil {
		return nil, err
	}
	tmpByte := []byte(data)
	if strings.Index(data, "\"goal_id\":") > -1 {
		goal := &Goal{}
		if decodeErr := json.Unmarshal(tmpByte, &goal); decodeErr != nil {
			Logf(LogLevelWarning, errDecodeGoalFailed.Error()+" data: %s", data)
			return nil, errDecodeGoalFailed
		}
		poster = goal
		taskType = TaskTypeGoal
	} else {
		conversion := &Conversion{}
		if decodeErr := json.Unmarshal(tmpByte, &conversion); decodeErr != nil {
			Logf(LogLevelWarning, errDecodeConversionFailed.Error()+" data:%s", data)
			return nil, errDecodeConversionFailed
		}
		poster = conversion
		taskType = TaskTypeConversion
	}
	return &Task{poster, taskType}, nil
}

// Save task into task queue once once the server shutdown
func (t *Task) SaveTask() error {
	data, err := t.TaskPoster.Json()
	if err != nil {
		return err
	}
	err = taskQueue.LPush(data)
	return err
}

// Get the postback url of the task
func (t *Task) Url() string {
	return t.TaskPoster.Url()
}
