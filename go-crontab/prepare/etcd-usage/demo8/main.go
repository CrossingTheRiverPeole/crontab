package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		kv     clientv3.KV
		putOp clientv3.Op
		opResp clientv3.OpResponse
		getOp clientv3.Op
	)
	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	// 创建kv
	kv = clientv3.NewKV(client)

	//创建op：operation
	putOp = clientv3.OpPut("/cron/jobs/job8","123")
	// 执行op，向etcd中写入数据
	if opResp, err = kv.Do(context.TODO(),putOp); err != nil{
		fmt.Println(err)
		return
	}


	fmt.Println(opResp.Put().Header.Revision)


	//创建op：从etcd中获取kv键值对
	getOp = clientv3.OpGet("/cron/jobs/job8")

	if opResp,err = kv.Do(context.TODO(),getOp); err != nil{
		fmt.Println(err)
		return
	}
	//打印获得数据
	fmt.Println(opResp.Get().Kvs[0].ModRevision)
	fmt.Println("数据value", string(opResp.Get().Kvs[0].Value))


}
