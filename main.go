package main

import (
	dbs "autolike/databases"
	"fmt"
	"runtime"
	"time"
)

func main() {
	cpuNume := runtime.NumCPU() //获取cpu核数
	runtime.GOMAXPROCS(cpuNume - 1)
	var Current int
	//ticker := time.NewTicker(time.Minute * time.Duration(1))
	ticker := time.NewTicker(time.Second * time.Duration(10))
	for range ticker.C {
		nr := dbs.NewTimeRecord{}
		if row, err := dbs.DB.Query("select id,user_id,time_cat_id from new_timer_record where like_num < 30 and id > ? limit 20", Current); err == nil {
			for row.Next() {
				row.Scan(&nr.Id, &nr.UserId, &nr.TimeCatId)
				fmt.Println(nr.Id)
				fmt.Printf("当前id %d", nr.Id)
				nr.RootTask()
			}
			Current = nr.Id
			fmt.Printf("Current %d", Current)
		} else {
			fmt.Println("main", err)
		}

	}
}
