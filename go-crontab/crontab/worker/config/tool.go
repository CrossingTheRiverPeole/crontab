package config

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"go-crontab/crontab/common"
	"strings"
	"time"
)

/**
反序列化job
 */
func UnpackJob(value []byte) (ret *common.Job, err error) {
	var (
		job *common.Job
	)

	job = &common.Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

/**
构建jobEvent
 */
func BuildJobEvent(eventType int, job *common.Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

/**
获取jobName
 */
func ExtractJobName(key string) (jobName string) {
	return strings.TrimPrefix(key, common.JOB_SAVE_DIR)
}

/**
获取强杀任务的名称
 */
func ExtracrKillerName(key string) (jobName string) {
	jobName = strings.TrimPrefix(key, common.JOB_KILLER_DIR)
	return

}

/**
构建jobSchedulerPlan
 */
func BuildJobSchedulerPlan(job *common.Job) (jobSchedulerPlan *JobSchedulerPlan, err error) {

	var (
		expr *cronexpr.Expression
	)
	//解析表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}
	//构建任务调度计划
	jobSchedulerPlan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

/**
构建任务执行信息
 */
func BuildJobExecuteInfo(jobPlan *JobSchedulerPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobPlan.Job,
		PlanTime: jobPlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.ctx, jobExecuteInfo.cancelFunc = context.WithCancel(context.TODO())
	return
}
