package main

import (
	"testing"
)

func TestTask(t *testing.T) {
	conf = &Flags{
		RedisHost: "127.0.0.1",
		RedisPort: 6379,
		RedisDB:   0,
		RedisPWD:  "",
		RedisKey:  "hoff:test:task",
	}
	testCases := map[string]string{
		`{"trans_id":"1234560000","uid":"1000"}`:                          `{"trans_id":"1234560000","uid":"1000"}`,
		`{"trans_id":"abcdedfaaf","uid":"999999","create_time":10000000}`: `{"trans_id":"abcdedfaaf","uid":"999999"}`,
		`{"trans_id":"zxcv123123","goal_id":"80"}`:                        `{"trans_id":"zxcv123123","goal_id":"80"}`,
		`{"trans_id":"xyzjfklajf","goal_id":"64","create_time":10000000}`: `{"trans_id":"xyzjfklajf","goal_id":"64"}`,
	}
	taskQueue = NewTaskQueue()
	for k, v := range testCases {
		taskQueue.LPush(k)
		task, _ := NewTask()
		task.SaveTask()
		if ret, _ := taskQueue.LPop(); v != ret {
			t.Errorf("excepted:%s, got:%s", v, ret)
		}
	}
}
