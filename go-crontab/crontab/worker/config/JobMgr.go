package config

import (
	"context"
	"fmt"
	"go-crontab/crontab/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

var (
	// 全局单例
	G_jobMgr *JobMgr
)

type JobMgr struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		lease   clientv3.Lease
		kv      clientv3.KV
		watcher clientv3.Watcher
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println("连接etcd出错", err)
		return
	}

	// 创建
	kv = clientv3.NewKV(client)
	//创建lease
	lease = clientv3.NewLease(client)
	// 创建watcher
	watcher = clientv3.NewWatcher(client)

	//给全局单例变量赋值
	G_jobMgr = &JobMgr{
		Client:  client,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	// 监听任务
	G_jobMgr.watcherJobs()

	//强制杀死任务
	G_jobMgr.watchKiller()

	return
}

/**

 */
func (jobMgr *JobMgr) watcherJobs() {
	var (
		err                error
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		job                *common.Job
		jobEvent           *JobEvent
		watchChan          clientv3.WatchChan
		watchStartRevision int64
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
	)
	//get一下/cron/jobs/目录下的所有任务，并且获知当前revision
	if getResp, err = jobMgr.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		fmt.Println("获取任务revision失败")
		return
	}

	// 获取所有job并进行发序列化
	for _, kvPair = range getResp.Kvs {
		if job, err = UnpackJob(kvPair.Value); err == nil {
			// job解析成jobEvent
			jobEvent = BuildJobEvent(common.JOB_EVENT_SAVE, job)
			// 同步传送给调度协程
			G_scheduler.PushJobEvent(jobEvent)
		}
	}

	// 启动协程监听/cron/jobs下的任务，从revision开始监听
	go func() {
		watchStartRevision = getResp.Header.Revision
		// 监听/cron/jobs/目录的后续变化
		watchChan = G_jobMgr.Watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case clientv3.EventTypePut:
					if job, err = UnpackJob(watchEvent.Kv.Value); err != nil {
						continue // 如果etcd中监听到的值无法发序列化为job，不处理，继续下一次
					}
					jobEvent = BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case clientv3.EventTypeDelete:
					// delete /cron/jobs/job10 获取job名称
					jobName = ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					// 构建jobEvent
					jobEvent = BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				// 监听到事件之后推送到scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
}

/**
强杀方法：监听/cron/killer路径，用来取消任务
 */
func (jobMgr *JobMgr) watchKiller() {
	var (
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		event     *clientv3.Event
		job       *common.Job
		jobName   string
		jobEvent  *JobEvent
	)
	//启动携程去监听
	go func() {
		// 监听/cron/killer目录的变化
		watchChan = jobMgr.Watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, event = range watchResp.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					fmt.Println("监听到强制杀死任务事件")
					//构建jobEvent
					jobName = ExtracrKillerName(string(event.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					//构建jobEvent
					jobEvent = BuildJobEvent(common.JOB_EVNET_KILL, job)

					//发送jobEvent
					G_scheduler.jobEventChan <- jobEvent
				case clientv3.EventTypeDelete:
					//如果是删除时间不处理，因为强杀事件写入etcd中的keyvalue值是一个有租约的值，很快就会过期并删除
				}
			}
		}

	}()

}

/**
创建锁
 */
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitLock(jobName, jobMgr.Kv, jobMgr.Lease)
	return jobLock
}
