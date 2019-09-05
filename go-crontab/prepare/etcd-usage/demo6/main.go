package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
监听etcd中value变化，根据prevision进行监听
 */
func main() {
	var (
		config             clientv3.Config
		client             *clientv3.Client
		err                error
		kv                 clientv3.KV
		getResp            *clientv3.GetResponse
		watchStartRevision int64
		watcher            clientv3.Watcher
		cancelFunc         context.CancelFunc
		ctx                context.Context
		watchRespChan clientv3.WatchChan
		watchResp     clientv3.WatchResponse
		event *clientv3.Event
	)
	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	kv = clientv3.NewKV(client)
	// 模拟etcd中kv的变化
	go func() {
		for {
			kv.Put(context.TODO(), "/cron/jobs/job7", "i am job 7")
			kv.Delete(context.TODO(), "/cron/jobs/job7")
			time.Sleep(1 * time.Second)
		}
	}()

	//先get到当前的值，并监听后续变化
	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/job7"); err != nil {
		fmt.Println(err)
		return
	}

	if len(getResp.Kvs) != 0 {
		fmt.Println("当前值:", string(getResp.Kvs[0].Value))
	}

	//当前etcd集群事务ID，单调递增
	watchStartRevision = getResp.Header.Revision + 1

	watcher = clientv3.NewWatcher(client)

	// 启动监听
	fmt.Println("从该版本向后进行监听", watchStartRevision)
	ctx, cancelFunc = context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})

	watchRespChan = watcher.Watch(ctx, "/cron/jobs/job7", clientv3.WithRev(watchStartRevision))

	for watchResp = range watchRespChan {
		for _, event = range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Println("修改为:", string(event.Kv.Value), "Revision", event.Kv.CreateRevision, event.Kv.ModRevision)
			case clientv3.EventTypeDelete:
				fmt.Println("删除了", string(event.Kv.Key), "Revision:", event.Kv.ModRevision)
			}
		}

	}
}
