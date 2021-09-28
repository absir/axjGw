namespace go gw

// 网关内部
service GatewayInner {
    // 服务编号
    void aid(1: i64 cid, 2: string name, 3: i32 aid)
    // 组更新、删除
    void dirty(1: string sid);
    // 推送 // uri 主题 // binary 消息体 // qs 消息质量，0 内存发送成功 1 持久发送成功 2 客户端ark // unique 唯一标识(消息队列，一个标识只需要最新数据)
    bool push(1: i64 cid, 2: i64 uid, 3: string sid, 4: string uri, 5: binary bytes, 6: i32 qs, 7: string unique);
}