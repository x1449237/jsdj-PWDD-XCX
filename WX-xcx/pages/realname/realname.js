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

    // 实时错误提示
    nameErr: '',
    idCardErr: '',

    // 协议勾选
    agreementChecked: false,

    // 按钮可用
    canSubmit: false,

    // 完成度百分比(0-100)
    completePercent: 0,

    // 当前步骤高亮(1/2/3)
    currentStep: 1,

    // 活体检测
    livenessStatus: 'idle', // idle | scanning | success | fail
    livenessStatusText: '待检测',
    livenessLoading: false,
    livenessFailReason: '',
    livenessModalVisible: false,

    // 未成年人
    isMinor: false,
    overLimit: false,
    showMinorTip: false,

    submitting: false
  },

  onLoad(options) {
    if (options && options.overLimit) {
      this.setData({ overLimit: true });
    }
    this._updateAllDerived();
  },

  // 统一更新所有派生状态
  _updateAllDerived() {
    const d = this.data;

    // 1. canSubmit
    const baseValid =
      !!d.idCardFront && !!d.idCardBack &&
      !!d.validityType &&
      !!d.realName && !!d.idCard &&
      !!d.startDate &&
      (d.validityType === 'permanent' || !!d.endDate) &&
      !!d.agreementChecked;

    // 2. completePercent: 逐项计分,共 7 项(人像/国徽/有效期/姓名/身份证/起始/协议) + 截止日期(非长期时)
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

    // 3. currentStep: 按进度高亮
    let currentStep = 1;
    if (sum >= 2 && d.idCardFront && d.idCardBack) currentStep = 2;
    if (baseValid) currentStep = 3;

    // 4. 批量 setData
    const patch = {
      canSubmit: baseValid,
      completePercent: percent,
      currentStep: currentStep
    };
    this.setData(patch);
  },

  // 包装 setData
  setDataMerged(patch) {
    this.setData(patch, () => {
      this._updateAllDerived();
    });
  },

  onBack() {
    wx.navigateBack({ delta: 1 }).catch(() => {
      wx.switchTab({ url: '/pages/index/index' });
    });
  },

  onNameInput(e) {
    const val = (e.detail.value || '').trim();
    let err = '';
    if (val.length > 0) {
      // 姓名: 中文或少数名中间点, 2-20 位
      const reg = /^[\u4e00-\u9fa5·]{2,20}$/;
      if (!reg.test(val)) {
        err = '请输入2-20位中文姓名';
      }
    }
    this.setDataMerged({ realName: val, nameErr: err });
  },

  onIdCardInput(e) {
    const val = (e.detail.value || '').trim().toUpperCase();
    let err = '';
    if (val.length > 0 && val.length !== 18) {
      err = '身份证号需为18位';
    } else if (val.length === 18) {
      // 基础格式校验
      const reg = /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dX]$/;
      if (!reg.test(val)) {
        err = '身份证号格式不正确';
      }
    }
    this.setDataMerged({ idCard: val, idCardErr: err });
  },

  // 失焦时做更严格的校验码检查
  onIdCardBlur() {
    const d = this.data;
    const val = (d.idCard || '').trim();
    if (val.length !== 18) return;
    // 校验码计算
    const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2];
    const codes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'];
    let sum = 0;
    for (let i = 0; i < 17; i++) sum += parseInt(val[i], 10) * weights[i];
    const expected = codes[sum % 11];
    if (val[17] !== expected) {
      this.setDataMerged({ idCardErr: '身份证校验码错误，请检查' });
    }
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
              const patch = {};
              patch[type === 'front' ? 'idCardFront' : 'idCardBack'] = tempPath;
              this.setDataMerged(patch);
            }
          },
          fail: () => {
            const patch = {};
            patch[type === 'front' ? 'idCardFront' : 'idCardBack'] = tempPath;
            this.setDataMerged(patch);
          },
          complete: () => wx.hideLoading()
        });
      }
    });
  },

  // 活体状态文字映射
  _updateLivenessStatusText() {
    const s = this.data.livenessStatus;
    const map = { idle: '待检测', scanning: '检测中', success: '已通过', fail: '未通过' };
    this.setData({ livenessStatusText: map[s] || '待检测' });
  },

  // 点击付费认证:先校验必填 → 弹活体检测
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

    // 二次校验身份证校验码错误
    if (this.data.idCardErr) {
      wx.showToast({ title: '请先修正证件号码错误', icon: 'none' });
      return;
    }
    if (this.data.nameErr) {
      wx.showToast({ title: '请先修正姓名错误', icon: 'none' });
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

  // 开始活体检测(先扣费再检测,失败不退款)
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
    }).then((res) => {
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
