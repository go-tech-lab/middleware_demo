package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"strings"
	"time"
)

var etcdClient *clientv3.Client //锁client
var lockKey = "TestParallelLockSameKey-lockKey"

func init() {
	intClient()
}

func PutValue() {
	ctx := context.TODO()
	put1, _ := etcdClient.Put(ctx, "key-test-one", "key-test-one")
	if put1.PrevKv != nil {
		println("put-1=", string(put1.PrevKv.Key), "value=", string(put1.PrevKv.Value))
	}
	put2, _ := etcdClient.Put(ctx, "key-test-one", "key-test-one2")
	if put2.PrevKv != nil {
		println("put-2=", string(put2.PrevKv.Key), "value=", string(put2.PrevKv.Value))
	}
	resp, err1 := etcdClient.Get(ctx, "key-test-one", clientv3.WithPrefix())
	if err1 != nil {
		println(err1.Error())
	}
	kv := resp.Kvs
	if len(resp.Kvs) == 0 {
		println("g1 kvs empty first")
	} else {
		for i := 0; i < len(kv); i++ {
			println("g1-key=", string(kv[i].Key), "value=", string(kv[i].Value))
		}
	}
}

func Test() {
	EtcdEndpoints := "etcd-node1.i.insurance.test.shopee.io:2579,etcd-node2.i.insurance.test.shopee.io:2579,etcd-node3.i.insurance.test.shopee.io:2579"
	endpoints := strings.Split(EtcdEndpoints, ",")
	//请将etcd配置项补充完整
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(50) * time.Second,
	})

	etcdClient = cli

	if err != nil {
		panic("connect etcd error")
	}

	go func() {
		println("g1 start")
		ctx := context.Background()
		resp, err1 := etcdClient.Get(ctx, lockKey, clientv3.WithPrefix())
		if err1 != nil {
			println(err1.Error())
		}
		kv := resp.Kvs
		if len(resp.Kvs) == 0 {
			println("g1 kvs empty first")
		} else {
			for i := 1; i < len(kv); i++ {
				println("g1-key=", string(kv[i].Key), "value=", string(kv[i].Value))
			}
		}

		session, _ := concurrency.NewSession(etcdClient, concurrency.WithTTL(5))
		lock := GetLock(session)
		lock.Lock(ctx)
		resp, err2 := etcdClient.Get(ctx, lockKey, clientv3.WithPrefix())
		if err2 != nil {
			println(err2.Error())
		}
		kv = resp.Kvs
		for i := 1; i < len(kv); i++ {
			println("g1-key=", string(kv[i].Key), "value=", string(kv[i].Value))
		}
	}()

	//go func() {
	//	println("g2 start")
	//	time.Sleep(time.Second * 3)
	//	ctx := context.Background()
	//	lock := GetLock(NewSession())
	//	defer func() {
	//		lock.Unlock(ctx)
	//	}()
	//	resp, err1 := etcdClient.Get(ctx, lockKey, clientv3.WithPrefix())
	//	if err1 != nil {
	//		println(err1.Error())
	//	}
	//	if len(resp.Kvs) == 0 {
	//		println("g2 kvs empty first")
	//		return
	//	}
	//	kv := resp.Kvs
	//	for i := 1; i < len(kv); i++ {
	//		println("g2-key=", string(kv[i].Key), "value=", string(kv[i].Value))
	//	}
	//	lock.Lock(ctx)
	//	time.Sleep(time.Second * 2)
	//	resp, err2 := etcdClient.Get(ctx, lockKey, clientv3.WithPrefix())
	//	if err2 != nil {
	//		println(err2.Error())
	//	}
	//	kv = resp.Kvs
	//	for i := 1; i < len(kv); i++ {
	//		println("g1-key=", string(kv[i].Key), "value=", string(kv[i].Value))
	//	}
	//}()
}

func intClient() {
	EtcdEndpoints := "etcd-node1.i.insurance.test.shopee.io:2579,etcd-node2.i.insurance.test.shopee.io:2579,etcd-node3.i.insurance.test.shopee.io:2579"
	endpoints := strings.Split(EtcdEndpoints, ",")
	//请将etcd配置项补充完整
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(50) * time.Second,
	})
	if err != nil {
		panic(err.Error())
	}
	etcdClient = cli
}

func GetLock(session *concurrency.Session) *concurrency.Mutex {
	lock := concurrency.NewMutex(session, lockKey)
	return lock
}

func NewSession() *concurrency.Session {
	session, _ := concurrency.NewSession(etcdClient, concurrency.WithTTL(5))
	return session
}

func MillSecondElapse(fun func()) int64 {
	timeStart := time.Now().UnixNano()
	fun()
	timeEnd := time.Now().UnixNano()
	return (timeEnd - timeStart) / 1000 / 1000
}
