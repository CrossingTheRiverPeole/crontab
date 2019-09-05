package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-crontab/crontab/common"
	"go-crontab/crontab/master/config"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"strings"
)

func SaveJobService(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		jobValue  []byte
		oldJobObj common.Job
		putResp   *clientv3.PutResponse
	)
	// job的key
	jobKey = common.JOB_SAVE_DIR + job.Name
	//任务信息
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	// 向etcd中写入数据
	if putResp, err = config.G_jobMgr.Kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}

	if putResp.PrevKv != nil {
		//对旧值进行反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

/**
根据前缀获取所有的任务列表
 */
func ListJobs() (jobList []*common.Job, err error) {
	var (
		jobKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)
	// jobs前缀
	jobKey = common.JOB_SAVE_DIR

	// 获取所有的任务列表
	if getResp, err = config.G_jobMgr.Kv.Get(context.TODO(), jobKey, clientv3.WithPrefix()); err != nil {
		fmt.Println("获取任务列表出错", err)
		return
	}

	//实例化一个切片
	jobList = make([]*common.Job, 0)
	//编列获取到的任务并进行反序列化
	for _, kvPair = range getResp.Kvs {
		// 实例化job
		job = &common.Job{}
		//反序列化job
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	// 返回获取到的job
	return
}

/**
根据任务名称删除任务
 */
func JobRemove(jobName string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)
	jobKey = common.JOB_SAVE_DIR + jobName
	// 删除任务
	if delResp, err = config.G_jobMgr.Kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		fmt.Println("删除", jobName, "任务失败", err)
		return
	}

	// 对任务进行反序列化
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			fmt.Println("反序列化失败", err)
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

/**
任务强制杀死：把任务名称写入到目录下，worker进行监听，监听到之后杀死任务
 */
func JobKill(jobName string) (err error) {
	var (
		jobKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)
	// 要杀死的任务的key值
	jobKey = common.JOB_KILLER_DIR + jobName

	//获取leaseId：租约设置为一秒，写入进去之后，一秒之后过期
	if leaseGrantResp, err = config.G_jobMgr.Lease.Grant(context.TODO(), 1); err != nil {
		fmt.Println("设置租约出现错误", err)
		return
	}
	// 获取leaseid
	leaseId = leaseGrantResp.ID
	// 写入etcd要杀死的job
	if _, err = config.G_jobMgr.Kv.Put(context.TODO(), jobKey, "", clientv3.WithLease(leaseId)); err != nil {
		fmt.Println("写入etcd要杀死job失败", err)
		return
	}
	return
}

/**
获取worker所在节点的ip（所有worker）
 */
func GetWorkerNodeIp() (ipAddrs []string, err error) {
	var (
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		workerNodeIp string
	)

	if getResp, err = config.G_jobMgr.Kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
		fmt.Println("获取节点ip出现错误", err)
		return
	}

	ipAddrs = make([]string, 0)
	// 遍历获取到的keyvalue值
	for _, kvPair = range getResp.Kvs {
		workerNodeIp = extractWorkerName(string(kvPair.Key))
		ipAddrs = append(ipAddrs,workerNodeIp)
	}
	return

}

func extractWorkerName(key string) (workIp string) {
	return strings.TrimPrefix(key, common.JOB_WORKER_DIR)
}
