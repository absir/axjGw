module("luci.controller.axjGw", package.seeall)

function index()
    entry({"admin", "vpn", "axjGw"}, cbi("axjGw"), _("Axj 代理配置"), 10).dependent = true
end