namespace go gw

include "gw.thrift"

// 网关内部
service GatewayI {
    // 服务编号
    void rid(1: i64 cid, 2: string name, 3: i32 rid)
    // 服务编号
    void rids(1: i64 cid, 2: map<string, i32> rids)
    // 组更新、删除
    void dirty(1: string sid);
    // 推送
    bool push(1: i64 cid, 2: string uri, 3: binary bytes);
    // 推送 oneway
    oneway void pushO(1: i64 cid, 2: string uri, 3: binary bytes);
}