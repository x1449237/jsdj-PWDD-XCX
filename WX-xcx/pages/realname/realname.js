const request = require('../../utils/request');

// 架构铁律:身份证校验位算法、姓名/身份证格式校验、年龄计算、活体认证逻辑
// 全部由 Go 后端执行。前端只负责:
//   1) 收集用户输入 → 调用后端 API
//   2) 渲染后端返回的字段
//   3) 纯 UI 反馈(进度条百分比、loading、toast)
// 前端禁止实现:校验位算法、正则格式校验、业务规则判断。

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

    // 实时错误提示(由后端 /thin/id-card/validate 异步返回)
    nameErr: '',
    idCardErr: '',

    // 协议勾选
    agreementChecked: false,

    // 完成度百分比(纯 UI 进度展示,非业务规则)
    completePercent: 0,

    // 当前步骤高亮(1/2/3,纯 UI 展示)
    currentStep: 1,

    // 活体检测(UI 状态机,非业务状态)
    livenessStatus: 'idle', // idle | scanning | success | fail
    livenessStatusText: '待检测',
    livenessLoading: false,
    livenessFailReason: '',
    livenessModalVisible: false,

    // 未成年人(由后端 submit 返回)
    isMinor: false,
    overLimit: false,
    showMinorTip: false,

    submitting: false
  },

  onLoad(options) {
    if (options && options.overLimit) {
      this.setData({ overLimit: true });
    }
    this._updateProgressUI();
  },

  // 纯 UI 进度展示(仅统计已填写字段数,不做业务规则判定)
  // 提交资格由后端 /user/realname/submit 权威校验
  _updateProgressUI() {
    const d = this.data;
    const items = [
      !!d.idCardFront ? 1 : 0,
      !!d.idCardBack ? 1 : 0,
      !!d.validityType ? 1 : 0,
      !!d.realName ? 1 : 0,
      !!d.idCard ? 1 : 0,
      !!d.startDate ? 1 : 0,
      !!d.agreementChecked ? 1 : 0
    ];
    let total = 7;
    if (d.validityType === 'non_permanent') {
      items.push(!!d.endDate ? 1 : 0);
      total = 8;
    }
    const sum = items.reduce((a, b) => a + b, 0);
    const percent = Math.round((sum / total) * 100);

    let currentStep = 1;
    if (sum >= 2 && d.idCardFront && d.idCardBack) currentStep = 2;
    if (sum === total) currentStep = 3;

    this.setData({ completePercent: percent, currentStep: currentStep });
  },

  // 包装 setData
  setDataMerged(patch) {
    this.setData(patch, () => {
      this._updateProgressUI();
    });
  },

  onBack() {
    wx.navigateBack({ delta: 1 }).catch(() => {
      wx.switchTab({ url: '/pages/index/index' });
    });
  },

  // 姓名输入:仅 setData,不做格式校验(校验由后端 submit 执行)
  onNameInput(e) {
    const val = (e.detail.value || '').trim();
    this.setDataMerged({ realName: val, nameErr: '' });
  },

  // 身份证输入:仅 setData,不做格式校验
  onIdCardInput(e) {
    const val = (e.detail.value || '').trim().toUpperCase();
    this.setDataMerged({ idCard: val, idCardErr: '' });
  },

  // 失焦时调后端做权威校验(校验位算法+年龄计算在后端)
  onIdCardBlur() {
    const val = (this.data.idCard || '').trim();
    if (val.length !== 18) return;
    request.post('/thin/id-card/validate', { id_card: val }).then((data) => {
      if (!data.valid) {
        this.setDataMerged({ idCardErr: data.message || '身份证号不正确' });
      } else {
        this.setDataMerged({ idCardErr: '' });
      }
    }).catch(() => {
      // 校验失败不阻断输入,最终以后端 submit 校验为准
    });
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
                wx.showToast({ title: '采集成功', icon: 'success' });
              } else {
                wx.showToast({ title: data.message || '上传失败', icon: 'none' });
              }
            } catch (e2) {
              wx.showToast({ title: '上传失败', icon: 'none' });
            }
          },
          fail: () => {
            wx.showToast({ title: '上传失败', icon: 'none' });
          },
          complete: () => wx.hideLoading()
        });
      }
    });
  },

  // 活体状态文字映射(纯 UI 状态机文案,非业务状态)
  _updateLivenessStatusText() {
    const s = this.data.livenessStatus;
    const map = { idle: '待检测', scanning: '检测中', success: '已通过', fail: '未通过' };
    this.setData({ livenessStatusText: map[s] || '待检测' });
  },

  // 点击付费认证:提示未填项(纯 UI 反馈),资格校验由后端执行
  onPayAndAuth() {
    const d = this.data;
    const missing = [];
    if (!d.idCardFront) missing.push('身份证人像面');
    if (!d.idCardBack) missing.push('身份证国徽面');
    if (!d.validityType) missing.push('有效期类型');
    if (!d.realName) missing.push('用户姓名');
    if (!d.idCard) missing.push('证件号码');
    if (!d.startDate) missing.push('起始日期');
    if (d.validityType === 'non_permanent' && !d.endDate) missing.push('截止日期');
    if (!d.agreementChecked) missing.push('用户协议');
    if (missing.length > 0) {
      wx.showToast({ title: '请完善' + missing.join('、'), icon: 'none' });
      return;
    }

    this.setData({
      livenessModalVisible: true,
      livenessStatus: 'idle',
      livenessFailReason: ''
    }, () => this._updateLivenessStatusText());
  },

  onCloseLiveness() {
    if (this.data.livenessStatus === 'scanning') return;
    this.setData({ livenessModalVisible: false });
  },

  // 开始活体检测(扣费+检测逻辑由后端执行,前端只触发并渲染状态)
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
    }, () => this._updateLivenessStatusText());

    request.post('/user/realname/face-verify', {
      real_name: realName,
      id_card: idCard
    }).then(() => {
      this.setData({
        livenessStatus: 'success',
        livenessLoading: false
      }, () => this._updateLivenessStatusText());
    }).catch((err) => {
      if (err.message && err.message.indexOf('余额不足') > -1) {
        this.setData({
          livenessModalVisible: false,
          livenessStatus: 'idle',
          livenessLoading: false
        }, () => this._updateLivenessStatusText());
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
      }, () => this._updateLivenessStatusText());
    });
  },

  // 活体通过后提交认证(所有业务校验由后端 /user/realname/submit 执行)
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

    request.post('/user/realname/submit', payload).then((data) => {
      this.setData({
        submitting: false,
        livenessModalVisible: false
      });
      const isMinor = data.is_minor || false;
      const overLimit = data.over_limit || false;
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
