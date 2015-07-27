package main

import (
	"errors"
	"gopkg.in/redis.v3"
	"strconv"
	"time"
)

// Task queue used to fetch task from redis list
type TaskQueue struct {
	client *redis.Client
}

// Defined some errors
var (
	errLPopFailed  = errors.New("task queue: client LPop failed.")
	errLPushFailed = errors.New("task queue: client LPush failed.")
)

// Create a task queue
func NewTaskQueue() *TaskQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.RedisHost + ":" + strconv.FormatUint(uint64(conf.RedisPort), 10),
		Password: conf.RedisPWD,
		DB:       conf.RedisDB,
	})
	// exit with error if can't connect to redis server
	if _, err := client.Ping().Result(); err != nil {
		panic("failed to connect redis server")
	}
	return &TaskQueue{client}
}

func (queue *TaskQueue) LPop() (string, error) {
	data, err := queue.client.BLPop(5*time.Second, conf.RedisKey).Result()
	if err != nil {
		Log(LogLevelWarning, errLPopFailed.Error()+"detail:"+err.Error())
		return "", errLPopFailed
	}
	tmpData := data[1]
	return tmpData, nil
}

func (queue *TaskQueue) LPush(data string) error {
	if _, err := queue.client.LPush(conf.RedisKey, data).Result(); err != nil {
		Log(LogLevelWarning, errLPushFailed.Error()+"detail:"+err.Error())
		return errLPushFailed
	}
	return nil
}
