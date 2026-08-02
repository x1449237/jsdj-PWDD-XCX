const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    // 基本信息
    realName: '',
    idCard: '',
    idCardFront: '',  // 人像面
    idCardBack: '',   // 国徽面
    validityType: '', // 有效期类型 non_permanent / permanent
    validityTypeText: '',
    startDate: '',    // 起始日期
    endDate: '',      // 截止日期

    // 协议勾选
    agreementChecked: false,

    // 活体检测
    livenessStatus: 'idle', // idle | scanning | success | fail
    livenessLoading: false,
    livenessFailReason: '',
    livenessModalVisible: false,

    // 未成年人
    isMinor: false,
    overLimit: false,
    showMinorTip: false,

    submitting: false
  },

  // 计算属性: 付费认证按钮是否可点
  computed: {
    canSubmit() {
      const d = this.data;
      const baseValid =
        d.idCardFront && d.idCardBack &&
        d.validityType &&
        d.realName && d.idCard &&
        d.startDate &&
        (d.validityType === 'permanent' || d.endDate) &&
        d.agreementChecked;
      return baseValid;
    }
  },

  onLoad(options) {
    if (options && options.overLimit) {
      this.setData({ overLimit: true });
    }
    // 小程序端 computed 需要通过 watcher 模拟,此处手动初始化一次
    this._updateCanSubmit();
  },

  // setData wrapper: 每次变更后更新 canSubmit
  setDataMerged(patch) {
    this.setData(patch, () => {
      this._updateCanSubmit();
    });
  },

  _updateCanSubmit() {
    const d = this.data;
    const baseValid =
      !!d.idCardFront && !!d.idCardBack &&
      !!d.validityType &&
      !!d.realName && !!d.idCard &&
      !!d.startDate &&
      (d.validityType === 'permanent' || !!d.endDate) &&
      !!d.agreementChecked;
    if (this.data.canSubmit !== baseValid) {
      // 直接 setData 避免递归
      this.data.canSubmit = baseValid;
      this.setData({ canSubmit: baseValid });
    } else {
      // 初始注入
      if (typeof this.data.canSubmit === 'undefined') {
        this.setData({ canSubmit: baseValid });
      }
    }
  },

  onBack() {
    wx.navigateBack({ delta: 1 }).catch(() => {
      wx.switchTab({ url: '/pages/index/index' });
    });
  },

  onNameInput(e) {
    this.setDataMerged({ realName: e.detail.value });
  },

  onIdCardInput(e) {
    this.setDataMerged({ idCard: e.detail.value });
  },

  onStartDateChange(e) {
    this.setDataMerged({ startDate: e.detail.value });
  },

  onEndDateChange(e) {
    this.setDataMerged({ endDate: e.detail.value });
  },

  onValidityChange() {
    wx.showActionSheet({
      itemList: ['非长期有效', '长期有效'],
      success: (res) => {
        const type = res.tapIndex === 0 ? 'non_permanent' : 'permanent';
        const text = res.tapIndex === 0 ? '非长期有效' : '长期有效';
        const patch = { validityType: type, validityTypeText: text };
        // 长期有效清空截止日期
        if (type === 'permanent') {
          patch.endDate = '';
        }
        this.setDataMerged(patch);
      }
    });
  },

  toggleAgreement() {
    this.setDataMerged({ agreementChecked: !this.data.agreementChecked });
  },

  // 上传身份证图片
  onUploadIdCard(e) {
    const type = e.currentTarget.dataset.type; // front | back
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        const tempFile = res.tempFiles[0];
        const tempPath = tempFile.tempFilePath;
        // 调用上传接口
        wx.showLoading({ title: '上传中', mask: true });
        wx.uploadFile({
          url: request.baseURL + '/upload/image',
          filePath: tempPath,
          name: 'file',
          header: request.getAuthHeader ? request.getAuthHeader() : {},
          success: (upRes) => {
            try {
              const data = JSON.parse(upRes.data);
              if (data.code === 0 || data.code === 200 || data.data) {
                const url = (data.data && data.data.url) || data.data || tempPath;
                const patch = {};
                patch[type === 'front' ? 'idCardFront' : 'idCardBack'] = url;
                this.setDataMerged(patch);
                wx.showToast({ title: '上传成功', icon: 'success' });
              } else {
                wx.showToast({ title: data.message || '上传失败', icon: 'none' });
              }
            } catch (e2) {
              // 上传失败时用本地路径(开发环境兼容)
              const patch = {};
              patch[type === 'front' ? 'idCardFront' : 'idCardBack'] = tempPath;
              this.setDataMerged(patch);
            }
          },
          fail: () => {
            // 上传失败兜底:本地路径
            const patch = {};
            patch[type === 'front' ? 'idCardFront' : 'idCardBack'] = tempPath;
            this.setDataMerged(patch);
          },
          complete: () => wx.hideLoading()
        });
      }
    });
  },

  // 点击付费认证:先校验必填 → 弹窗活体检测
  onPayAndAuth() {
    if (!this.data.canSubmit) {
      const missing = [];
      const d = this.data;
      if (!d.idCardFront) missing.push('身份证人像面');
      if (!d.idCardBack) missing.push('身份证国徽面');
      if (!d.validityType) missing.push('有效期类型');
      if (!d.realName) missing.push('用户姓名');
      if (!d.idCard) missing.push('证件号码');
      if (!d.startDate) missing.push('起始日期');
      if (d.validityType === 'non_permanent' && !d.endDate) missing.push('截止日期');
      if (!d.agreementChecked) missing.push('用户协议');
      wx.showToast({ title: '请完善' + missing.join('、'), icon: 'none' });
      return;
    }
    this.setData({
      livenessModalVisible: true,
      livenessStatus: 'idle',
      livenessFailReason: ''
    });
  },

  onCloseLiveness() {
    // 检测中不可关闭
    if (this.data.livenessStatus === 'scanning') return;
    this.setData({ livenessModalVisible: false });
  },

  // 开始活体检测(调用后端接口,先扣费再检测)
  onStartLiveness() {
    const { realName, idCard } = this.data;
    if (!realName || !idCard) {
      wx.showToast({ title: '请先填写姓名和证件号码', icon: 'none' });
      return;
    }
    this.setData({
      livenessStatus: 'scanning',
      livenessLoading: true,
      livenessFailReason: ''
    });

    // 调用后端活体检测接口(先收费再认证,失败不退款)
    request.post('/user/realname/face-verify', {
      real_name: realName,
      id_card: idCard
    }).then((res) => {
      this.setData({
        livenessStatus: 'success',
        livenessLoading: false
      });
    }).catch((err) => {
      // 余额不足时,关闭弹层,跳充值
      if (err.message && err.message.indexOf('余额不足') > -1) {
        this.setData({
          livenessModalVisible: false,
          livenessStatus: 'idle',
          livenessLoading: false
        });
        wx.showModal({
          title: '余额不足',
          content: '活体认证需 ¥2.00，请先充值后再进行认证。',
          confirmText: '去充值',
          success: (mRes) => {
            if (mRes.confirm) {
              wx.navigateTo({ url: '/pages/wallet/wallet' });
            }
          }
        });
        return;
      }
      this.setData({
        livenessStatus: 'fail',
        livenessLoading: false,
        livenessFailReason: (err && err.message) || '活体检测未通过，请重试'
      });
    });
  },

  // 活体通过后提交认证
  onSubmitAuth() {
    if (this.data.livenessStatus !== 'success') {
      wx.showToast({ title: '请先完成活体检测', icon: 'none' });
      return;
    }
    this.setData({ submitting: true });

    const d = this.data;
    const payload = {
      real_name: d.realName,
      id_card: d.idCard,
      id_card_front: d.idCardFront,
      id_card_back: d.idCardBack,
      validity_type: d.validityType,
      start_date: d.startDate,
      end_date: d.endDate
    };

    request.post('/user/realname/submit', payload).then((res) => {
      this.setData({
        submitting: false,
        livenessModalVisible: false
      });
      const isMinor = res.is_minor || false;
      const overLimit = res.over_limit || false;
      if (isMinor) {
        this.setData({
          isMinor: true,
          overLimit: overLimit,
          showMinorTip: true
        });
      } else {
        wx.showToast({ title: '实名认证成功', icon: 'success' });
        setTimeout(() => {
          wx.navigateBack();
        }, 1500);
      }
    }).catch((err) => {
      this.setData({ submitting: false });
      wx.showToast({ title: (err && err.message) || '认证失败，请重试', icon: 'none' });
    });
  },

  onCloseMinorTip() {
    this.setData({ showMinorTip: false });
    setTimeout(() => wx.navigateBack(), 500);
  },

  onGoGuardian() {
    this.setData({ showMinorTip: false });
    wx.navigateTo({ url: '/pages/guardian/guardian' });
  }
});
