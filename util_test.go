package main

import (
	"testing"
)

func TestUptimeFormat(t *testing.T) {
	cases := map[uint32]string{
		86433:     "1day 33seconds",
		86400 * 2: "2days",
		3601:      "1hour 1second",
		7202:      "2hours 2seconds",
		7262:      "2hours 1minute 2seconds",
		1:         "1second",
		0:         "",
	}
	for k, v := range cases {
		ret := UptimeFormat(k)
		if ret != v {
			t.Errorf("seconds:%d, excepted:%s, got:%s", k, v, ret)
		}
	}
}
