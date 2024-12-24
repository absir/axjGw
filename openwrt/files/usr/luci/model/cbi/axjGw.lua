local m, s, o
local sys = require "luci.sys"

m = Map("axjGw", "Axj 代理配置", "配置网关服务")

-- 运行状态部分
s = m:section(NamedSection, "settings", "axjGw", "配置")
local status = s:option(DummyValue, "_status", "当前状态")
status.rawhtml = true
local running = (sys.call("pgrep axjGw >/dev/null") == 0)
status.value = running and 
    [[<b><font color="green">运行中</font></b>]] or 
    [[<b><font color="red">已停止</font></b>]]

-- 原有配置部分

o = s:option(Value, "Proxy", "服务端地址")
o.default = "127.0.0.1:8774"

o = s:option(Value, "ClientKey", "客户端密钥")
o.default = "11111"

-- 添加日志查看按钮
s = m:section(NamedSection, "settings", "axjGw", "运行状态")
local logview = s:option(TextValue, "logview", "运行日志")
logview.rows = 20
logview.wrap = "off"
logview.cfgvalue = function(self, section)
    return sys.exec("cat /tmp/axjGw.log 2>/dev/null")
end
logview.readonly = "readonly"

local refresh = s:option(Button, "_refresh", "刷新日志")
refresh.inputstyle = "reload"
refresh.write = function()
    luci.http.redirect(luci.dispatcher.build_url("admin", "vpn", "axjGw"))
end

return m