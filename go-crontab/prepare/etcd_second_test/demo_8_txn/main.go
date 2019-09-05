package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	var (
		client                 *clientv3.Client
		config                 clientv3.Config
		err                    error
		lease                  clientv3.Lease
		leaseId                clientv3.LeaseID
		leaseGrantResp         *clientv3.LeaseGrantResponse
		leaseKeepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		leaseKeepAliveResp     *clientv3.LeaseKeepAliveResponse
		ctx                    context.Context
		cancelFunc             context.CancelFunc
		kv                     clientv3.KV
		txn                    clientv3.Txn
		txnResp                *clientv3.TxnResponse
	)

	config = clientv3.Config{
		Endpoints:   []string{"10.20.1.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	// 创建client
	if client, err = clientv3.New(config); err != nil {
		fmt.Println("client create")
		return
	}

	// 创建租约
	lease = clientv3.NewLease(client)
	//
	leaseGrantResp, err = lease.Grant(context.TODO(), 10)
	// 获取租约id
	leaseId = leaseGrantResp.ID

	//取消租约的func
	ctx, cancelFunc = context.WithCancel(context.TODO())

	// 取消锁
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId)
	defer client.Close()

	// 租约自动续约
	leaseKeepAliveRespChan, err = lease.KeepAlive(ctx, leaseId)

	// 启动协程：获取租约自动续约的应答
	go func() {
		for {
			select {
			case leaseKeepAliveResp = <-leaseKeepAliveRespChan:
				if nil == leaseKeepAliveResp {
					fmt.Println("租约已失效")
				} else {
					fmt.Println("收到自动续约应答", leaseKeepAliveResp)
				}
			}
		}
	}()

	kv = clientv3.NewKV(client)

	//开启事务
	txn = kv.Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/jobs/jobtxn"), "=", 0)).
		Then(clientv3.OpPut("/cron/jobs/jobtxn", "xxx", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/jobs/jobtxn")) //否则抢锁失败

	/*// 如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/lock/job9", "xxx", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/lock/job9")) // 否则抢锁失败*/

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println("事务提交失败")
		return
	}

	if !txnResp.Succeeded {
		//抢锁失败，退出
		fmt.Println("抢锁失败", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}
	// 事务提交之后执行业务
	fmt.Println("开始处理业务")
	time.Sleep(time.Second * 10)
	fmt.Println("业务处理完毕")
}
