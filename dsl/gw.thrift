struct Login {
    // 数字编号
    i64 uid;
    // 字符编号
    string sid ;
    // 唯一标识(一个标识，只允许一个Conn)
    string unique;
    // 最大连接数
    i16 max;
}

struct Conn {
    // 数字编号
    i64 uid;
    // 字符编号
    string sid ;
}

struct Group {
    // 组连接
    list<Conn> conns;
    // 读扩散、写扩散
    bool read;
}

// 访问控制
service Acl {
    // 登录
    Login login(binary bytes);
    // 组查询
    Group group(string sid);
}

// 转发
service Pass {
    // 请求
    binary req(i64 uid, string sid, string uri, binary bytes);
    // 发送
    oneway void send(i64 uid, string sid, string uri, binary bytes);
}

struct Serv {
    // 服务名
    string name;
    // 服务地址(空,为反向调用服务)
    string addr;
    // 持久域(有值，注册后一直生效)
    string scope;
}

struct Msg {
    // 主题
    string uri;
    // 消息体
    binary bytes;
    // 消息质量，0 最低 1 不丢失 2 客户端ack
    i32 qs;
    // 唯一标识(消息队列，一个标识只需要最新数据)
    string unique;
}

// 网关
service Gateway {
    // 转发请求
    binary req(i64 uid, string sid, string uri, binary bytes);
    // 转发发送
    oneway void send(i64 uid, string sid, string uri, binary bytes);
    // 注册服务
    void reg(Serv serv);
    // 心跳包
    void beat();
    // 推送
    void push(i64 uid, string sid, Msg msg);
    // 组更新、删除
    void group(string sid, Group group, bool deleted);
}