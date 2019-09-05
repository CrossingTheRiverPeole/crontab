package config

import (
	"context"
	"errors"
	"fmt"
	"go-crontab/crontab/common"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

var (
	G_register *Register
)

/**

 */
type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	ip     string
}

func InitRegister() (err error) {
	//定义变量
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		ip     string
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println("创建etcdClient错误", err)
		goto ERR
	}
	// 创建kv
	kv = clientv3.NewKV(client)

	//创建lease
	lease = clientv3.NewLease(client)

	if ip, err = getIp(); err != nil {
		goto ERR
	}

	G_register = &Register{
		client: client,
		kv:     kv,
		lease:  lease,
		ip:     ip,
	}

	// 初始话节点注册
	go G_register.RegisterWorker()

ERR:
	fmt.Println("出现错误", err)
	return
}

/**
获取主机IP
 */
func getIp() (ipv4 string, err error) {

	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // ip地址
		isIpNet bool
	)

	// 获取所有的网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}

	//获取第一个非Local的IP
	for _, addr = range addrs {
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = errors.New("ip not found")
	return
}

/**
把worker节点所在机器的ip注册到etcd，检测worker节点的健康状态
 */
func (register *Register) RegisterWorker() {

	var (
		leaseId            clientv3.LeaseID
		ctx                context.Context
		cancelFunc         context.CancelFunc
		leaseGrantResp     *clientv3.LeaseGrantResponse
		err                error
		leaseKeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		leaseKeepAliveResp *clientv3.LeaseKeepAliveResponse
	)

	for {

		cancelFunc = nil
		// 创建lease
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		// leaseId
		leaseId = leaseGrantResp.ID

		// 获取context以及cancelFunc
		ctx, cancelFunc = context.WithCancel(context.TODO())

		// 租约续约
		if leaseKeepAliveChan, err = register.lease.KeepAlive(ctx, leaseId); err != nil {
			fmt.Println("租约续约失败：", err)
			goto RETRY
		}

		// 向etcd中注册worker节点所在ip
		if _, err = register.kv.Put(context.TODO(), common.JOB_WORKER_DIR+register.ip, "", clientv3.WithLease(leaseId)); err != nil {
			goto RETRY
		}

		// 处理租约续约应答
		for {
			select {
			case leaseKeepAliveResp = <-leaseKeepAliveChan:
				if leaseKeepAliveResp == nil || leaseKeepAliveChan == nil {
					goto RETRY
				}
			}
		}


	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

}
