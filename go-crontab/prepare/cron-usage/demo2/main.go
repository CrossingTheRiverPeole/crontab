package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type CronJob struct {
	expr     *cronexpr.Expression
	nextTime time.Time
}

func main() {
	var (
		expr          *cronexpr.Expression
		err           error
		now           time.Time
		scheduleTable map[string]*CronJob
		cronJob       *CronJob
	)

	scheduleTable = make(map[string]*CronJob)

	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
		return
	}
	// 当前时间
	now = time.Now()
	//定义两个cronjob
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job1"] = cronJob

	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
		return
	}
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job2"] = cronJob

	go func() {
		var (
			jobName string
			cronJob *CronJob
			now     time.Time
		)
		for {
			//定时检查一下任务调度表
			now = time.Now()
			for jobName, cronJob = range scheduleTable {
				if cronJob.nextTime.Before(now) || cronJob.nextTime.Equal(now) {
					go func(jobName string) {
						fmt.Println("执行:", jobName)
					}(jobName)

					// 计算下次执行的时间
					cronJob.nextTime = cronJob.expr.Next(now)
					fmt.Println(jobName, "下次执行的时间：", cronJob.nextTime)
				}
			}

			//睡眠100毫秒
			select {
			case <-time.NewTimer(100 * time.Millisecond).C:
			}
		}
	}()
	time.Sleep(100 * time.Second)
}
