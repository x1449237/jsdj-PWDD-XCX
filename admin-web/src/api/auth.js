import request from '@/utils/request'

// 登录
export function login(data) {
  return request({ url: '/login', method: 'post', data })
}

// 获取初始化状态
export function getInitStatus() {
  return request({ url: '/init/status', method: 'get' })
}

// 初始化-修改密码
export function changePassword(data) {
  return request({ url: '/init/change-password', method: 'post', data })
}

// 初始化-绑定邮箱
export function bindEmail(data) {
  return request({ url: '/init/bind-email', method: 'post', data })
}

// 初始化-发送邮箱验证码
export function sendVerifyCode(data) {
  return request({ url: '/init/send-verify-code', method: 'post', data })
}

// 初始化-校验邮箱验证码
export function verifyEmail(data) {
  return request({ url: '/init/verify-email', method: 'post', data })
}

// 忘记账号
export function forgotAccount(data) {
  return request({ url: '/forgot-account', method: 'post', data })
}

// 忘记密码
export function forgotPassword(data) {
  return request({ url: '/forgot-password', method: 'post', data })
}

// 获取当前管理员信息
export function getAdminInfo() {
  return request({ url: '/auth/info', method: 'get' })
}

// 登出
export function logout() {
  return request({ url: '/auth/logout', method: 'post' })
}
