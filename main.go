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
	RedisHost     string
	RedisPort     uint16
	RedisDB       int64
	RedisPWD      string
	RedisKey      string
	RetryTimes    uint8
	RetryInterval uint16
	PoolSize      uint16
	LogQueueSize  uint32
	LogBufferSize uint16
}

var (
	err     error
	poster  Poster
	flags   *Flags
	logChan chan bool
)

func main() {
	parseFlags()

	logChan = make(chan bool)
	logStart()
	<-logChan

	adminStart()

	var data []string
	client := NewClient()
	pool := make(chan bool, flags.PoolSize)

	for {
		pool <- true
		if data, err = client.BLPop(0, flags.RedisKey).Result(); err != nil {
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
			defer func() { <-pool }()
			url := poster.Url()
			postback(url)
		}()
	}
}

func NewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     flags.RedisHost + ":" + strconv.FormatUint(uint64(flags.RedisPort), 10),
		Password: flags.RedisPWD,
		DB:       flags.RedisDB,
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
	poolSize := flag.Uint("n", 1000, "max pool size")
	logQueueSize := flag.Uint("log-queue", 1000, "log queue size")
	logBufferSize := flag.Uint("log-buffer", 2, "log buffer size")

	flag.Parse()

	if *redisKey == "" {
		fmt.Println("please specify the redis key")
		os.Exit(1)
	}

	flags = &Flags{
		RedisHost:     *redisHost,
		RedisPort:     uint16(*redisPort),
		RedisDB:       int64(*redisDB),
		RedisPWD:      *redisPWD,
		RedisKey:      *redisKey,
		RetryTimes:    uint8(*retryTimes),
		RetryInterval: uint16(*retryInterval),
		PoolSize:      uint16(*poolSize),
		LogQueueSize:  uint32(*logQueueSize),
		LogBufferSize: uint16(*logBufferSize),
	}
	fmt.Println(flags)
}

func postback(url string) {
	var (
		times uint8 = 0
		ret   bool
	)
	ret = sendRequest(url, times)
	if ret {
		return
	}
	c := time.Tick(time.Duration(flags.RetryInterval) * time.Second)
	for range c {
		if times >= flags.RetryTimes {
			Logf(LevelWarning, "reach max times. throw it away. detail: url=%s,times=%d", url, times)
			break
		}
		times++
		ret = sendRequest(url, times)
		if ret {
			break
		}
	}
}

func sendRequest(url string, times uint8) bool {
	Logf(LevelInfo, "url:%s,times:%d", url, times)
	resp, err := http.Get(url)
	if err != nil {
		Logf(LevelWarning, "failed to send request. detail: url=%s,times=%d", url, times)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logf(LevelWarning, "failed to read response. detail: url=%s,times=%d", url, times)
		return false
	}
	bodyStr := string(body)
	if strings.Index(bodyStr, "success=true;") > -1 {
		return true
	}
	Logf(LevelWarning, "failed to send request2. detail: response=%s,url=%s,times=%d", bodyStr, url, times)
	return false
}
