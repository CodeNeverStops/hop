package main

import (
	"encoding/json"
	"gopkg.in/redis.v3"
	"strconv"
	"strings"
)

type Conversion struct {
	TransId string `json:"trans_id"`
	UserId  string `json:"uid"`
}

func (this *Conversion) Url() string {
	const convTpl string = "http://arcgames.go2cloud.org/aff_lsr?transaction_id=[trans_id]&adv_sub=[user_id]"
	s := strings.Replace(convTpl, "[trans_id]", this.TransId, 1)
	s = strings.Replace(s, "[user_id]", this.UserId, 1)
	return s
}

type Goal struct {
	TransId string `json:"trans_id"`
	GoalId  string `json:"goal_id"`
}

func (this *Goal) Url() string {
	const goalTpl string = "http://arcgames.go2cloud.org/aff_goal?a=lsr&transaction_id=[trans_id]&goal_id=[goal_id]"
	s := strings.Replace(goalTpl, "[trans_id]", this.TransId, 1)
	s = strings.Replace(s, "[goal_id]", this.GoalId, 1)
	return s
}

type Task struct {
	url string
}

func NewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.RedisHost + ":" + strconv.FormatUint(uint64(conf.RedisPort), 10),
		Password: conf.RedisPWD,
		DB:       conf.RedisDB,
	})
	if _, err := client.Ping().Result(); err != nil {
		panic("failed to connect redis server")
	}
	return client
}

func (t *Task) fetchTask() (bool, string) {
	var data []string
	var url string
	if data, err = client.BLPop(0, conf.RedisKey).Result(); err != nil {
		Log("redis blpop failed")
		return false, ""
	}
	tmpData := data[1]
	tmpByte := []byte(tmpData)
	if strings.Index(tmpData, "\"goal_id\":") > -1 {
		goal := &Goal{}
		if decodeErr := json.Unmarshal(tmpByte, &goal); decodeErr != nil {
			Log("decode goal failed")
			return false, ""
		}
		url = goal.Url()
	} else {
		conversion := &Conversion{}
		if decodeErr := json.Unmarshal(tmpByte, &conversion); decodeErr != nil {
			Log("decode conversion failed")
			return false, ""
		}
		url = conversion.Url()
	}
	return true, url
}

func (t *Task) saveTask() {

}
