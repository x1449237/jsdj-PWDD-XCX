const TOKEN_KEY = 'token';
const USER_INFO_KEY = 'user_info';
const AGREEMENT_KEY = 'agreement_accepted';

const getToken = () => {
  try {
    return wx.getStorageSync(TOKEN_KEY) || '';
  } catch (e) {
    return '';
  }
};

const setToken = (token) => {
  try {
    wx.setStorageSync(TOKEN_KEY, token);
  } catch (e) {
    console.error('保存token失败:', e);
  }
};

const removeToken = () => {
  try {
    wx.removeStorageSync(TOKEN_KEY);
  } catch (e) {
    console.error('移除token失败:', e);
  }
};

const getStoredUserInfo = () => {
  try {
    return wx.getStorageSync(USER_INFO_KEY) || null;
  } catch (e) {
    return null;
  }
};

const setStoredUserInfo = (userInfo) => {
  try {
    wx.setStorageSync(USER_INFO_KEY, userInfo);
  } catch (e) {
    console.error('保存用户信息失败:', e);
  }
};

const isLogin = () => {
  const token = getToken();
  return !!token;
};

const isAgreementAccepted = () => {
  try {
    return wx.getStorageSync(AGREEMENT_KEY) || false;
  } catch (e) {
    return false;
  }
};

const acceptAgreement = () => {
  try {
    wx.setStorageSync(AGREEMENT_KEY, true);
  } catch (e) {
    console.error('保存协议同意状态失败:', e);
  }
};

const wxLogin = () => {
  return new Promise((resolve, reject) => {
    wx.login({
      success(res) {
        if (res.code) {
          resolve(res.code);
        } else {
          reject(new Error('登录失败'));
        }
      },
      fail(err) {
        reject(err);
      }
    });
  });
};

const getPhoneNumber = (e) => {
  return new Promise((resolve, reject) => {
    if (e.detail.errMsg === 'getPhoneNumber:ok') {
      resolve(e.detail);
    } else {
      reject(new Error('获取手机号失败'));
    }
  });
};

// 以下接口对接 Go 后端 /api/v1，懒加载 request 以避免与 request.js 形成循环依赖
const loginWithWx = (phoneDetail) => {
  // phoneDetail: { code, encryptedData, iv }，code 由 wxLogin() 获取
  const request = require('./request');
  return wxLogin().then((code) => {
    return request.post('/auth/wx-login', {
      code: code,
      encrypted_data: phoneDetail.encryptedData,
      iv: phoneDetail.iv
    });
  }).then((res) => {
    if (res && res.token) {
      setToken(res.token);
      if (res.user_info) {
        setStoredUserInfo(res.user_info);
      }
    }
    return res;
  });
};

const register = (data) => {
  // data: { nickname, avatar, invite_code, role }
  const request = require('./request');
  return request.post('/auth/register', data).then((res) => {
    if (res && res.user_info) {
      setStoredUserInfo(res.user_info);
    }
    return res;
  });
};

const getRealnameStatus = () => {
  const request = require('./request');
  return request.get('/user/realname/status');
};

const auth = {
  getToken,
  setToken,
  removeToken,
  getStoredUserInfo,
  setStoredUserInfo,
  isLogin,
  isAgreementAccepted,
  acceptAgreement,
  wxLogin,
  getPhoneNumber,
  loginWithWx,
  register,
  getRealnameStatus
};

module.exports = auth;