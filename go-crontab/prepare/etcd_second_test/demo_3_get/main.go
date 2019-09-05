package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
 get操作:该操作可以进行各种with操作
 */
func main() {
	var (
		client  *clientv3.Client
		config  clientv3.Config
		err     error
		getResp *clientv3.GetResponse
	)

	// 获取
	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	// 获取client
	client, _ = clientv3.New(config)

	// 获取kvclient并进行获取值操作
	kv := clientv3.NewKV(client)

	if getResp, err = kv.Get(context.TODO(), "/cron/job/",clientv3.WithPrefix()); err != nil {
		fmt.Println("获取操作值错误")
		return
	} else {
		fmt.Println(getResp.Kvs, getResp.Count)
	}

}
