package config

import (
	"os/exec"
	"time"
)

var (
	G_executor Executor
)

type Executor struct {
}

/**
执行任务
 */
func (executor Executor) ExecuteJob(info *JobExecuteInfo) {

	var (
		cmd              *exec.Cmd
		outPut           []byte
		err              error
		jobExecuteResult *JobExecuteResult
		jobLock          *JobLock
	)

	//初始化分布式锁
	jobLock = G_jobMgr.CreateJobLock(info.Job.Name)
	err = jobLock.TryLock()
	// 释放锁
	defer jobLock.Unlock()

	//构建任务执行结果
	jobExecuteResult = &JobExecuteResult{
		ExecuteInfo: info,
		StartTime:   time.Now(),
	}
	if err != nil {
		jobExecuteResult.Err = err
		jobExecuteResult.EndTime = time.Now()
		jobExecuteResult.Output = []byte("抢锁失败")
	} else {
		cmd = exec.CommandContext(info.ctx, "C:/cygwin64/bin/bash.exe", "-c", info.Job.Command)
		// 任务结束时间
		jobExecuteResult.EndTime = time.Now()
		// 捕获并输出
		outPut, err = cmd.CombinedOutput()
		// 任务输出
		jobExecuteResult.Output = outPut
		jobExecuteResult.Err = err
	}
	// 把结果发送到管道，由scheduler来处理
	G_scheduler.jobExecuteResultChan <- jobExecuteResult
}
