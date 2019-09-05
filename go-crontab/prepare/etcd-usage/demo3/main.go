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
		err error
		kv clientv3.KV
		getResp *clientv3.GetResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil{
		fmt.Println(err)
		return
	}

	// 创建一个kvClient
	kv =  clientv3.NewKV(client)



	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(getResp.Count)
	fmt.Println(getResp.Kvs)



}
