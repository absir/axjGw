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
    Result conn(1: i64 cid, 2: string gid, 3: string unique)
    // 断开
    oneway void disc(1: i64 cid, 2: string gid, 3: string unique, 4: i32 connVer)
    // 存活
    Result alive(1: i64 cid)
    // 服务编号
    Result rid(1: i64 cid, 2: string name, 3: i32 rid)
    // 服务编号
    Result rids(1: i64 cid, 2: map<string, i32> rids)
    // 最新消息通知
    Result last(1: i64 cid, 2: string gid, 3: i32 connVer)
    // 推送
    Result push(1: i64 cid, 2: string uri, 3: binary bytes, 4: bool isolate, 5: i64 id);
    // 消息队列初始化
    Result gQueue(1: string gid, 2: i64 cid, 3: string unique, 4: bool clear);
    // 消息队列清理
    Result gClear(1: string gid, 2: bool queue, 3: bool last)
    // 主动获取消息
    Result gLasts(1: string gid, 2: i64 cid, 3: string unique, 4: i64 lastId);
    // 推送
    Result gPush(1: string gid, 2: string uri, 3: binary bytes, 4: bool isolate, 6: i32 qs, 7: bool queue, 8: string unique, 9: i64 fid)
    // 组更新、删除
    Result gDirty(1: string gid);
}