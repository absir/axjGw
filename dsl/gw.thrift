namespace go gw

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
    // 登录回调
    8: bool back
}

struct Group {
    // 版本
    1: i64 version
    // 用户列表
    2: list<string> users;
    // 读扩散、写扩散
    3: bool readFeed;
}

// 访问控制
service Acl {
    // 登录
    Login login(1: i64 cid, 2: binary bytes);
    // 登录回调
    void loginBack(1: i64 cid, 2: i64 uid, 3: string sid);
    // 组查询
    Group group(1: string gid);
    // 挤掉线
    binary kickBs();
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
    void rid(1: i64 cid, 2: string name, 3: i32 rid)
    // 服务编号
    void rids(1: i64 cid, 2: map<string, i32> rids)
    // 注册监听
    void conn(1: i64 cid, 2: string sid)
    // 推送 // uri 主题 // binary 消息体 // qs 消息质量，0 内存发送成功 1 队列发送[unique 唯一标识(消息队列，一个标识只需要最新数据)] 2 last队列 3 last 队列持久化
    bool push(1: i64 cid, 2: i64 uid, 3: string sid, 4: string uri, 5: binary bytes, 6: i32 qs, 7: string unique);
    // 组更新、删除
    void dirty(1: string sid);
}