namespace go gw
namespace java gen.gw

struct Login {
    // 数字编号
    1: i64 uid;
    // 字符编号
    2: string sid;
    // 唯一标识(一个标识，只允许一个Conn)
    3: string unique;
    // 最大请求数
    4: i32 limit;
    // 路由服务编号
    5: i32 rid;
    // 路由服务映射
    6: map<string, i32> rids;
    // 登录返回
    7: binary data;
    // 清理队列
    8: bool clear
    // 登录回调
    9: bool back
}

struct Team {
    // 版本
    1: i64 version
    // 用户列表
    2: list<Member> members;
    // 读扩散、写扩散
    3: bool readFeed;
}

struct Member {
    // 用户编号
    1: string gid;
    // 写扩散时，不推送，需要点击查看
    2: bool nofeed;
}

// 访问控制
service Acl {
    // 登录
    Login login(1: i64 cid, 2: binary bytes);
    // 登录回调
    void loginBack(1: i64 cid, 2: i64 uid, 3: string sid);
    // 组查询
    Team team(1: string tid);
}

// 转发
service Pass {
    // 请求
    binary req(1: i64 cid, 2: i64 uid, 3: string sid, 4: string uri, 5: binary bytes);
    // 发送
    oneway void send(1: i64 cid, 2: i64 uid, 3: string sid, 4: string uri, 5: binary bytes);
}

// 网关
service Gateway {
    // 关闭连接
    bool close(1: i64 cid, 2: string reason)
    // 软关闭连接
    bool kick(1: i64 cid, 2: binary bytes)
    // 服务编号
    bool rid(1: i64 cid, 2: string name, 3: i32 rid)
    // 服务编号
    bool rids(1: i64 cid, 2: map<string, i32> rids)
    // 简单推送
    bool push(1: i64 cid, 2: string uri, 3: binary bytes)
    // 组推送 // uri 主题 // binary 消息体 // qs 消息质量，0 内存发送成功 1 队列发送[unique 唯一标识(消息队列，一个标识只需要最新数据)] 2 last队列 3 last 队列持久化
    bool gPush(1: string gid, 2: string uri, 3: binary bytes, 4: i32 qs, 5: string unique, 6: bool queue)
    // 注册监听gid
    bool gConn(1: i64 cid, 2: string gid, 3: string unique)
    // 断开监听gid
    bool gDisc(1: i64 cid, 2: string gid, 3: string unique, 4: i32 connVer)
    // 获取更新消息
    bool gLasts(1: string gid, 2: i64 cid, 3: string unique, 4: i64 lastId);
    // 点对点聊天
    bool send(1: string fromId, 2: string toId, 3: string uri, 4: binary bytes, 5: bool db)
    // readfeed读扩散，常用于聊天室
    bool tPush(1: string fromId, 2: string tid, 3: bool readfeed, 4: string uri, 5: binary bytes, 6: bool db, 7: bool queue)
    // 组更新、删除
    bool tDirty(1: string tid);
}