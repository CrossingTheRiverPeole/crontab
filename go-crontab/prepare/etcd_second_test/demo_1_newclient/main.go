package main

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	var (
		client *clientv3.Client
		config clientv3.Config
		err    error
	)

	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: time.Second * 5,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println("创建客户端错误")
		return
	}

	client = client

}
