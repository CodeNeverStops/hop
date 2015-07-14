package main

import (
	"fmt"
	"strings"
)

func UptimeFormat(secs uint32) string {
	timeUnits := map[uint32][2]string{
		1:     [2]string{"second", "seconds"},
		60:    [2]string{"minute", "minutes"},
		3600:  [2]string{"hour", "hours"},
		86400: [2]string{"day", "days"},
	}
	timeSeq := [4]uint32{86400, 3600, 60, 1}
	result := make([]string, len(timeSeq))

	i := 0
	for _, index := range timeSeq {
		if v, prs := timeUnits[index]; prs {
			if secs >= index {
				num := secs / index
				secs = secs % index
				unit := v[0]
				if num > 1 {
					unit = v[1]
				}
				result[i] = fmt.Sprintf("%d%s", num, unit)
				i++
			}
		}
	}
	return strings.Join(result[:i], " ")
}
