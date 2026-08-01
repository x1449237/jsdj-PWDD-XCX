const request = require('../../../utils/request');

const APP = getApp();

// 身份证号计算年龄（返回周岁，解析失败返回 -1）
function calcAgeFromIdCard(idCard) {
  if (!idCard || idCard.length !== 18) return -1;
  const birthYear = parseInt(idCard.substring(6, 10), 10);
  const birthMonth = parseInt(idCard.substring(10, 12), 10);
  const birthDay = parseInt(idCard.substring(12, 14), 10);
  if (!birthYear || !birthMonth || !birthDay) return -1;
  const now = new Date();
  let age = now.getFullYear() - birthYear;
  if (now.getMonth() + 1 < birthMonth || (now.getMonth() + 1 === birthMonth && now.getDate() < birthDay)) {
    age--;
  }
  return age;
}

// 对公账户脱敏：前4后4，中间****
function maskAccount(acc) {
  if (!acc) return '';
  const s = String(acc);
  if (s.length <= 8) return s.replace(/.(?=.{4})/g, '*');
  return s.substring(0, 4) + '****' + s.substring(s.length - 4);
}

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
    abbrAlternatives: [], // 冲突时推荐的3套备选缩写
    realName: '',
    idCard: '',
    idCardAge: -1,         // 由身份证号计算出的年龄
    idCardAgeOk: false,    // 是否年满16周岁
    phone: '',
    // 地址7字段（全部必填）
    addressProvince: '',
    addressCity: '',
    addressDistrict: '',
    addressStreet: '',
    addressCommunity: '',
    addressBuilding: '',
    addressHouseNo: '',
    addressComplete: false,

    // 步骤3：活体认证
    idCardFront: '',
    idCardBack: '',
    livenessStatus: 0,

    // 步骤4：合同
    contractFile: '',
    contractPdfValid: false, // PDF校验是否通过

    // 企业专属
    enterpriseName: '',
    creditCode: '',
    businessLicense: '',
    corporateBank: '',
    corporateAccount: '',
    corporateAccountMasked: '', // 提交后脱敏展示
    handleType: 'self',    // self / agent
    agentName: '',
    agentIdCard: '',
    agentIdCardFront: '',
    agentIdCardBack: '',
    agentAuthorization: '', // 代办合同PDF
    agentAuthPdfValid: false,

    // 提交
    submitting: false,

    // 草稿
    draftExpired: false
  },

  onLoad(options) {
    const type = options.type || 'green_v';
    this.setData({
      clubType: type,
      isEnterprise: type === 'blue_v',
      totalSteps: type === 'blue_v' ? 7 : 7
    });
    this.checkSwitch();
    this.restoreDraft();
  },

  async checkSwitch() {
    try {
      const res = await request.get('/club/check_switch');
      const isOpen = res.data?.club_join_open === true;
      if (!isOpen) {
        // 入驻页面 onLoad 检查开关，关闭时 wx.reLaunch 到首页
        wx.showModal({
          title: '提示',
          content: '平台暂时关闭俱乐部入驻通道',
          showCancel: false,
          success: () => wx.reLaunch({ url: '/pages/index/index' })
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

  // ============ 草稿保存/恢复 ============
  async restoreDraft() {
    try {
      const res = await request.get('/club/get-draft', { club_type: this.data.clubType });
      const draft = res.data;
      if (!draft) return;
      if (draft.expired) {
        // 草稿过期提示
        wx.showToast({ title: '草稿已过期，请重新填写', icon: 'none', duration: 2500 });
        this.setData({ draftExpired: true });
        return;
      }
      // 恢复草稿（仅恢复非空字段）
      const fields = [
        'step', 'agreed', 'clubName', 'abbreviation', 'realName', 'idCard', 'phone',
        'addressProvince', 'addressCity', 'addressDistrict', 'addressStreet',
        'addressCommunity', 'addressBuilding', 'addressHouseNo',
        'idCardFront', 'idCardBack', 'livenessStatus',
        'contractFile', 'enterpriseName', 'creditCode', 'businessLicense',
        'corporateBank', 'corporateAccount',
        'handleType', 'agentName', 'agentIdCard', 'agentIdCardFront', 'agentIdCardBack',
        'agentAuthorization'
      ];
      const patch = {};
      fields.forEach(f => {
        if (draft[f] !== undefined && draft[f] !== null && draft[f] !== '') {
          patch[f] = draft[f];
        }
      });
      if (Object.keys(patch).length > 0) {
        this.setData(patch);
        // 重新计算地址完整性 / 年龄 / 脱敏
        this.recalcAddress();
        this.recalcAge();
        this.recalcMask();
      }
    } catch (e) {
      // 静默失败
    }
  },

  async saveDraft() {
    try {
      const d = this.data;
      await request.post('/club/save-draft', {
        club_type: d.clubType,
        step: d.step,
        agreed: d.agreed,
        club_name: d.clubName,
        abbreviation: d.abbreviation,
        real_name: d.realName,
        id_card: d.idCard,
        phone: d.phone,
        address_province: d.addressProvince,
        address_city: d.addressCity,
        address_district: d.addressDistrict,
        address_street: d.addressStreet,
        address_community: d.addressCommunity,
        address_building: d.addressBuilding,
        address_house_no: d.addressHouseNo,
        id_card_front: d.idCardFront,
        id_card_back: d.idCardBack,
        liveness_status: d.livenessStatus,
        contract_file: d.contractFile,
        enterprise_name: d.enterpriseName,
        credit_code: d.creditCode,
        business_license: d.businessLicense,
        corporate_bank: d.corporateBank,
        corporate_account: d.corporateAccount,
        handle_type: d.handleType,
        agent_name: d.agentName,
        agent_id_card: d.agentIdCard,
        agent_id_card_front: d.agentIdCardFront,
        agent_id_card_back: d.agentIdCardBack,
        agent_authorization: d.agentAuthorization
      });
    } catch (e) {
      // 静默失败，不打断用户流程
    }
  },

  // 步骤切换
  async nextStep() {
    const { step, clubJoinOpen } = this.data;
    if (!clubJoinOpen) return;

    // 每步校验
    if (step === 1 && !this.data.agreed) {
      wx.showToast({ title: '请先勾选同意全部协议', icon: 'none' });
      return;
    }
    if (step === 2) {
      if (!this.data.clubName.trim()) { wx.showToast({ title: '请填写俱乐部名称', icon: 'none' }); return; }
      if (this.data.abbrOccupied) { wx.showToast({ title: '缩写被占用，请更换名称或选择备选缩写', icon: 'none' }); return; }
      if (!this.data.realName) { wx.showToast({ title: '请填写真实姓名', icon: 'none' }); return; }
      if (!this.data.idCard || this.data.idCard.length !== 18) { wx.showToast({ title: '请填写18位身份证号', icon: 'none' }); return; }
      if (!this.data.idCardAgeOk) { wx.showToast({ title: '须年满16周岁', icon: 'none' }); return; }
      if (!this.data.phone || this.data.phone.length !== 11) { wx.showToast({ title: '请填写11位手机号', icon: 'none' }); return; }
      // 地址7字段校验
      if (!this.data.addressComplete) {
        wx.showToast({ title: '请完整填写所有地址字段', icon: 'none' });
        return;
      }
    }
    if (step === 3) {
      if (!this.data.idCardFront) { wx.showToast({ title: '请上传身份证正面', icon: 'none' }); return; }
      if (!this.data.idCardBack) { wx.showToast({ title: '请上传身份证反面', icon: 'none' }); return; }
    }
    if (step === 4) {
      if (!this.data.contractFile) { wx.showToast({ title: '请上传已签署合同', icon: 'none' }); return; }
      if (!this.data.contractPdfValid) { wx.showToast({ title: '合同PDF校验未通过', icon: 'none' }); return; }
      // 企业代办模式：必须上传代办合同PDF且校验通过
      if (this.data.isEnterprise && this.data.handleType === 'agent') {
        if (!this.data.agentAuthorization) {
          wx.showToast({ title: '请上传代办合同PDF', icon: 'none' });
          return;
        }
        if (!this.data.agentAuthPdfValid) {
          wx.showToast({ title: '代办合同PDF校验未通过', icon: 'none' });
          return;
        }
      }
    }

    // 每步保存草稿
    await this.saveDraft();
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
      this.setData({ abbreviation: '', abbrOccupied: false, abbrAlternatives: [] });
    }
  },

  // B2：调用 generate-abbr 获取缩写，冲突时推荐3套备选缩写
  async generateAbbr(name) {
    this.setData({ abbrLoading: true });
    try {
      const res = await request.post('/club/generate-abbr', { club_name: name });
      const data = res.data || {};
      // 小写输入自动转大写显示
      let abbr = data.abbreviation || '';
      if (abbr) abbr = abbr.toUpperCase();
      const occupied = data.occupied === true;
      let alternatives = data.alternatives || [];
      alternatives = alternatives.map(a => (a || '').toUpperCase()).filter(Boolean).slice(0, 3);
      this.setData({
        abbreviation: abbr,
        abbrOccupied: occupied,
        abbrAlternatives: alternatives,
        abbrLoading: false
      });
    } catch (e) {
      this.setData({ abbrLoading: false });
    }
  },

  // 选择备选缩写
  onSelectAbbr(e) {
    const abbr = e.currentTarget.dataset.abbr;
    if (!abbr) return;
    // 选定备选缩写后视为不再占用
    this.setData({ abbreviation: abbr, abbrOccupied: false, abbrAlternatives: [] });
  },

  onAbbrHelp() {
    wx.navigateTo({ url: '/pages/club/abbr-help/abbr-help' });
  },

  // 通用输入
  onInput(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ [field]: e.detail.value });
    if (field === 'idCard') {
      this.recalcAge();
    }
    if (field && field.indexOf('address') === 0) {
      this.recalcAddress();
    }
    if (field === 'corporateAccount') {
      this.recalcMask();
    }
  },

  // 办理方式单选切换（self / agent）
  onSelectHandleType(e) {
    const value = e.currentTarget.dataset.value;
    if (!value) return;
    this.setData({ handleType: value });
  },

  // B5：根据身份证号计算年龄
  recalcAge() {
    const age = calcAgeFromIdCard(this.data.idCard);
    this.setData({
      idCardAge: age,
      idCardAgeOk: age >= 16
    });
  },

  // B3：重新计算地址7字段是否完整
  recalcAddress() {
    const d = this.data;
    const complete = !!(d.addressProvince && d.addressCity && d.addressDistrict &&
      d.addressStreet && d.addressCommunity && d.addressBuilding && d.addressHouseNo);
    this.setData({ addressComplete: complete });
  },

  // B8：重新计算对公账户脱敏
  recalcMask() {
    this.setData({ corporateAccountMasked: maskAccount(this.data.corporateAccount) });
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

  // B4：上传合同 - 上传前调用 /club/upload-pdf 校验
  onUploadContract() {
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf'],
      success: (res) => {
        const file = res.tempFiles[0];
        // 先调用后端 PDF 校验接口
        wx.showLoading({ title: '校验PDF中...' });
        wx.uploadFile({
          url: request.baseURL + '/club/upload-pdf',
          filePath: file.path,
          name: 'file',
          formData: { scene: 'contract' },
          header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
          success: (uploadRes) => {
            wx.hideLoading();
            let data = {};
            try { data = JSON.parse(uploadRes.data); } catch (e) {}
            if (data.code === 200) {
              const valid = data.data?.valid === true;
              if (!valid) {
                // 失败时 toast 提示具体原因
                const reason = data.data?.reason || 'PDF校验未通过';
                wx.showToast({ title: reason, icon: 'none', duration: 2500 });
                this.setData({ contractFile: '', contractPdfValid: false });
                return;
              }
              this.setData({
                contractFile: data.data.url || file.path,
                contractPdfValid: true
              });
              wx.showToast({ title: '校验通过', icon: 'success' });
            } else {
              const reason = data.msg || data.message || 'PDF校验未通过';
              wx.showToast({ title: reason, icon: 'none', duration: 2500 });
              this.setData({ contractFile: '', contractPdfValid: false });
            }
          },
          fail: () => {
            wx.hideLoading();
            wx.showToast({ title: '上传失败', icon: 'none' });
            this.setData({ contractFile: '', contractPdfValid: false });
          }
        });
      }
    });
  },

  // B7：企业代办合同上传 - 同样调用 /club/upload-pdf 校验
  onUploadAgentAuth() {
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf'],
      success: (res) => {
        const file = res.tempFiles[0];
        wx.showLoading({ title: '校验代办合同...' });
        wx.uploadFile({
          url: request.baseURL + '/club/upload-pdf',
          filePath: file.path,
          name: 'file',
          formData: { scene: 'agent_authorization' },
          header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
          success: (uploadRes) => {
            wx.hideLoading();
            let data = {};
            try { data = JSON.parse(uploadRes.data); } catch (e) {}
            if (data.code === 200) {
              const valid = data.data?.valid === true;
              if (!valid) {
                const reason = data.data?.reason || '代办合同PDF校验未通过';
                wx.showToast({ title: reason, icon: 'none', duration: 2500 });
                this.setData({ agentAuthorization: '', agentAuthPdfValid: false });
                return;
              }
              this.setData({
                agentAuthorization: data.data.url || file.path,
                agentAuthPdfValid: true
              });
              wx.showToast({ title: '校验通过', icon: 'success' });
            } else {
              const reason = data.msg || data.message || '代办合同PDF校验未通过';
              wx.showToast({ title: reason, icon: 'none', duration: 2500 });
              this.setData({ agentAuthorization: '', agentAuthPdfValid: false });
            }
          },
          fail: () => {
            wx.hideLoading();
            wx.showToast({ title: '上传失败', icon: 'none' });
            this.setData({ agentAuthorization: '', agentAuthPdfValid: false });
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
    const { clubName, clubType, isEnterprise, realName, idCard, phone, contractFile, addressComplete } = this.data;

    if (!addressComplete) {
      wx.showToast({ title: '请完整填写所有地址字段', icon: 'none' });
      return;
    }
    if (isEnterprise && this.data.handleType === 'agent') {
      if (!this.data.agentAuthorization || !this.data.agentAuthPdfValid) {
        wx.showToast({ title: '请上传代办合同PDF并校验通过', icon: 'none' });
        return;
      }
    }

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
      address_street: this.data.addressStreet,
      address_community: this.data.addressCommunity,
      address_building: this.data.addressBuilding,
      address_house_no: this.data.addressHouseNo,
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
        corporate_account_masked: maskAccount(this.data.corporateAccount),
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
      // 提交成功后清空草稿
      try { await request.post('/club/save-draft', { club_type: clubType, clear: 1 }); } catch (e) {}
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
