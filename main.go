package main

import (
	"encoding/json"
	"gopkg.in/redis.v3"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	redisKey   string = "hoff:postback:v1"
	convTpl    string = "http://arcgames.go2cloud.org/aff_lsr?transaction_id=[trans_id]&adv_sub=[user_id]"
	goalTpl    string = "http://arcgames.go2cloud.org/aff_goal?a=lsr&transaction_id=[trans_id]&goal_id=[goal_id]"
	retryTimes int    = 5
)

type Poster interface {
	Url() string
}

type Conversion struct {
	TransId string `json:"trans_id"`
	UserId  string `json:"uid"`
}

func (this *Conversion) Url() string {
	s := strings.Replace(convTpl, "[trans_id]", this.TransId, 1)
	s = strings.Replace(s, "[user_id]", this.UserId, 1)
	return s
}

type Goal struct {
	TransId string `json:"trans_id"`
	GoalId  string `json:"goal_id"`
}

func (this *Goal) Url() string {
	s := strings.Replace(goalTpl, "[trans_id]", this.TransId, 1)
	s = strings.Replace(s, "[goal_id]", this.GoalId, 1)
	return s
}

var (
	data   []string
	err    error
	poster Poster
)

func main() {
	client := NewClient()
	pool := make(chan bool, 10)
	for {
		pool <- true
		if data, err = client.BLPop(0, redisKey).Result(); err != nil {
			continue
		}
		tmpData := data[1]
		tmpByte := []byte(tmpData)
		if strings.Index(tmpData, "\"goal_id\":") > -1 {
			goal := &Goal{}
			if decodeErr := json.Unmarshal(tmpByte, &goal); decodeErr != nil {
				continue
			}
			poster = goal
		} else {
			conversion := &Conversion{}
			if decodeErr := json.Unmarshal(tmpByte, &conversion); decodeErr != nil {
				continue
			}
			poster = conversion
		}
		Log(LevelInfo, poster)
		go func() {
			defer func() { <-pool }()
			url := poster.Url()
			postback(url)
		}()
	}
}

//func NewClient(host string, port uint16, password string, db uint8) *redis.Client {
func NewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	Log(LevelInfo, pong, err)
	return client
}

func postback(url string) {
	times := 0
	ret := sendRequest(url, times)
	if ret {
		return
	}
	c := time.Tick(5 * time.Second)
	//c := time.Tick(10 * time.Minute)
	for now := range c {
		if times >= retryTimes {
			break
		}
		times++
		Logf(LevelInfo, "%s,%d", now, times)
		sendRequest(url, times)
	}
}

func sendRequest(url string, times int) bool {
	Logf(LevelInfo, "url:%s,times:%d", url, times)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	bodyStr := string(body)
	Log(LevelInfo, bodyStr)
	if strings.Index(bodyStr, "success=true;") > -1 {
		return true
	}
	return false
}
