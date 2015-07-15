package main

import (
	"testing"
)

func TestUptimeFormatOneSection(t *testing.T) {
	cases := map[uint32]string{
		86433:     "1 day",
		86400 * 2: "2 days",
		3601:      "1 hour",
		7202:      "2 hours",
		7262:      "2 hours",
		1:         "1 second",
		0:         "",
	}
	for k, v := range cases {
		ret := UptimeFormat(k, 1)
		if ret != v {
			t.Errorf("seconds:%d, excepted:%s, got:%s", k, v, ret)
		}
	}
}

func TestUptimeFormatTwoSections(t *testing.T) {
	cases := map[uint32]string{
		86433:     "1 day, 33 seconds",
		86400 * 2: "2 days",
		3601:      "1 hour, 1 second",
		7202:      "2 hours, 2 seconds",
		7262:      "2 hours, 1 minute",
		1:         "1 second",
		0:         "",
	}
	for k, v := range cases {
		ret := UptimeFormat(k, 2)
		if ret != v {
			t.Errorf("seconds:%d, excepted:%s, got:%s", k, v, ret)
		}
	}
}

func TestUptimeFormatThreeSections(t *testing.T) {
	cases := map[uint32]string{
		86433:        "1 day, 33 seconds",
		86400 * 2:    "2 days",
		86400 + 3601: "1 day, 1 hour, 1 second",
		7202:         "2 hours, 2 seconds",
		7262:         "2 hours, 1 minute, 2 seconds",
		1:            "1 second",
		0:            "",
	}
	for k, v := range cases {
		ret := UptimeFormat(k, 3)
		if ret != v {
			t.Errorf("seconds:%d, excepted:%s, got:%s", k, v, ret)
		}
	}
}
