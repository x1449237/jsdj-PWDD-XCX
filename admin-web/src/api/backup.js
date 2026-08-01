import request from '@/utils/request'

// 备份列表
export function getBackupList(params) {
  return request({ url: '/backups', method: 'get', params })
}

// 创建备份
export function createBackup() {
  return request({ url: '/backups', method: 'post' })
}

// 恢复备份
export function restoreBackup(id) {
  return request({ url: `/backups/${id}/restore`, method: 'post' })
}

// 下载备份
export function downloadBackup(id) {
  return request({ url: `/backups/${id}/download`, method: 'get', responseType: 'blob' })
}
