package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"time"
)

type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

/**
定义结构体作为插入到mongo的数据
 */
type LogRecord struct {
	JobName   string    `bson:"jobName"`   //任务名
	Command   string    `bson:"command"`   //shell命令
	Err       string    `bson:"err"`       //脚本错误
	Content   string    `bson:"content"`   // 脚本输出
	TimePoint TimePoint `bson:"timePoint"` //执行时间点
}

func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
		record *LogRecord
		result*mongo.InsertOneResult
		docId  primitive.ObjectID
	)

	if client, err = mongo.Connect(context.TODO(), "mongodb://10.20.1.185:27017"); err != nil {
		fmt.Println("连接MongoDB失败", err)
		return
	}
	//检测连接错误
	if err = client.Ping(context.TODO(),nil); err != nil{
		fmt.Println("连接失败", err)
		return
	}

	//选择连接的数据库
	database = client.Database("log")

	//选择表log
	collection = database.Collection("log")

	//创建一个记录
	record = &LogRecord{
		JobName: "job1",
		Command: "echo hello",
		Err: "",
		Content: "hello",
		TimePoint:TimePoint{
			StartTime: time.Now().Unix(),
			EndTime: time.Now().Unix() + 10,
		},
	}

	// 插入记录
	if result, err = collection.InsertOne(context.TODO(),record); err != nil{
		fmt.Println("插入数据错误", err)
		return
	}

	//_id：默认生成一个全局唯一的id，ObjectID：12字节的二进制
	docId = result.InsertedID.(primitive.ObjectID)
	fmt.Println("自增id", docId.Hex())
}
