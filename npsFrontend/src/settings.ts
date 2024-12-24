import type { RecursiveRequired, Settings } from '#/global'
import settingsDefault from '@/settings.default'
import { defaultsDeep } from 'lodash-es'

const globalSettings: Settings.all = {
  app: {
    colorScheme: '',
  },
  menu: {
    mode: 'single',
    subMenuUniqueOpened: false,
    enableSubMenuCollapseButton: true,
  },
  tabbar: {
    enable: true,
    enableIcon: true,
  },
  toolbar: {
    fullscreen: true,
    pageReload: true,
    colorScheme: true,
  },
  copyright: {
    enable: true,
    dates: '2024',
    company: 'EmiyaGm',
    website: 'https://github.com/EmiyaGm',
  },
}

export default defaultsDeep(globalSettings, settingsDefault) as RecursiveRequired<Settings.all>
