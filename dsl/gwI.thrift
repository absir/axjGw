namespace go gw

include "gw.thrift"

enum Result {
    // 失败
    Fail,
    // 成功
    Succ,
    // ID不存在
    IdNone,
    // 分布冲突
    ProdErr,
}

struct Msg {
    1: string uri
    2: binary bytes
}

// 网关内部
service GatewayI {
    // 关闭连接
    Result close(1: i64 cid, 2: string reason)
    // 软关闭连接
    Result kick(1: i64 cid, 2: binary bytes)
    // 连接
    Result conn(1: i64 cid, 2: string sid, 3: string unique)
    // 存活
    Result alive(1: i64 cid)
    // 服务编号
    Result rid(1: i64 cid, 2: string name, 3: i32 rid)
    // 服务编号
    Result rids(1: i64 cid, 2: map<string, i32> rids)
    // 最新消息通知
    Result last(1: i64 cid)
    // 推送
    Result push(1: i64 cid, 2: string uri, 3: binary bytes, 4: bool isolate);
    // 组更新、删除
    Result dirty(1: string gid);
}