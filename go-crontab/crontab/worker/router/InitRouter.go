package router

import (
	"github.com/gin-gonic/gin"
	"go-crontab/crontab/worker/config"
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

	apiv1 = r.Group("/api")

	{
		   apiv1.GET("/test",)
	}

	return r
}
