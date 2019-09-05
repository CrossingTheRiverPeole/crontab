package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
处理自动续约以及事务
 */
func main() {
	var (
		config          clientv3.Config
		client          *clientv3.Client
		err             error
		kv              clientv3.KV
		lease           clientv3.Lease
		leaseId         clientv3.LeaseID
		leaseGreantResp *clientv3.LeaseGrantResponse
		ctx             context.Context
		cancelFunc      context.CancelFunc
		keepRespChan    <-chan *clientv3.LeaseKeepAliveResponse
		keepResp        *clientv3.LeaseKeepAliveResponse
		txn             clientv3.Txn
		txnResp         *clientv3.TxnResponse
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
	leaseGreantResp, err = lease.Grant(context.TODO(), 5)
	// 获取一个租约的id
	leaseId = leaseGreantResp.ID

	// 设置租约自动续租的上下文
	ctx, cancelFunc = context.WithCancel(context.TODO())
	// 取消自动续约
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId)
	//租约自动续租
	if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		fmt.Println(err)
		return
	}

	//处理续约应答的协程
	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepRespChan == nil || keepResp == nil {
					fmt.Println("租约已失效")
					goto END
				}
				fmt.Println("收到自动续约的应答, 续约的revision", keepResp.Revision)
			}
		}
	END:
	}()

	// 创建kv：
	kv = clientv3.NewKV(client)
	//创建事务
	txn = kv.Txn(context.TODO())
	// 事务：if 不存在key， then设置它，else抢锁失败
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/jobs/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/jobs/job9", "XXX", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/jobs/job9")) //否则抢锁失败

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	// 判断是否抢到了锁
	if !txnResp.Succeeded {
		fmt.Println("锁被占用，", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	// 处理业务
	fmt.Println("处理业务")
	time.Sleep(10 * time.Second)

	//  释放锁（取消自动租约，释放租约）
	//defer 会把租约释放掉，关联的KV就被删除了

}
