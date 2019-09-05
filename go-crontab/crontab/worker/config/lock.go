package config

import (
	"context"
	"fmt"
	"go-crontab/crontab/common"
	"go.etcd.io/etcd/clientv3"
)

type JobLock struct {
	JobName    string // 任务名
	CancelFunc context.CancelFunc
	LeaseId    clientv3.LeaseID
	Kv         clientv3.KV
	Lease      clientv3.Lease
	IsLock     bool
}

/**
试图上锁
 */
func (jobLock *JobLock) TryLock() (err error) {

	var (
		leaseGrantResp      *clientv3.LeaseGrantResponse
		leaseId             clientv3.LeaseID
		ctx                 context.Context
		cancelFunc          context.CancelFunc
		leaseKeepAlivedResp <-chan *clientv3.LeaseKeepAliveResponse
		keepResp            *clientv3.LeaseKeepAliveResponse
		txn                 clientv3.Txn
		txnResp             *clientv3.TxnResponse
		lockKey             string
	)

	//1.生成租约，5秒
	if leaseGrantResp, err = jobLock.Lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	// 2.租约续约
	leaseId = leaseGrantResp.ID
	//用于取消自动续约
	ctx, cancelFunc = context.WithCancel(context.TODO())
	// 自动续约失败
	if leaseKeepAlivedResp, err = jobLock.Lease.KeepAlive(ctx, leaseId); err != nil {
		goto FAIL
	}
	// 处理自动续约应答
	go func() {
		for {
			select {
			case keepResp = <-leaseKeepAlivedResp:
				goto END
			}
		}
	END:
	}()

	//4.创建事务
	txn = jobLock.Kv.Txn(context.TODO())
	//锁路径
	lockKey = common.JOB_LOCK_DIR + jobLock.JobName
	// 事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, jobLock.JobName, clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	// 5 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	// 6.成功返回，失败则释放租约
	if !txnResp.Succeeded {
		fmt.Println("抢锁失败")
		goto FAIL
	}

	// 抢锁成功
	jobLock.LeaseId = leaseId
	jobLock.IsLock = true
	jobLock.CancelFunc = cancelFunc
FAIL:
	cancelFunc()
	jobLock.Lease.Revoke(context.TODO(), leaseId)
	return
}

/**
初始化锁
 */
func InitLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		Kv:      kv,
		JobName: jobName,
		Lease:   lease,
	}
	return
}

/**
释放锁
 */
func (jobLock *JobLock) Unlock() {
	if jobLock.IsLock {
		jobLock.CancelFunc()
		jobLock.Lease.Revoke(context.TODO(), jobLock.LeaseId)
	}
}
