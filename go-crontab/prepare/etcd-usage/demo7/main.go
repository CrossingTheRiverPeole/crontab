package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
监听kv值并进行续约
 */
func main() {
	var (
		config         clientv3.Config
		client         *clientv3.Client
		err            error
		kv             clientv3.KV
		putResp        *clientv3.PutResponse
		getResp        *clientv3.GetResponse
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		keepResponse   *clientv3.LeaseKeepAliveResponse
		ctx            context.Context
		cancelFunc     context.CancelFunc
	)
	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	lease = clientv3.NewLease(client)
	//生成一个10秒的租约
	leaseGrantResp, err = lease.Grant(context.TODO(), 10)
	// 拿到租约的id
	leaseId = leaseGrantResp.ID

	//五秒之后,取消自动续约
	ctx, cancelFunc = context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})

	if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		fmt.Println(err)
		return
	}

	//处理续约应答的协程
	go func() {
		for {
			select {
			case keepResponse = <-keepRespChan:
				if keepRespChan == nil || keepResponse == nil{
					fmt.Println("租约已失效")
					goto END
				} else {
					fmt.Println("收到自动续约应答：", keepResponse.ID)
				}
			}
		}
	END:
	}()

	// 生成kv客户端
	kv = clientv3.NewKV(client)

	//向etcd中写入一个kv值，监听
	if putResp, err = kv.Put(context.TODO(), "/cron/lock/job1", "test", clientv3.WithLease(leaseId)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("写入成功：", putResp.Header.Revision)

	//定时的看一下key值有没有过期（如果过期之后kv的值就没有了）
	for {
		if getResp, err = kv.Get(context.TODO(), "/cron/lock/job1"); err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("kv过期了")
			break
		}

		fmt.Println("kv还没过期", getResp.Kvs)
		time.Sleep(2 * time.Second)
	}
}
