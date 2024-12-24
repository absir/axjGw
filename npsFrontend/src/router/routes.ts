import type { Route } from '#/global'
import type { RouteRecordRaw } from 'vue-router'
import useSettingsStore from '@/store/modules/settings'
import generatedRoutes from 'virtual:generated-pages'
import { setupLayouts } from 'virtual:meta-layouts'

// 固定路由（默认路由）
const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login.vue'),
    meta: {
      title: '登录',
    },
  },
  {
    path: '/:all(.*)*',
    name: 'notFound',
    component: () => import('@/views/[...all].vue'),
    meta: {
      title: '找不到页面',
    },
  },
]

// 系统路由
const systemRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/index.vue'),
    meta: {
      title: () => useSettingsStore().settings.home.title,
      breadcrumb: false,
    },
    children: [
      {
        path: '',
        component: () => import('@/views/index.vue'),
        meta: {
          title: () => useSettingsStore().settings.home.title,
          icon: 'i-ant-design:home-twotone',
          breadcrumb: false,
        },
      },
      {
        path: 'reload',
        name: 'reload',
        component: () => import('@/views/reload.vue'),
        meta: {
          title: '重新加载',
          breadcrumb: false,
        },
      },
    ],
  },
]

// 动态路由（异步路由、导航栏路由）
const asyncRoutes: Route.recordMainRaw[] = [
  {
    meta: {
      title: 'yyNps',
      icon: 'i-uim:box',
    },
    children: [
      {
        path: '/client',
        component: () => import('@/layouts/index.vue'),
        meta: {
          title: '客户端',
          icon: 'arcticons:rdclient',
        },
        children: [
          {
            path: '',
            component: () => import('@/views/list/client.vue'),
            meta: {
              title: '客户端',
              menu: false,
            },
          },
        ],
      },
      {
        path: '/host',
        component: () => import('@/layouts/index.vue'),
        meta: {
          title: '域名',
          icon: 'stash:domain-light',
        },
        children: [
          {
            path: '',
            component: () => import('@/views/list/host.vue'),
            meta: {
              title: '域名',
              menu: false,
            },
          },
        ],
      },
      {
        path: '/tcp',
        component: () => import('@/layouts/index.vue'),
        meta: {
          title: 'TCP',
          icon: 'carbon:tcp-ip-service',
        },
        children: [
          {
            path: '',
            component: () => import('@/views/list/tcp.vue'),
            meta: {
              title: 'TCP',
              menu: false,
            },
          },
        ],
      },
    ],
  },
]

const constantRoutesByFilesystem = generatedRoutes.filter((item) => {
  return item.meta?.enabled !== false && item.meta?.constant === true
})

const asyncRoutesByFilesystem = setupLayouts(generatedRoutes.filter((item) => {
  return item.meta?.enabled !== false && item.meta?.constant !== true && item.meta?.layout !== false
}))

export {
  asyncRoutes,
  asyncRoutesByFilesystem,
  constantRoutes,
  constantRoutesByFilesystem,
  systemRoutes,
}
