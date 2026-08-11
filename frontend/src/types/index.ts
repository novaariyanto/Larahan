export type ComponentStatus = 'not_installed' | 'stopped' | 'running' | 'ready' | 'error' | 'installing'

export type ComponentType = 'apache' | 'php' | 'mysql' | 'phpmyadmin'

export interface ComponentInfo {
  type: ComponentType
  version: string
  status: ComponentStatus
  path: string
  installed: boolean
}

export interface PHPVersionInfo {
  version: string
  installed: boolean
  active: boolean
  path: string
}

export interface DashboardSummary {
  apache: ComponentInfo
  mysql: ComponentInfo
  active_php: string
  php: ComponentInfo
  phpmyadmin: ComponentInfo
}

export interface Settings {
  install_path: string
  active_php: string
  apache_port: number
  mysql_port: number
  first_run: boolean
}

export interface Paths {
  root: string
  apache: string
  php: string
  mysql: string
  phpmyadmin: string
  downloads: string
  logs: string
  config: string
  temp: string
}

export interface Result {
  ok: boolean
  message: string
}

export interface AppInfo {
  name: string
  version: string
  description: string
}

export function statusLabel(status: ComponentStatus): string {
  switch (status) {
    case 'running':
      return 'Running'
    case 'ready':
      return 'Ready'
    case 'stopped':
      return 'Stopped'
    case 'installing':
      return 'Installing'
    case 'error':
      return 'Error'
    default:
      return 'Not Installed'
  }
}
