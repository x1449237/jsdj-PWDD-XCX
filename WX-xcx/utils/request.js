const auth = require('./auth');

// Go(Gin) 后端 API 前缀，所有业务路径均相对该前缀
const baseURL = 'https://your-domain.com/api/v1';

let requestCount = 0;

const generateTraceId = () => {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 10);
  const seq = (++requestCount).toString(36).padStart(4, '0');
  return `${timestamp}${random}${seq}`;
};

const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = auth.getToken();
    const traceId = generateTraceId();

    const header = {
      'Content-Type': 'application/json',
      'X-Trace-Id': traceId,
      ...options.header
    };

    if (token) {
      header['Authorization'] = `Bearer ${token}`;
    }

    wx.request({
      url: `${baseURL}${options.url}`,
      method: options.method || 'GET',
      data: options.data || {},
      header: header,
      timeout: options.timeout || 15000,
      success(res) {
        const { statusCode, data } = res;

        // Go 后端统一响应格式: {code: 200, msg: "成功", data: {}, trace_id: "uuid"}
        if (statusCode === 200 && data && data.code === 200) {
          resolve(data.data);
        } else if (statusCode === 401 || (data && data.code === 401)) {
          auth.removeToken();
          const app = getApp();
          if (app && app.globalData) {
            app.globalData.isLogin = false;
            app.globalData.token = null;
          }
          wx.reLaunch({
            url: '/pages/login/login'
          });
          reject(data || { code: 401, msg: '登录已过期' });
        } else {
          wx.showToast({
            title: (data && (data.msg || data.message)) || '请求失败',
            icon: 'none',
            duration: 2000
          });
          reject(data || { code: statusCode, msg: '请求失败' });
        }
      },
      fail(err) {
        wx.showToast({
          title: '网络异常，请检查网络',
          icon: 'none',
          duration: 2000
        });
        reject(err);
      }
    });
  });
};

const get = (url, data = {}, options = {}) => {
  return request({ url, data, method: 'GET', ...options });
};

const post = (url, data = {}, options = {}) => {
  return request({ url, data, method: 'POST', ...options });
};

const put = (url, data = {}, options = {}) => {
  return request({ url, data, method: 'PUT', ...options });
};

const del = (url, data = {}, options = {}) => {
  return request({ url, data, method: 'DELETE', ...options });
};

// 文件上传封装，路径相对 baseURL（不含 /api/v1 前缀），resolve(data.data)
const upload = (url, filePath, formData = {}, options = {}) => {
  return new Promise((resolve, reject) => {
    const token = auth.getToken();
    const header = {
      'X-Trace-Id': generateTraceId(),
      ...options.header
    };
    if (token) {
      header['Authorization'] = `Bearer ${token}`;
    }

    wx.uploadFile({
      url: `${baseURL}${url}`,
      filePath: filePath,
      name: options.name || 'file',
      formData: formData,
      header: header,
      timeout: options.timeout || 60000,
      success(res) {
        try {
          const data = JSON.parse(res.data);
          if (data.code === 200) {
            resolve(data.data);
          } else if (data.code === 401) {
            auth.removeToken();
            const app = getApp();
            if (app && app.globalData) {
              app.globalData.isLogin = false;
              app.globalData.token = null;
            }
            wx.reLaunch({ url: '/pages/login/login' });
            reject(data);
          } else {
            wx.showToast({
              title: data.msg || '上传失败',
              icon: 'none',
              duration: 2000
            });
            reject(data);
          }
        } catch (e) {
          wx.showToast({ title: '上传失败', icon: 'none', duration: 2000 });
          reject(e);
        }
      },
      fail(err) {
        wx.showToast({
          title: '网络异常，请检查网络',
          icon: 'none',
          duration: 2000
        });
        reject(err);
      }
    });
  });
};

module.exports = {
  request,
  get,
  post,
  put,
  del,
  upload,
  baseURL,
  // 兼容旧代码中 request.getBaseUrl() 的调用
  getBaseUrl: () => baseURL
};
