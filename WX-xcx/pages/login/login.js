const auth = require('../../utils/auth');
const request = require('../../utils/request');
const app = getApp();

const AGREEMENT_KEYS = {
  PLAYER: 'agreeAgreementPlayer',
  PRIVACY: 'agreePrivacy',
  BOOSTER: 'agreeAgreementBooster',
  DISTRIBUTOR: 'agreeDistributor',
  CLUB: 'agreeAgreementClub'
};

Page({
  data: {
    agreed: false,
    loading: false,
    showAgreementModal: false,
    isFirstLogin: true,
    agreeAgreementPlayer: false,
    agreePrivacy: false,
    agreeAgreementBooster: false,
    agreeDistributor: false,
    agreeAgreementClub: false,
    hasSignedCache: {},
    roleTabs: [
      { key: 'player', label: '用户', tip: '普通下单玩家必填', required: true },
      { key: 'booster', label: '打手', tip: '想接单成为大神必看', required: false },
      { key: 'distributor', label: '分销商', tip: '推广分销赚钱必看', required: false },
      { key: 'club', label: '俱乐部', tip: '俱乐部入驻必看', required: false }
    ],
    activeRoleTab: 'player',
    documents: {}
  },

  onLoad(options) {
    if (auth.isLogin()) {
      wx.switchTab({
        url: '/pages/index/index'
      });
      return;
    }

    const cache = this.loadAgreementCache();
    const defaultChecked = cache[AGREEMENT_KEYS.PLAYER] && cache[AGREEMENT_KEYS.PRIVACY];

    this.setData({
      showAgreementModal: !auth.isAgreementAccepted(),
      hasSignedCache: cache,
      agreeAgreementPlayer: !!cache[AGREEMENT_KEYS.PLAYER],
      agreePrivacy: !!cache[AGREEMENT_KEYS.PRIVACY],
      agreeAgreementBooster: !!cache[AGREEMENT_KEYS.BOOSTER],
      agreeDistributor: !!cache[AGREEMENT_KEYS.DISTRIBUTOR],
      agreeAgreementClub: !!cache[AGREEMENT_KEYS.CLUB],
      agreed: !!defaultChecked
    });

    this.loadDocuments();
  },

  loadDocuments() {
    request.get('/documents/categories').then((res) => {
      const docs = res.data?.list || [];
      const documentMap = {};
      docs.forEach(d => {
        const key = (d.category || d.type || '').toLowerCase();
        documentMap[key] = d;
      });
      this.setData({ documents: documentMap });
    }).catch(() => {});
  },

  loadAgreementCache() {
    try {
      const cache = wx.getStorageSync('agreement_sign_cache');
      return cache ? JSON.parse(cache) : {};
    } catch (e) {
      return {};
    }
  },

  saveAgreementCache(patch) {
    try {
      const current = this.loadAgreementCache();
      const merged = Object.assign({}, current, patch);
      wx.setStorageSync('agreement_sign_cache', JSON.stringify(merged));
      this.setData({ hasSignedCache: merged });
    } catch (e) {}
  },

  onSwitchRoleTab(e) {
    const key = e.currentTarget.dataset.key;
    this.setData({ activeRoleTab: key });
  },

  onToggleAgreementItem(e) {
    const { field } = e.currentTarget.dataset;
    const newValue = !this.data[field];
    this.setData({ [field]: newValue });

    if (newValue) {
      this.reportAgreementLog(field, true);
    }

    this.saveAgreementCache({ [field]: newValue });

    const { agreeAgreementPlayer, agreePrivacy } = this.data;
    const allRequired = agreeAgreementPlayer && agreePrivacy;
    this.setData({ agreed: allRequired });
  },

  reportAgreementLog(field, isAgree) {
    const roleMap = {
      agreeAgreementPlayer: 'player',
      agreePrivacy: 'privacy',
      agreeAgreementBooster: 'booster',
      agreeDistributor: 'distributor',
      agreeAgreementClub: 'club'
    };
    const agreementType = roleMap[field] || field;
    const signature = `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;

    request.post('/agreement/sign-logs', {
      agreement_type: agreementType,
      agreement_field: field,
      is_agree: isAgree,
      signature: signature,
      source: 'wx_login',
      user_agent: 'miniprogram'
    }).catch(() => {});
  },

  onToggleAgreement() {
    if (this.data.agreed) {
      this.setData({
        agreed: false,
        agreeAgreementPlayer: false,
        agreePrivacy: false
      });
      this.saveAgreementCache({
        [AGREEMENT_KEYS.PLAYER]: false,
        [AGREEMENT_KEYS.PRIVACY]: false
      });
    } else {
      this.setData({
        showAgreementModal: true
      });
    }
  },

  onShowAgreement(e) {
    const { type = 'service', doc = '' } = e.currentTarget.dataset;
    const targetDoc = this.data.documents[doc] || null;
    let title = '用户服务协议';
    let content = '用户服务协议详情...';

    if (type === 'privacy') {
      title = '隐私政策';
      content = '隐私政策详情...';
    } else if (type === 'booster') {
      title = '打手入驻协议';
      content = '打手入驻协议详情...';
    } else if (type === 'distributor') {
      title = '分销商协议';
      content = '分销商协议详情...';
    } else if (type === 'club') {
      title = '俱乐部合作协议';
      content = '俱乐部合作协议详情...';
    }

    if (targetDoc && targetDoc.file_url) {
      wx.downloadFile({
        url: targetDoc.file_url,
        success: (res) => {
          if (res.statusCode === 200) {
            wx.openDocument({
              filePath: res.tempFilePath,
              showMenu: true,
              fail: () => {
                wx.showModal({ title, content, showCancel: false });
              }
            });
          } else {
            wx.showModal({ title, content, showCancel: false });
          }
        },
        fail: () => {
          wx.showModal({ title, content, showCancel: false });
        }
      });
    } else {
      wx.showModal({ title, content, showCancel: false });
    }
  },

  onGetPhoneNumber(e) {
    if (!this.data.agreeAgreementPlayer || !this.data.agreePrivacy) {
      wx.showToast({
        title: '请先勾选用户协议与隐私政策',
        icon: 'none'
      });
      return;
    }

    auth.getPhoneNumber(e).then((detail) => {
      return this.doLogin(detail);
    }).catch((err) => {
      console.error('获取手机号失败:', err);
      wx.showToast({
        title: '获取手机号失败，请重试',
        icon: 'none'
      });
    });
  },

  doLogin(phoneDetail) {
    this.setData({ loading: true });

    auth.loginWithWx(phoneDetail).then((res) => {
      let userInfo = {};
      if (res.user_info) {
        userInfo = res.user_info;
      }

      app.setLoginState(res.token, userInfo);

      if (res.is_new_user) {
        wx.redirectTo({
          url: '/pages/register/register'
        });
      } else {
        wx.switchTab({
          url: '/pages/index/index'
        });
      }
    }).catch((err) => {
      console.error('登录失败:', err);
      this.setData({ loading: false });
    });
  },

  onVisitorLogin() {
    wx.switchTab({
      url: '/pages/index/index'
    });
  },

  onCloseAgreementModal() {
  },

  onAcceptAgreement() {
    auth.acceptAgreement();
    const patch = {
      [AGREEMENT_KEYS.PLAYER]: true,
      [AGREEMENT_KEYS.PRIVACY]: true
    };
    this.saveAgreementCache(patch);
    this.reportAgreementLog(AGREEMENT_KEYS.PLAYER, true);
    this.reportAgreementLog(AGREEMENT_KEYS.PRIVACY, true);
    this.setData({
      showAgreementModal: false,
      agreed: true,
      agreeAgreementPlayer: true,
      agreePrivacy: true
    });
  },

  onRejectAgreement() {
    wx.showModal({
      title: '提示',
      content: '需要同意协议才能使用服务',
      showCancel: false,
      success: () => {
        wx.exitMiniProgram();
      }
    });
  }
});