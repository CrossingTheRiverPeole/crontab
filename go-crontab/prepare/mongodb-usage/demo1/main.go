package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
)

/**
mongodb 客户端
 */
func main() {
	var(
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
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

	//选择数据库
	database = client.Database("my_db_test")

	//选择表
	collection = database.Collection("my_collections")

	fmt.Println(collection.Name())



}
