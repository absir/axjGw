package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func register(ctx context.Context, client *clientv3.Client, name string, val string, ttl int64, keep int64) func(sleep bool) error {
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	var leaseId clientv3.LeaseID = 0
	return func(sleep bool) error {
		if leaseId == 0 {
			resp, err := lease.Grant(ctx, ttl)
			if err != nil {
				return err
			}

			key := name + fmt.Sprintf("%d", resp.ID)
			if _, err := kv.Put(ctx, key, val, clientv3.WithLease(resp.ID)); err != nil {
				return err
			}

			leaseId = resp.ID

		} else {
			// 续约租约，如果租约已经过期将curLeaseId复位到0重新走创建租约的逻辑
			if _, err := lease.KeepAliveOnce(ctx, leaseId); err == rpctypes.ErrLeaseNotFound {
				leaseId = 0
				return nil
			}

			if keep <= 0 {
				keep = ttl - 10
			}

			time.Sleep(time.Duration(keep) * time.Second)
		}

		return nil
	}
}

// 监控服务目录下的事件
func watchPrefix(ctx context.Context, client *clientv3.Client, name string) {
	watcher := clientv3.NewWatcher(client)
	// Watch 服务目录下的更新
	watchChan := watcher.Watch(context.TODO(), name, clientv3.WithPrefix())
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			service.mutex.Lock()
			switch event.Type {
			case mvccpb.PUT: //PUT事件，目录下有了新key
				service.nodes[string(event.Kv.Key)] = string(event.Kv.Value)
			case mvccpb.DELETE: //DELETE事件，目录中有key被删掉(Lease过期，key 也会被删掉)
				delete(service.nodes, string(event.Kv.Key))
			}
			service.mutex.Unlock()
		}
	}
}
