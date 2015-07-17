package main

import (
	"flag"
	"fmt"
	"gopkg.in/redis.v3"
	"os"
)

const (
	Version = "0.1.0"
)

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
	conf   *Flags
	client *redis.Client
	pool   chan bool
)

func main() {
	parseFlags()
	logStart()
	statsStart()
	adminStart()
	workerHub.run()
	client = NewClient()
	pool = make(chan bool, conf.WorkerPoolSize)

	for {
		task := &Task{}
		succ, url := task.fetchTask()
		if !succ {
			continue
		}
		w := NewWorker(url)
		w.Run()
	}
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
