package config

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"time"
)

var (
	G_logSink *LogSink
)

type JobLog struct {
	JobName      string `bson:"jobName"`
	Command      string `bson:"command"`
	OutPut       string `bson:"outPut"`
	Err          string `bson:"err"`
	PlanTime     int64  `bson:"planTime"`
	ScheduleTime int64  `bson:"scheduleTime"`
	StartTime    int64  `bson:"startTime"`
	EndTime      int64  `bson:"endTime"`
}

/**
批量存放日志
 */
type LogBatch struct {
	Logs []interface{}
}

/**
日志提交相关结构体
 */
type LogSink struct {
	client         *mongo.Client // mongo client
	logCollection  *mongo.Collection
	logChan        chan *JobLog
	autoCommitChan chan *LogBatch
}

/**
初始化jobLog：
1）初始化mongodb
2)选择collection
 */
func InitJobLog() (err error) {
	var (
		client *mongo.Client
	)
	uri := G_config.MongodbUri
	fmt.Println("mongouri", uri)
	client, err = mongo.Connect(context.TODO(), uri)
	// 测试是否可以ping通
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("mongodb连接错误", err)
		return
	}

	// 初始化
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("log").Collection("log"),
		logChan:        make(chan *JobLog, 1000),
		autoCommitChan: make(chan *LogBatch, 1000),
	}

	//处理channel接收的日志
	go G_logSink.writeLoop()
	return
}

/**
向mongodb中批量插入数据
 */
func (logSink *LogSink) saveLogs(logs []interface{}) {
	logSink.logCollection.InsertMany(context.TODO(), logs)
}

/**
处理
 */
func (logSink *LogSink) writeLoop() () {
	var (
		jobLog       *JobLog
		logBatch     *LogBatch
		commitTimer  *time.Timer
		timeoutBatch *LogBatch
	)

	// 获取日志通道传递过来的日志信息
	for {
		select {
		case jobLog = <-G_logSink.logChan:
			if logBatch == nil {
				logBatch = &LogBatch{}
			}
			commitTimer = time.AfterFunc(time.Duration(1000)*time.Millisecond,
				func(batch *LogBatch) func() {
					return func() {
						logSink.autoCommitChan <- logBatch
					}
				}(logBatch))
			// 把日志放在logBatch中
			logBatch.Logs = append(logBatch.Logs, jobLog)
			// 如果批次满了，立即写入到mongo中
			if len(logBatch.Logs) >= G_config.LogBatchSize {
				logSink.saveLogs(logBatch.Logs)
				// 把logBatch重新置为空
				logBatch = nil
				//取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan:
			if timeoutBatch != logBatch { // 当处于一秒的时间触发器的时候正好批次满了并且触发了时间触发器此时可能就存在批次重复问题
				continue
			}
			//保存日志
			logSink.saveLogs(logBatch.Logs)
			// 批次置为空
			logBatch = nil

		}

	}
}

/**
向日志通道中添加日志
 */
func (logSink *LogSink) appendLog(jobLog *JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default: // 日志满了不做处理
	}
}
