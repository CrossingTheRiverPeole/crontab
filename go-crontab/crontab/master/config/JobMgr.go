package config

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	// 全局单例
	G_jobMgr *JobMgr
)
type JobMgr struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		lease  clientv3.Lease
		kv     clientv3.KV
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

	//给全局单例变量赋值
	G_jobMgr = &JobMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	return
}
