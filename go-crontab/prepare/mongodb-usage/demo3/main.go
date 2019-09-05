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
	EndTime int64 `bson:"endTime"`
}

type LogRecord struct {
	JobName string `bson:"jobName"`
	Command string `bson:"command"`
	Err string `bson:"err"`
	Content string `bson:"content"`
	TimePoint TimePoint `bson:"timePoint"`

}

/**
插入多条记录
 */
func main()  {

	var(
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
		record *LogRecord
		records []interface{}
		result *mongo.InsertManyResult
		insertId interface{}
		docId primitive.ObjectID
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

	// 获取数据库
	database = client.Database("log")
	// 获取表即collection
	collection = database.Collection("log")

	//创建record
	record = &LogRecord{
		JobName: "job3",
		Command: "echo hello",
		Err: "",
		Content: "hello",
		TimePoint:TimePoint{
			StartTime: time.Now().Unix(),
			EndTime: time.Now().Unix() + 10,
		},
	}

	// 创建多个record
	records = []interface{}{record,record,record}

	if result, err = collection.InsertMany(context.TODO(),records); err != nil{
		fmt.Println("插入多条记录出错", err)
		return
	}

	for _, insertId = range result.InsertedIDs{
		docId = insertId.(primitive.ObjectID)
		fmt.Println("自增ID", docId.Hex())
	}

}
