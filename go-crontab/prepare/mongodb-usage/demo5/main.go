package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"time"
)

type TimeBeforCond struct {
	Before int64 `bson:"$lt"`
}

type DeleteCon struct {
	BeforeCond TimeBeforCond `bson:"timePoint.startTime"`
}

/**
根据条件删除mongo中的内容
 */
func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
		delCond    *DeleteCon
		delResult  *mongo.DeleteResult
	)

	client, err = mongo.Connect(context.TODO(), "mongodb://10.20.1.185:27017")
	if err != nil {
		fmt.Println("连接错误", err)
		return
	}
	//check connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("连接错误", err)
		return
	}
	// 获取database
	database = client.Database("log")
	// 获取collection（即表）
	collection = database.Collection("log")

	//删除条件：delete("{timePoint.startTime:{"$lt":当前时间}}")
	delCond = &DeleteCon{
		BeforeCond: TimeBeforCond{Before: time.Now().Unix()},
	}
	if delResult, err = collection.DeleteMany(context.TODO(), delCond); err != nil {
		fmt.Println("根据条件删除失败", err)
		return
	}
	fmt.Println("删除行数", delResult.DeletedCount)
}
