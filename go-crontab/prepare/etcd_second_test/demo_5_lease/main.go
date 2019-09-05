package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
租约自动续约
 */
func main() {
	var (
		client         *clientv3.Client
		config         clientv3.Config
		err            error
		getResp        *clientv3.GetResponse
		kv             clientv3.KV
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
		putResp        *clientv3.PutResponse
	    leaseKeepAliveRespChan  <-chan *clientv3.LeaseKeepAliveResponse
	    leaseKeepAliveResp *clientv3.LeaseKeepAliveResponse
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
	// 获取kv
	kv = clientv3.NewKV(client)

	// 获取租约client
	lease = clientv3.NewLease(client)
	// 获取租约
	if leaseGrantResp, err = lease.Grant(context.TODO(), 10); err != nil {
		fmt.Println("获取租约失败")
		return
	}
	// 获取租约id
	leaseId = leaseGrantResp.ID
	// 为key自动续约
	if leaseKeepAliveRespChan, err = lease.KeepAlive(context.TODO(), leaseId); err != nil{
		fmt.Println("自动续约失败")
		return
	}

	// 启动一个协程，从续约应答中获取返回的结果
	go func() {
		for  {
			select {
			case leaseKeepAliveResp =  <- leaseKeepAliveRespChan:
				if leaseKeepAliveResp == nil {
					fmt.Println("租约已经失效了")
						goto END
				}else {
					fmt.Println("收到自动续约应答", leaseKeepAliveResp)
				}
			}
		}
		END:
	}()

	// 向etcd中存入数据
	if putResp, err = kv.Put(context.TODO(), "/cron/jobs/jobtest", "jobtest", clientv3.WithLease(leaseId)); err != nil {
		fmt.Println("向etcd中存入数据失败")
		return
	}
	fmt.Println("写入成功", putResp.Header.Revision)

	for {
		if getResp, err = kv.Get(context.TODO(), "/cron/jobs/jobtest"); err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("key 过期了")
			break
		}
		time.Sleep(2 * time.Second)
		fmt.Println("还没过期", getResp.Kvs)
	}

}
