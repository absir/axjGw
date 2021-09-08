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
}

struct Conn {
    // 数字编号
    1: i64 uid;
    // 字符编号
    2: string sid;
}

struct Group {
    // 组连接
    1: list<Conn> conns;
    // 读扩散、写扩散
    2: bool readFeed;
}

// 访问控制
service Acl {
    // 登录
    Login login(1: string cid, 2: binary bytes);
    // 组查询
    Group group(1: string sid);
}

// 转发
service Pass {
    // 请求
    binary req(1: i64 uid, 2: string sid, 3: string uri, 4: binary bytes);
    // 发送
    oneway void send(1: i64 uid, 2: string sid, 3: string uri, 4: binary bytes);
}

struct Serv {
    // 服务名
    1: string name;
    // 服务地址(空,为反向调用服务)
    2: string addr;
    // 持久域(有值，注册后一直生效)
    3: string scope;
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
    // 转发请求
    binary req(1: string cid, 2: string uri, 3: binary bytes);
    // 转发发送
    oneway void send(1: string cid, 2: string uri, 3: binary bytes);
    // 注册服务
    void reg(1: Serv serv);
    // 心跳包
    void beat();
    // 推送
    void push(1: string id, 2: Msg msg);
    // 组更新、删除
    void group(1: string sid, 2: Group group, 3: bool deleted);
}