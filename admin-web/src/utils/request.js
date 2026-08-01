import axios from 'axios'
import { ElMessage } from 'element-plus'
import { getToken, removeToken } from '@/utils/auth'

const instance = axios.create({
  baseURL: '/api/v1/admin',
  timeout: 30000
})

instance.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    config.headers['X-Trace-Id'] = generateTraceId()
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

instance.interceptors.response.use(
  (response) => {
    const res = response.data
    // 适配 Go(Gin) 后端统一响应格式：{code, msg, data, trace_id}
    if (res && typeof res === 'object' && 'code' in res) {
      if (res.code !== 200) {
        ElMessage.error(res.msg || '请求失败')
        if (res.code === 401) {
          removeToken()
          window.location.hash = '#/login'
        }
        return Promise.reject(new Error(res.msg || '请求失败'))
      }
      return res
    }
    // 非标准响应（如文件流）直接返回
    return res
  },
  (error) => {
    const { response } = error
    if (response) {
      const { status, data } = response
      if (status === 401) {
        removeToken()
        window.location.hash = '#/login'
        ElMessage.error('登录已过期，请重新登录')
      } else if (status === 403) {
        ElMessage.error('没有权限访问')
      } else if (status === 500) {
        ElMessage.error('服务器内部错误')
      } else {
        ElMessage.error(data?.msg || `请求错误 ${status}`)
      }
    } else {
      ElMessage.error('网络连接失败，请检查网络')
    }
    return Promise.reject(error)
  }
)

function generateTraceId() {
  return 'trace_' + Date.now() + '_' + Math.random().toString(36).substring(2, 11)
}

export default instance
