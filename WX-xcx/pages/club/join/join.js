const request = require('../../../utils/request');

Page({
  data: {
    step: 1,
    totalSteps: 7,
    clubType: '',          // 'green_v' 个人 / 'blue_v' 企业
    isEnterprise: false,
    clubJoinOpen: true,
    personalDeposit: 0,
    enterpriseDeposit: 0,

    // 步骤1：须知
    agreed: false,

    // 步骤2：基础资料
    clubName: '',
    abbreviation: '',
    abbrOccupied: false,
    abbrLoading: false,
    realName: '',
    idCard: '',
    phone: '',
    addressProvince: '',
    addressCity: '',
    addressDistrict: '',
    addressDetail: '',

    // 步骤3：活体认证
    idCardFront: '',
    idCardBack: '',
    livenessStatus: 0,

    // 步骤4：合同
    contractFile: '',

    // 步骤5：预览
    // 企业专属
    businessLicense: '',
    corporateBank: '',
    corporateAccount: '',
    handleType: 'self',    // self / agent
    agentName: '',
    agentIdCard: '',
    agentIdCardFront: '',
    agentIdCardBack: '',
    agentAuthorization: '',

    // 提交
    submitting: false
  },

  onLoad(options) {
    const type = options.type || 'green_v';
    this.setData({
      clubType: type,
      isEnterprise: type === 'blue_v',
      totalSteps: type === 'blue_v' ? 7 : 7
    });
    this.checkSwitch();
  },

  async checkSwitch() {
    try {
      const res = await request.get('/club/check_switch');
      const isOpen = res.data?.club_join_open === true;
      if (!isOpen) {
        wx.showModal({
          title: '提示',
          content: '平台暂时关闭俱乐部入驻通道',
          showCancel: false,
          success: () => wx.navigateBack()
        });
      }
      this.setData({
        clubJoinOpen: isOpen,
        personalDeposit: res.data?.personal_deposit || 0,
        enterpriseDeposit: res.data?.enterprise_deposit || 0
      });
    } catch (e) {
      wx.showToast({ title: '网络异常', icon: 'none' });
    }
  },

  // 步骤切换
  nextStep() {
    const { step, clubJoinOpen } = this.data;
    if (!clubJoinOpen) return;

    // 每步校验
    if (step === 1 && !this.data.agreed) {
      wx.showToast({ title: '请先勾选同意全部协议', icon: 'none' });
      return;
    }
    if (step === 2) {
      if (!this.data.clubName.trim()) { wx.showToast({ title: '请填写俱乐部名称', icon: 'none' }); return; }
      if (this.data.abbrOccupied) { wx.showToast({ title: '缩写被占用，请更换名称', icon: 'none' }); return; }
      if (!this.data.realName) { wx.showToast({ title: '请填写真实姓名', icon: 'none' }); return; }
      if (!this.data.idCard || this.data.idCard.length !== 18) { wx.showToast({ title: '请填写18位身份证号', icon: 'none' }); return; }
      if (!this.data.phone || this.data.phone.length !== 11) { wx.showToast({ title: '请填写11位手机号', icon: 'none' }); return; }
    }
    if (step === 3) {
      if (!this.data.idCardFront) { wx.showToast({ title: '请上传身份证正面', icon: 'none' }); return; }
      if (!this.data.idCardBack) { wx.showToast({ title: '请上传身份证反面', icon: 'none' }); return; }
    }
    if (step === 4) {
      if (!this.data.contractFile) { wx.showToast({ title: '请上传已签署合同', icon: 'none' }); return; }
    }

    this.setData({ step: step + 1 });
  },

  prevStep() {
    if (this.data.step > 1) {
      this.setData({ step: this.data.step - 1 });
    }
  },

  // 步骤1：勾选协议
  toggleAgree() {
    this.setData({ agreed: !this.data.agreed });
  },

  // 步骤2：俱乐部名称输入 → 实时生成缩写
  onClubNameInput(e) {
    const name = e.detail.value;
    this.setData({ clubName: name });
    if (name.length >= 2) {
      this.generateAbbr(name);
    } else {
      this.setData({ abbreviation: '', abbrOccupied: false });
    }
  },

  async generateAbbr(name) {
    this.setData({ abbrLoading: true });
    try {
      const res = await request.post('/club/generate_abbr', { club_name: name });
      this.setData({
        abbreviation: res.data?.abbreviation || '',
        abbrOccupied: res.data?.occupied || false,
        abbrLoading: false
      });
    } catch (e) {
      this.setData({ abbrLoading: false });
    }
  },

  onAbbrHelp() {
    wx.navigateTo({ url: '/pages/club/abbr-help/abbr-help' });
  },

  // 通用输入
  onInput(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ [field]: e.detail.value });
  },

  // 上传图片
  onUploadImage(e) {
    const field = e.currentTarget.dataset.field;
    wx.chooseImage({
      count: 1,
      sizeType: ['compressed'],
      sourceType: ['camera', 'album'],
      success: (res) => {
        wx.showLoading({ title: '上传中...' });
        wx.uploadFile({
          url: request.baseURL + '/upload/image',
          filePath: res.tempFilePaths[0],
          name: 'file',
          header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
          success: (uploadRes) => {
            wx.hideLoading();
            const data = JSON.parse(uploadRes.data);
            if (data.code === 200) {
              this.setData({ [field]: data.data.url });
            } else {
              wx.showToast({ title: '上传失败', icon: 'none' });
            }
          },
          fail: () => {
            wx.hideLoading();
            wx.showToast({ title: '上传失败', icon: 'none' });
          }
        });
      }
    });
  },

  // 活体认证
  onLivenessCheck() {
    wx.showToast({ title: '活体认证中...', icon: 'loading' });
    // 调用微信活体认证API
    setTimeout(() => {
      this.setData({ livenessStatus: 1 });
      wx.showToast({ title: '认证通过', icon: 'success' });
    }, 2000);
  },

  // 上传合同
  onUploadContract() {
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf'],
      success: (res) => {
        wx.showLoading({ title: '上传中...' });
        wx.uploadFile({
          url: request.baseURL + '/upload/file',
          filePath: res.tempFiles[0].path,
          name: 'file',
          header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
          success: (uploadRes) => {
            wx.hideLoading();
            const data = JSON.parse(uploadRes.data);
            if (data.code === 200) {
              this.setData({ contractFile: data.data.url });
            } else {
              wx.showToast({ title: '上传失败', icon: 'none' });
            }
          },
          fail: () => {
            wx.hideLoading();
            wx.showToast({ title: '上传失败', icon: 'none' });
          }
        });
      }
    });
  },

  // 预览合同
  onPreviewContract() {
    if (!this.data.contractFile) return;
    wx.downloadFile({
      url: this.getBaseUrl() + this.data.contractFile,
      success: (res) => {
        wx.openDocument({ filePath: res.tempFilePath, fileType: 'pdf' });
      }
    });
  },

  // 下载合同模板
  onDownloadTemplate() {
    wx.showToast({ title: '下载合同模板中...', icon: 'loading' });
    // 实际应调用后端接口下载对应类型的合同模板
  },

  // 步骤5：提交
  async onSubmit() {
    const { clubName, clubType, isEnterprise, realName, idCard, phone, contractFile } = this.data;

    this.setData({ submitting: true });

    const postData = {
      club_name: clubName,
      club_type: clubType,
      real_name: realName,
      id_card: idCard,
      phone: phone,
      address_province: this.data.addressProvince,
      address_city: this.data.addressCity,
      address_district: this.data.addressDistrict,
      address_detail: this.data.addressDetail,
      id_card_front: this.data.idCardFront,
      id_card_back: this.data.idCardBack,
      liveness_status: this.data.livenessStatus,
      contract_file: contractFile,
    };

    if (isEnterprise) {
      Object.assign(postData, {
        business_license: this.data.businessLicense,
        corporate_bank: this.data.corporateBank,
        corporate_account: this.data.corporateAccount,
        handle_type: this.data.handleType,
        agent_name: this.data.agentName,
        agent_id_card: this.data.agentIdCard,
        agent_id_card_front: this.data.agentIdCardFront,
        agent_id_card_back: this.data.agentIdCardBack,
        agent_authorization: this.data.agentAuthorization,
      });
    }

    try {
      await request.post('/club/submit', postData);
      wx.showToast({ title: '入驻申请已提交', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 1500);
    } catch (e) {
      wx.showToast({ title: e.message || '提交失败', icon: 'none' });
    } finally {
      this.setData({ submitting: false });
    }
  },

  getBaseUrl() {
    return 'https://your-domain.com';
  }
});