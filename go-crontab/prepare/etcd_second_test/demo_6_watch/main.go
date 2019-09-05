package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	var (
		client    *clientv3.Client
		config    clientv3.Config
		err       error
		kv        clientv3.KV
		getResp   *clientv3.GetResponse
		revision  int64
		watch     clientv3.Watcher
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		events    []*clientv3.Event
		event     *clientv3.Event
		ctx context.Context
		cancelFunc context.CancelFunc
	)

	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	//获取kv
	if client, err = clientv3.New(config); err != nil {
		fmt.Println("创建客户端错误")
		return
	}
	watch = clientv3.NewWatcher(client)
	// 获取kv
	kv = clientv3.NewKV(client)
	//先向etcd中写入一个值
	go func() {
		for {
			kv.Put(context.TODO(), "/cron/jobs/testsong", "song")
			kv.Delete(context.TODO(), "/cron/jobs/testsong")
			time.Sleep(2 * time.Second)
		}
	}()

	//先获取到一个key的revision
	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/testsong"); err != nil {
		fmt.Println("获取失败")
		return
	}

	// 从这个revision开始监听
	revision = getResp.Header.Revision + 1

	//5秒之后取消监听
	ctx, cancelFunc = context.WithCancel(context.TODO())
	time.AfterFunc(5 * time.Second, func() {
		cancelFunc()
	})

	watchChan = watch.Watch(ctx, "/cron/jobs/testsong", clientv3.WithRev(revision))
	//遍历chan: chan可以通过for进行循环遍历，
	for watchResp = range watchChan {
		events = watchResp.Events
		for _, event = range events {
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Println("put事件")
			case clientv3.EventTypeDelete:
				fmt.Println("delete事件")
			}
		}
	}
}
