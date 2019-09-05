package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
)
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime int64 `bson:"endTime"`
}

type LogRecord struct {
	JobName string `bson:"jobName"`
	Command string `bson:"command"`
	Err string `bson:"err"`
	Content string `bson:"content"`
	TimePoint TimePoint `bson:"timePoint"`
}

type FindByJobName struct {
	JobName string `bson:"jobName"`
}
/**
获取数据并反序列化为结构体
 */
func main()  {
	var(
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
		cond *FindByJobName
		cursor mongo.Cursor
		logRecord *LogRecord
	)

	client, err = mongo.Connect(context.TODO(), "mongodb://10.20.1.185:27017")
	if err != nil{
		fmt.Println("连接错误", err)
		return
	}
	//check connection
	err  = client.Ping(context.TODO(),nil)
	if err != nil {
		fmt.Println("连接错误", err)
		return
	}
	// 连接到数据库
	database = client.Database("log")
	//获取表
	collection = database.Collection("log")

	//从表中获取数据并发序列化为结构体(根据条件进行查找)
	cond = &FindByJobName{
		JobName: "job3",
	}
	// 根据条件进行获取，返回的是游标，然后根据游标获取collection（即表）查询的内容
	if cursor,err = collection.Find(context.TODO(),cond); err != nil{
		fmt.Println("查询出现问题", err)
		return
	}
	// 获取结果并反序列化为结构体
	for cursor.Next(context.TODO()){
		logRecord = &LogRecord{}
		if err = cursor.Decode(logRecord); err != nil{
			fmt.Println("反序列胡失败", err)
			return
		}
		fmt.Println(*logRecord)
	}











}
