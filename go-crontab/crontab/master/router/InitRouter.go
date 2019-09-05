package router

import (
	"github.com/gin-gonic/gin"
	"go-crontab/crontab/master/config"
	"go-crontab/crontab/master/controller"
)

func InitRouter() *gin.Engine {
	var (
		r     *gin.Engine
		apiv1 *gin.RouterGroup
	)
	//创建engine
	r = gin.New()
	r.Use(gin.Recovery())
	//设置运行模式
	gin.SetMode(config.G_config.Mode)
	// 注册路由
	apiv1 = r.Group("/api")

	{
		apiv1.POST("/job/save", controller.SaveJob)          // 保存job
		apiv1.GET("/job/list", controller.HandleJobList)     // 获取所有任务列表
		apiv1.DELETE("/job/del", controller.HandleJobRemove) // 根据任务名称删除任务
		apiv1.POST("/job/kill", controller.HandleJobKill)    // 强制杀死任务
		apiv1.GET("/worker/ip", controller.GetWorkerNodeIp)  // 获取注册到etcd中的worker节点所在主机ip
	}
	return r
}
