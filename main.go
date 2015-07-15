package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/redis.v3"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "0.1.0"
)

type Poster interface {
	Url() string
}

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

type Flags struct {
	RedisHost      string
	RedisPort      uint16
	RedisDB        int64
	RedisPWD       string
	RedisKey       string
	RetryTimes     uint8
	RetryInterval  uint16
	WorkerPoolSize uint16
	LogQueueSize   uint32
	LogBufferSize  uint16
	AdminPort      uint16
}

var (
	err    error
	poster Poster
	conf   *Flags
)

func main() {
	parseFlags()
	logStart()
	statsStart()
	adminStart()

	var data []string
	client := NewClient()
	pool := make(chan bool, conf.WorkerPoolSize)

	for {
		pool <- true
		if data, err = client.BLPop(0, conf.RedisKey).Result(); err != nil {
			Log("redis blpop failed")
			continue
		}
		tmpData := data[1]
		tmpByte := []byte(tmpData)
		if strings.Index(tmpData, "\"goal_id\":") > -1 {
			goal := &Goal{}
			if decodeErr := json.Unmarshal(tmpByte, &goal); decodeErr != nil {
				Log("decode goal failed")
				continue
			}
			poster = goal
		} else {
			conversion := &Conversion{}
			if decodeErr := json.Unmarshal(tmpByte, &conversion); decodeErr != nil {
				Log("decode conversion failed")
				continue
			}
			poster = conversion
		}
		go func() {
			SendStats(StatsCmdNewWorker)
			defer func() {
				<-pool
				SendStats(StatsCmdCloseWorker)
			}()
			url := poster.Url()
			postback(url)
		}()
	}
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

func parseFlags() {
	redisHost := flag.String("redis-host", "127.0.0.1", "redis host")
	redisPort := flag.Uint("redis-port", 6379, "redis port")
	redisDB := flag.Uint("redis-db", 0, "redis db")
	redisPWD := flag.String("redis-pwd", "", "redis password")
	redisKey := flag.String("redis-key", "", "redis key")
	retryTimes := flag.Uint("t", 5, "retry times")
	retryInterval := flag.Uint("i", 10, "retry interval, unit: Second")
	workerPoolSize := flag.Uint("n", 1000, "max pool size of workers")
	logQueueSize := flag.Uint("log-queue", 1000, "log queue size")
	logBufferSize := flag.Uint("log-buffer", 2, "log buffer size")
	adminPort := flag.Uint("admin-port", 8888, "admin port")

	flag.Parse()

	if *redisKey == "" {
		fmt.Println("please specify the redis key")
		os.Exit(1)
	}

	conf = &Flags{
		RedisHost:      *redisHost,
		RedisPort:      uint16(*redisPort),
		RedisDB:        int64(*redisDB),
		RedisPWD:       *redisPWD,
		RedisKey:       *redisKey,
		RetryTimes:     uint8(*retryTimes),
		RetryInterval:  uint16(*retryInterval),
		WorkerPoolSize: uint16(*workerPoolSize),
		LogQueueSize:   uint32(*logQueueSize),
		LogBufferSize:  uint16(*logBufferSize),
		AdminPort:      uint16(*adminPort),
	}
	fmt.Println(conf)
}

func postback(url string) {
	var (
		times uint8 = 0
		ret   bool
	)
	ret = sendRequest(url, times)
	if ret {
		SendStats(StatsCmdSuccTask)
		return
	}
	c := time.Tick(time.Duration(conf.RetryInterval) * time.Second)
	for range c {
		if times >= conf.RetryTimes {
			Logf(LogLevelWarning, "reach max times. throw it away. detail: url=%s,times=%d", url, times)
			SendStats(StatsCmdFailedTask)
			break
		}
		times++
		ret = sendRequest(url, times)
		if ret {
			SendStats(StatsCmdSuccTask)
			break
		}
	}
}

func sendRequest(url string, times uint8) bool {
	url = ""
	Logf(LogLevelInfo, "url:%s,times:%d", url, times)
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
