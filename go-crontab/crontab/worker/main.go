package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-crontab/crontab/worker/config"
	"go-crontab/crontab/worker/router"
	"net/http"
	"runtime"
	"time"
)

var (
	configFile string //配置文件路径
)

func initArgs() {
	flag.StringVar(&configFile, "config", "crontab/worker/config/config.yaml", "指定configFile路径")
	flag.Parse()
}

/**
初始化环境服务
 */
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
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

	// 初始化worker节点注册
	if err = config.InitRegister(); err != nil {
		fmt.Println("初始化注册worker节点出错")
		goto ERR
	}

	// 初始化mongodb连接
	if config.InitJobLog(); err != nil {
		fmt.Println("mongodb连接出错", err)
		goto ERR
	}

	// 初始化调度器
	if err = config.InitScheduler(); err != nil {
		fmt.Println("初始化Scheduler出现错误", err)
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
}
