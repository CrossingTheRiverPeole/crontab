package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main()  {
	var (
		client *clientv3.Client
		config clientv3.Config
		kv clientv3.KV
		putOp clientv3.Op
	)

	config = clientv3.Config{
		Endpoints: []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	// 获取client
	client,_ = clientv3.New(config)

	// 获取kv
	kv = clientv3.NewKV(client)

	//putOP
	putOp = clientv3.OpPut("/cron/jobs/job7", "song")
	opResp, _ := kv.Do(context.TODO(), putOp)
	fmt.Println("put revision",opResp.Put().Header.Revision)

	//get操作
	getOp := clientv3.OpGet("/cron/jobs/job7")
	opGet, _ := kv.Do(context.TODO(), getOp)
	fmt.Println("get revision",opGet.Get().Kvs[0].ModRevision)
}
