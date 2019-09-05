package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
使用kv进行put的操作的时候，每个key值会有一个revision，每次提交的时候revision会加一
put的时候必须带着withPrevision，只要这样才会获取到上一个版本的key所对应的value
 */
func main() {

	var (
		client *clientv3.Client
		config clientv3.Config
		err    error
		resp   *clientv3.PutResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: time.Second * 5,
	}

	//
	if client, err = clientv3.New(config); err != nil {
		fmt.Println("创建客户端失败")
		return
	}
	// 获取kv
	kv := clientv3.NewKV(client)
	//进行put操作，带着withPrevKV,只有这样会带着上一个版本的value
	if resp, err = kv.Put(context.TODO(), "/cron/job/", "jobdddd", clientv3.WithPrevKV()); err != nil {
		fmt.Println("put 操作失败")
	} else {
		// 获取revision
		fmt.Println(resp.Header.Revision)
		//如果是第一次的话是没有值的，判断一下value的值是否为空
		if resp.PrevKv != nil {
			fmt.Println(string(resp.PrevKv.Value))
		}
	}

}
