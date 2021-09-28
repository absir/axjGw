namespace go gw

struct Login {
    // 数字编号
    1: i64 uid;
    // 字符编号
    2: string sid;
    // 唯一标识(一个标识，只允许一个Conn)
    3: string unique;
    // 服务编号
    4: i32 aid;
    // 最大请求数
    5: i32 poolG;
}

struct Conn {
    // 数字编号
    1: i64 uid;
    // 字符编号
    2: string sid;
}

struct Group {
    // 版本
    1: i64 version
    // 组连接
    2: list<Conn> conns;
    // 读扩散、写扩散
    3: bool readFeed;
}

// 访问控制
service Acl {
    // 登录
    Login login(1: i64 cid, 2: binary bytes);
    // 组查询
    Group group(1: string sid);
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
    // 服务编号
    void aid(1: i64 cid, 2: string name, 3: i32 aid)
    // 组更新、删除
    void dirty(1: string sid);
    // 推送 // uri 主题 // binary 消息体 // qs 消息质量，0 内存发送成功 1 持久发送成功 2 客户端ark // unique 唯一标识(消息队列，一个标识只需要最新数据)
    bool push(1: i64 cid, 2: i64 uid, 3: string sid, 4: string uri, 5: binary bytes, 6: i32 qs, 7: string unique);
}