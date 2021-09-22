namespace go gw

struct Login {
    // 数字编号
    1: i64 uid;
    // 字符编号
    2: string sid;
    // 唯一标识(一个标识，只允许一个Conn)
    3: string unique;
    // 最大连接数
    4: i16 max;
    // 服务编号
    5: i32 aid;
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

struct Serv {
    // 服务编号
    1: i32 sid;
    // 服务地址(空,为反向调用服务)
    2: string addr;
    // 服务名
    3: string name;
}

struct Msg {
    // 主题
    1: string uri;
    // 消息体
    2: binary bytes;
    // 消息质量，0 内存发送成功 1 持久发送成功 2 客户端ark
    3: i32 qs;
    // 唯一标识(消息队列，一个标识只需要最新数据)
    4: string unique;
}

// 网关
service Gateway {
    // 服务编号
    void aid(1: i64 cid, 2: string name, 3: i32 aid)
    // 组更新、删除
    void dirty(1: string sid);
    // 推送
    bool push(1: i64 cid, 2: i64 uid, 3: string sid, 4: Msg msg);
}