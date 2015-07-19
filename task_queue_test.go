package main

import (
	"testing"
)

func TestTaskQueue(t *testing.T) {
	conf = &Flags{
		RedisHost: "127.0.0.1",
		RedisPort: 6379,
		RedisDB:   0,
		RedisPWD:  "",
		RedisKey:  "hoff:test:queue",
	}
	testCases := []string{
		"http://www.arcgames.com",
		`["abc", 123, 100.00]`,
		`{"num":5.0,"strs":["a","b"]}`,
	}
	queue := NewTaskQueue()
	for k, v := range testCases {
		queue.LPush(v)
		if ret, _ := queue.LPop(); v != ret {
			t.Errorf("index:%d, excepted:%s, got:%s", k, v, ret)
		}
	}
}
