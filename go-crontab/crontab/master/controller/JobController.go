package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go-crontab/crontab/common"
	"go-crontab/crontab/master/service"
	"io/ioutil"
	"net/http"
)

/**
保存任务controller
 */
func SaveJob(c *gin.Context) {
	var (
		jobBytes []byte      //要保存的job 二进制
		err      error       // 错误信息
		job      *common.Job // job
		errno    int         // 错误编码
		oldJob   *common.Job //旧的job
	)

	// 读取传递过来的job
	if jobBytes, err = ioutil.ReadAll(c.Request.Body); err != nil {
		errno = -1
		oldJob = nil
		return
	}

	//对job记性反序列化
	job = &common.Job{}
	if err = json.Unmarshal(jobBytes, job); err != nil {
		errno = -1
		oldJob = nil
	}
	//反序列化成功之后保存job到
	oldJob, err = service.SaveJobService(job)

	// 返回结果
	if err != nil {
		c.JSON(http.StatusOK, common.BuildResponse(errno, err.Error(), oldJob))
		return
	}
	c.JSON(http.StatusOK, common.BuildResponse(errno, "", oldJob))
}

/**
获取任务列表:根据任务名称前缀获取所有的任务列表
 */
func HandleJobList(c *gin.Context) {
	var (
		errno int
		msg   string
		data  []*common.Job
		err   error
	)
	// 判断任务名称是
	/*if jobName = c.Query("name"); jobName == "" {
		errno = -1
		msg = "任务名称为空"
		return
	}*/

	// 获取所有任务列表
	if data, err = service.ListJobs(); err != nil {
		errno = -1
		msg = err.Error()
	}
	// 返回结果，并且defer在此处不好使
	c.JSON(http.StatusOK, common.BuildResponse(errno, msg, data))
}

/**
根据任务名称删除任务：
 */
func HandleJobRemove(c *gin.Context) {
	var (
		jobName string //job名称
		errno   = 1
		msg     string
		data    *common.Job //删除的job数据
		err     error
	)

	// 获取job名称
	if jobName = c.Query("name"); jobName == "" {
		errno = -1
		msg = "任务名称为空无法删除"
		c.JSON(http.StatusOK, common.BuildResponse(errno, msg, data))
	}
	//根据任务名称查询
	if data, err = service.JobRemove(jobName); err != nil {
		errno = -1
		msg = err.Error()
		c.JSON(http.StatusOK, common.BuildResponse(errno, msg, data))
	}
	//返回结果
	c.JSON(http.StatusOK, common.BuildResponse(errno, "success", data))
}

/**
强制杀死任务：
 */
func HandleJobKill(c *gin.Context) {
	var (
		jobName string
		err     error
		errno   = 1
		msg     string
	)
	if jobName = c.Query("name"); jobName == "" {
		errno = -1
		msg = "任务名称不能为空"
		c.JSON(http.StatusOK, common.BuildResponse(errno, msg, nil))
		return
	}

	//获取结果
	if err = service.JobKill(jobName); err != nil {
		errno = -1
		msg = err.Error()
		c.JSON(http.StatusOK, common.BuildResponse(errno, msg, nil))
		return
	}
	c.JSON(http.StatusOK, common.BuildResponse(errno, "success", nil))
	return
}

/**
获取worker所在节点的Ip
 */
func GetWorkerNodeIp(c *gin.Context) {
	var (
		ipArr []string
		err   error
	)

	//调用获取ip的service方法
	if ipArr, err = service.GetWorkerNodeIp(); err != nil {
		c.JSON(http.StatusOK, common.BuildResponse(-1, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, common.BuildResponse(1, "success", ipArr))
}
