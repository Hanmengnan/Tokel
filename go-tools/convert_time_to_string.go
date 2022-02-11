package gotools

import (
	"fmt"
	"time"
)

/*
	将Unix时间戳解析为字符串
	eg. 1639014717 -> "2021-12-9 9:51:57"
*/

func ParseTime(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	return fmt.Sprintf("%d-%d-%d %d:%d:%d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
}
