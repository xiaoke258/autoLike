package tools

import (
	"fmt"
	"math/rand"
	"time"
)

func Random(min, max int) int {
	if min == max {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func TotalRandom() int {
	n := Random(20, 120)
	fmt.Printf("预计点赞总量 %d\n", n)
	return n
}

func TimesRandom() (n int) {
	n = Random(2, 5)
	fmt.Printf("预计点赞次数 %d\n", n)
	return
}

func CountRandom(total, times int) (n int) {
	n = Random(0, total/times)
	fmt.Printf("本次实际点赞数量 %d\n", n)
	return
}

func TotalTime() (n int) {
	n = Random(0, 24*60*60)
	fmt.Printf("预计总耗时 %d\n", n)
	return
}

func AfterTime(start, total int) (n int) {
	n = Random(start, total)
	fmt.Printf("本次延迟时间 %d\n", n)
	return
}
