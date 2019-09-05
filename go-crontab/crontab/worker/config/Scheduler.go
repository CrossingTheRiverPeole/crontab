package config

import (
	"fmt"
	"go-crontab/crontab/common"
	"time"
)

var (
	G_scheduler *Scheduler
)
//
type Scheduler struct {
	jobEventChan         chan *JobEvent               //job事件
	jobPlanTable         map[string]*JobSchedulerPlan // 任务计划表
	jobExecutingTable    map[string]*JobExecuteInfo   //	任务执行表
	jobExecuteResultChan chan *JobExecuteResult       // 接收任务执行结果
}

//初始化scheduler
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan:         make(chan *JobEvent, 1000),
		jobPlanTable:         make(map[string]*JobSchedulerPlan),
		jobExecutingTable:    make(map[string]*JobExecuteInfo),   // 任务执行表
		jobExecuteResultChan: make(chan *JobExecuteResult, 1000), // 接收任务执行结果
	}
	// 启动调度协程
	go G_scheduler.schedulerLoop()
	return
}

/**
启动调度协程
 */
func (scheduler *Scheduler) schedulerLoop() {
	var (
		jobEvent         *JobEvent
		schedulerAfter   time.Duration
		schedulerTimer   *time.Timer
		jobExecuteResult *JobExecuteResult
	)
	//调用trySchedule方法，判断下一次任务调度时间间隔，
	schedulerAfter = scheduler.tryScheduler()

	// 调度的延时定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	// 持续获取推送过来的jobEvent
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C: //最近的任务到

		case jobExecuteResult = <-scheduler.jobExecuteResultChan:
			// 处理任务执行结果，
			scheduler.HandleRusult(jobExecuteResult)
		}

		// 准备进行任务调度
		schedulerAfter = scheduler.tryScheduler()
		//任务调度完成之后重新设置调度间隔
		schedulerTimer.Reset(schedulerAfter)
	}
}

/**
处理任务调度
 */
func (scheduler *Scheduler) tryScheduler() (schedulerAfter time.Duration) {
	var (
		jobPlan  *JobSchedulerPlan
		now      time.Time
		nearTime *time.Time
	)

	// 如果任务表为空的话，随便睡眠多久
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}

	now = time.Now()

	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// 开始执行任务
			scheduler.tryStartJob(jobPlan)
			// 重新计算任务的下一次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		//统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	//下次调度的时间间隔，
	schedulerAfter = (*nearTime).Sub(now)
	return
}

/**
开始执行调度任务
 */
func (scheduler *Scheduler) tryStartJob(jobPlan *JobSchedulerPlan) {
	var (
		jobExecuteInfo *JobExecuteInfo
		jobExecuting   bool
	)
	// 判断任务执行表中是否存在该任务，若任务存在，则不再执行
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		fmt.Println(jobPlan.Job.Name, "任务正在执行中，不在进行此次任务调度执行")
		return // 如果任务执行表中有任务则不再执行
	}

	// 构建任务执行时间
	jobExecuteInfo = BuildJobExecuteInfo(jobPlan)
	// 把任务执行信息存放于map中
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	//执行任务
	G_executor.ExecuteJob(jobExecuteInfo)
}

/**
处理jobEvent解析成jobPlane
 */
func (scheduler *Scheduler) handleJobEvent(jobEvent *JobEvent) {
	var (
		err              error
		jobExist         bool
		jobSchedulerPlan *JobSchedulerPlan
		jobExecutingInfo *JobExecuteInfo
		jobExecuting     bool
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulerPlan, err = BuildJobSchedulerPlan(jobEvent.Job); err != nil {
			fmt.Println("构建jobSchedulePlan出现错误", err)
			return
		}
		// 把任务添加到jobPlanTable中
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulerPlan, jobExist = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExist {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVNET_KILL: // 强制杀死任务
		//判断jobExecutingTable中是否存在这个任务
		fmt.Println("任务列表中存在该任务", scheduler.jobExecutingTable[jobEvent.Job.Name])
		if jobExecutingInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			// 取消执行的任务
			fmt.Println("强杀任务执行")
			jobExecutingInfo.cancelFunc()
		}
	}
}

// 推送任务事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *JobEvent) {
	scheduler.jobEventChan <- jobEvent // 推送jobEvent到scheduler
}

//处理任务执行结果
func (scheduler *Scheduler) HandleRusult(jobExecuteResult *JobExecuteResult) {
	var (
		jobLog *JobLog
	)
	// 删除任务执行状态
	fmt.Println("任务执行结果", string(jobExecuteResult.Output))

	//处理执行结果并删除jobExecuteInfoTable中的数据
	delete(G_scheduler.jobExecutingTable, jobExecuteResult.ExecuteInfo.Job.Name)
	//生成日志结构体
	jobLog = &JobLog{
		JobName:      jobExecuteResult.ExecuteInfo.Job.Name,
		Command:      jobExecuteResult.ExecuteInfo.Job.Command,
		OutPut:       string(jobExecuteResult.Output),
		PlanTime:     jobExecuteResult.ExecuteInfo.PlanTime.Unix(),
		ScheduleTime: jobExecuteResult.ExecuteInfo.RealTime.Unix(),
		StartTime:    jobExecuteResult.StartTime.Unix(),
		EndTime:      jobExecuteResult.EndTime.Unix(),
	}
	if jobExecuteResult.Err != nil {
		jobLog.Err = jobExecuteResult.Err.Error()
	}

	// 发送到G_log channel进行日志的存储
	G_logSink.appendLog(jobLog)
}
