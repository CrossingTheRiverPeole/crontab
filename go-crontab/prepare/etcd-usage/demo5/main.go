package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

func main()  {
	var  (
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		delResponse *clientv3.DeleteResponse
		kvpair *mvccpb.KeyValue
	)

	config = clientv3.Config{
		Endpoints: []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	client,err = clientv3.New(config)
	kv  = clientv3.NewKV(client)


	//删除KV
	if delResponse, err = kv.Delete(context.TODO(), "/cron/jobs/job1", clientv3.WithFromKey(),clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
		return
	}
	//被删除之前的value是什么
	if len(delResponse.PrevKvs) != 0{
		for _,kvpair = range delResponse.PrevKvs  {
			fmt.Println("删除了：", string(kvpair.Key), string(kvpair.Value))
		}
	}

}
