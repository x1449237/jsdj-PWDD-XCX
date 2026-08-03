const auth = require('./utils/auth');
const websocket = require('./utils/websocket');
const request = require('./utils/request');

App({
  globalData: {
    userInfo: null,
    token: null,
    isLogin: false,
    systemInfo: null,
    baseURL: 'https://your-domain.com/api/v1',
    wsConnected: false,
    // ---- 99~582 需求配套全局缓存 ----
    clubJoinEnabled: true,                  // 入驻总开关 (216/433/650)
    badgeRenderConfigs: {},                 // V标渲染配置(196~215,212 禁止前端伪造)
    imUserPreference: null,                 // 用户 IM 个性化偏好
    shortPhraseList: [],                    // 快捷短语
    pendingSessionScrollPos: {}             // 会话列表位置记忆
  },

  onLaunch(options) {
    const that = this;
    this.getSystemInfo();
    this.checkUpdate();
    // ---- 99~582 初始化:公开接口,无需登录 ----
    this.preloadOpenConfigs();

    const token = auth.getToken();
    if (token) {
      this.globalData.token = token;
      this.globalData.isLogin = true;
      websocket.connect();
      this.checkAgreementVersions();
    }
  },

  // 99~582 全局预加载(入驻开关 + V标配置)
  async preloadOpenConfigs() {
    const that = this;
    // 入驻总开关
    request.get('/platform/join-switch').then(r => {
      that.globalData.clubJoinEnabled = !!r.enabled;
    }).catch(() => {});
    // V标渲染配置(禁止前端伪造 212,只按后端返回渲染)
    request.get('/badge-configs').then(r => {
      const cfg = {};
      (r.list || []).forEach(b => { cfg[b.badge_key] = b; });
      that.globalData.badgeRenderConfigs = cfg;
    }).catch(() => {});
  },

  onShow(options) {
    if (this.globalData.isLogin && !this.globalData.wsConnected) {
      websocket.connect();
    }
  },

  onHide() {
    websocket.close();
  },

  checkAgreementVersions() {
    const userInfo = wx.getStorageSync('user_info');
    if (!userInfo || !userInfo.role) return;

    const role = userInfo.role;
    const agreementTypes = ['user_service', 'privacy'];
    if (role === 'player') {
      agreementTypes.push('player_entry');
    } else if (role === 'club') {
      agreementTypes.push('club_entry');
    }

    let pendingAgreements = [];
    let checkedCount = 0;

    agreementTypes.forEach(type => {
      const signedKey = 'agreement_' + role + '_' + type + '_version';
      const signedVersion = wx.getStorageSync(signedKey);

      request.get('/compliance/agreement/latest', {
        role: role,
        agreement_type: type
      }).then(res => {
        if (res.version && res.version !== signedVersion) {
          pendingAgreements.push({
            type: type,
            version: res.version
          });
        }
        checkedCount++;
        if (checkedCount === agreementTypes.length && pendingAgreements.length > 0) {
          this.showAgreementSignPage(role, pendingAgreements[0].type);
        }
      }).catch(() => {
        checkedCount++;
      });
    });
  },

  showAgreementSignPage(role, type) {
    wx.navigateTo({
      url: '/pages/agreement/sign?role=' + role + '&type=' + type + '&force=1'
    });
  },

  getSystemInfo() {
    try {
      const systemInfo = wx.getSystemInfoSync();
      this.globalData.systemInfo = systemInfo;
      this.globalData.statusBarHeight = systemInfo.statusBarHeight;
      this.globalData.navBarHeight = systemInfo.platform === 'ios' ? 44 : 48;
      this.globalData.safeAreaBottom = systemInfo.screenHeight - systemInfo.safeArea.bottom;
    } catch (e) {
      console.error('获取系统信息失败:', e);
    }
  },

  checkUpdate() {
    if (wx.canIUse('getUpdateManager')) {
      const updateManager = wx.getUpdateManager();
      updateManager.onCheckForUpdate(function (res) {
        if (res.hasUpdate) {
          updateManager.onUpdateReady(function () {
            wx.showModal({
              title: '更新提示',
              content: '新版本已经准备好，是否重启应用？',
              success: function (res) {
                if (res.confirm) {
                  updateManager.applyUpdate();
                }
              }
            });
          });
          updateManager.onUpdateFailed(function () {
            wx.showModal({
              title: '更新提示',
              content: '新版本下载失败，请检查网络',
              showCancel: false
            });
          });
        }
      });
    }
  },

  setLoginState(token, userInfo) {
    this.globalData.token = token;
    this.globalData.userInfo = userInfo;
    this.globalData.isLogin = true;
    auth.setToken(token);
    websocket.connect();
  },

  logout() {
    this.globalData.token = null;
    this.globalData.userInfo = null;
    this.globalData.isLogin = false;
    auth.removeToken();
    websocket.close();
    wx.reLaunch({
      url: '/pages/login/login'
    });
  }
});