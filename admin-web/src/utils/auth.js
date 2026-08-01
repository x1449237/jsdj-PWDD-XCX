const TOKEN_KEY = 'admin_token'
const NEED_INIT_KEY = 'admin_need_init'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
}

// 首次登录初始化标记（登录后需要修改密码/绑定邮箱）
export function getNeedInit() {
  return localStorage.getItem(NEED_INIT_KEY) === '1'
}

export function setNeedInit(value) {
  if (value) {
    localStorage.setItem(NEED_INIT_KEY, '1')
  } else {
    localStorage.removeItem(NEED_INIT_KEY)
  }
}
