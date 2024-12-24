import type { Menu } from '#/global'

// import MultilevelMenuExample from './modules/multilevel.menu.example'

const menu: Menu.recordMainRaw[] = [
  {
    meta: {
      title: 'yyNps',
      icon: 'uim:box',
    },
    children: [
      {
        meta: {
          title: '客户端',
        },
        path: '/client',
      },
    ],
  },
]

export default menu
