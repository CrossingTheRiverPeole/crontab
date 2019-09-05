package test

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"testing"
	"time"
)

func TestEtchWatch(t *testing.T) {
	var (
		config   clientv3.Config
		client   *clientv3.Client
		err      error
		kvClient clientv3.KV
	)

	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	//获取client
	client, err = clientv3.New(config)
	if err != nil {
		log.Fatal("生成客户端出错", err)
	}
	//生成kvClient
	kvClient = clientv3.NewKV(client)

	//模拟etcd中key value值的变化
	go func() {
		for {
			kvClient.Put(context.TODO(), "/cron/jobs/job7", "I am job7")
			kvClient.Delete(context.TODO(), "/cron/jobs/job7")
			time.Sleep(time.Second * 1)
		}
	}()

	resp, err := kvClient.Get(context.TODO(), "/cron/jobs/job7")
	if err != nil {
		log.Fatal("获取监控开始的revision错误", err)
	}
	//从startRevision开始监控
	startRevision := resp.Header.Revision + 1

	//获取watchClient
	watchClient := clientv3.NewWatcher(client)

	//启动监听，5秒之后关闭(执行取消函数)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(time.Second*5, func() {
		cancelFunc()
	})

	watchChan := watchClient.Watch(ctx, "/cron/jobs/job7", clientv3.WithRev(startRevision))

	//遍历
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Println("事件类型", event.Type, "key", string(event.Kv.Key), "value", "value", string(event.Kv.Value), event.PrevKv)
			case clientv3.EventTypeDelete:
				fmt.Println("事件类型", event.Type, "key", string(event.Kv.Key), "value", "value", string(event.Kv.Value), event.PrevKv)
			}
		}
	}

}
