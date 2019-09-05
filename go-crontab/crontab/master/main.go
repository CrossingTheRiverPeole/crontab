package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-crontab/crontab/master/config"
	"go-crontab/crontab/master/router"
	"net/http"
	"runtime"
	"time"
)

var (
	configFile string //配置文件路径
)

func initArgs() {
	flag.StringVar(&configFile, "config", "crontab/master/config/config.yaml", "指定configFile路径")
	flag.Parse()
}

/**
初始化环境服务
 */
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

/**
进行初始化
 */
func main() {
	var (
		err error
		r   *gin.Engine
		s   *http.Server
	)
	// 初始化命令行参数
	initArgs() //获取配置文件
	//初始化线程
	initEnv()
	//加载配置
	if err = config.InitConfig(configFile); err != nil {
		fmt.Println("初始化config出现错误", err)
		goto ERR
	}
	//任务管理器
	if err = config.InitJobMgr(); err != nil {
		fmt.Println("初始化jobMgr出现错误", err)
		goto ERR
	}

	//初始化路由
	r = router.InitRouter()
	s = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.G_config.ApiPort),
		Handler:      r,
		ReadTimeout:  time.Duration(config.G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(config.G_config.ApiWriteTimeout) * time.Millisecond,
	}
	//启动服务
	s.ListenAndServe()

ERR:
	fmt.Println("运行出错", err)
}
