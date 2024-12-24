include $(TOPDIR)/rules.mk

PKG_NAME:=axjGw
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk

define Package/axjGw
		SECTION:=utils
		CATEGORY:=Utilities
		TITLE:=AxjGw Agent Plugin
		DEPENDS:=+luci-base
		PKGARCH:=all
endef

define Package/axjGw/description
    	代理网关
endef

define Build/Compile
		echo "axjGw-agent $(PKG_VERSION)-$(PKG_RELEASE) `git rev-parse HEAD` `date +%s`" > $(PKG_BUILD_DIR)/version.txt
endef

define Package/axjGw/install
	  	$(INSTALL_DIR) $(1)/usr/bin
	  	$(INSTALL_BIN) ./files/usr/bin/axjGw $(1)/usr/bin/axjGw

	  	$(INSTALL_DIR) $(1)/etc/init.d
	  	$(INSTALL_BIN) ./files/etc/init.d/axjGw $(1)/etc/init.d/axjGw

	  	$(INSTALL_DIR) $(1)/etc/config
	  	$(INSTALL_CONF) ./files/etc/config/axjGw $(1)/etc/config/axjGw

	  	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/controller
	  	$(INSTALL_DATA) ./files/usr/luci/controller/*.lua $(1)/usr/lib/lua/luci/controller/

		$(INSTALL_DIR) $(1)/usr/lib/lua/luci/model/cbi/axjGw
		$(INSTALL_DATA) ./files/usr/luci/model/cbi/*.lua $(1)/usr/lib/lua/luci/model/cbi/

		$(INSTALL_DIR) $(1)/usr/lib/lua/luci/view
		$(INSTALL_DATA) ./files/usr/luci/view/*.htm $(1)/usr/lib/lua/luci/view/
endef

$(eval $(call BuildPackage,axjGw))