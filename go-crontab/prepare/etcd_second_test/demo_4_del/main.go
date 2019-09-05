package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	var (
		client  *clientv3.Client
		config  clientv3.Config
		err     error
		delResp *clientv3.DeleteResponse
		kv      clientv3.KV
	)

	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println("创建客户端是吧")
		return
	}
	kv = clientv3.NewKV(client)

	if delResp, err = kv.Delete(context.TODO(), "/cron/job/", clientv3.WithPrefix(),clientv3.WithPrevKV()); err != nil {
		fmt.Println("删除错误")
	}

	for _, kvPir := range delResp.PrevKvs {
		fmt.Println(string(kvPir.Key), string(kvPir.Value))
	}

}
