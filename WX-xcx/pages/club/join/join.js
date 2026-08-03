const request = require('../../../utils/request');

const APP = getApp();

// 架构铁律：所有业务逻辑(身份证校验/年龄计算/脱敏/地址完整性/入驻资格/PDF校验)
// 全部由 Go 后端执行。前端只负责:
//   1) 收集用户输入 → 调用后端 API
//   2) 渲染后端返回的字段(_text/_masked/_valid 等)
//   3) 纯 UI 反馈(非空提示、loading、toast)
// 前端禁止实现:身份证校验位算法、年龄计算、脱敏算法、状态映射、金额换算、
// 业务规则校验(如"年满16周岁""地址7字段完整")、PDF 真伪判定。

Page({
  data: {
    step: 1,
    totalSteps: 7,
    clubType: '',          // 'green_v' 个人 / 'blue_v' 企业
    isEnterprise: false,
    clubJoinOpen: true,
    personalDeposit: 0,
    enterpriseDeposit: 0,
    personalDepositText: '',   // 后端下发的金额渲染文本
    enterpriseDepositText: '',

    // 步骤1：须知
    agreed: false,

    // 步骤2：基础资料
    clubName: '',
    abbreviation: '',
    abbrOccupied: false,
    abbrLoading: false,
    abbrAlternatives: [], // 后端冲突时推荐的3套备选缩写
    realName: '',
    idCard: '',
    idCardValid: false,       // 后端校验结果(异步)
    idCardAgeText: '',        // 后端下发的年龄文案(如"24岁 / 已成年")
    phone: '',
    // 地址7字段(全部必填,完整性由后端 validate-step 判定)
    addressProvince: '',
    addressCity: '',
    addressDistrict: '',
    addressStreet: '',
    addressCommunity: '',
    addressBuilding: '',
    addressHouseNo: '',

    // 步骤3：活体认证
    idCardFront: '',
    idCardBack: '',
    livenessStatus: 0,
    feeAgreed: false, // 活体认证收费知情同意

    // 步骤4：合同
    contractFile: '',
    contractPdfValid: false, // 后端 PDF 校验结果

    // 企业专属
    enterpriseName: '',
    creditCode: '',
    businessLicense: '',
    corporateBank: '',
    corporateAccount: '',
    corporateAccountMasked: '', // 后端返回的脱敏展示(仅展示用)
    handleType: 'self',    // self / agent
    agentName: '',
    agentIdCard: '',
    agentIdCardFront: '',
    agentIdCardBack: '',
    agentAuthorization: '', // 代办合同PDF
    agentAuthPdfValid: false,

    // 提交
    submitting: false,
    stepValidating: false,  // 步骤校验进行中(防重复点击)

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
      const res = await request.get('/clubs/join-switch');
      const isOpen = res?.club_join_open === true;
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
        personalDeposit: res?.personal_deposit || 0,
        enterpriseDeposit: res?.enterprise_deposit || 0,
        personalDepositText: res?.personal_deposit_text || '',
        enterpriseDepositText: res?.enterprise_deposit_text || ''
      });
    } catch (e) {
      wx.showToast({ title: '网络异常', icon: 'none' });
    }
  },

  // ============ 草稿保存/恢复 ============
  async restoreDraft() {
    try {
      const draft = await request.get('/clubs/draft', { club_type: this.data.clubType });
      if (!draft) return;
      if (draft.expired) {
        // 草稿过期提示
        wx.showToast({ title: '草稿已过期，请重新填写', icon: 'none', duration: 2500 });
        this.setData({ draftExpired: true });
        return;
      }
      // 恢复草稿(仅恢复非空字段,后端返回的渲染字段一并恢复)
      const fields = [
        'step', 'agreed', 'clubName', 'abbreviation', 'realName', 'idCard', 'phone',
        'addressProvince', 'addressCity', 'addressDistrict', 'addressStreet',
        'addressCommunity', 'addressBuilding', 'addressHouseNo',
        'idCardFront', 'idCardBack', 'livenessStatus',
        'contractFile', 'contractPdfValid',
        'enterpriseName', 'creditCode', 'businessLicense',
        'corporateBank', 'corporateAccount', 'corporateAccountMasked',
        'handleType', 'agentName', 'agentIdCard', 'agentIdCardFront', 'agentIdCardBack',
        'agentAuthorization', 'agentAuthPdfValid',
        'idCardValid', 'idCardAgeText'
      ];
      const patch = {};
      fields.forEach(f => {
        if (draft[f] !== undefined && draft[f] !== null && draft[f] !== '') {
          patch[f] = draft[f];
        }
      });
      if (Object.keys(patch).length > 0) {
        this.setData(patch);
      }
    } catch (e) {
      // 静默失败
    }
  },

  async saveDraft() {
    try {
      const d = this.data;
      await request.post('/clubs/draft', {
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

  // 步骤切换：所有步骤校验由后端 /clubs/validate-step 统一判定
  async nextStep() {
    const { step, clubJoinOpen, stepValidating } = this.data;
    if (!clubJoinOpen) return;
    if (stepValidating) return;

    this.setData({ stepValidating: true });
    try {
      // 调后端做权威步骤校验(身份证/年龄/地址完整性/PDF/代办合同等业务规则全部在后端)
      const res = await request.post('/clubs/validate-step', {
        club_type: this.data.clubType,
        step,
        ...this.collectStepPayload()
      });

      const canNext = res?.can_next === true;
      if (!canNext) {
        // 后端返回具体错误字段与提示文案,前端只负责 toast
        const msg = res?.message || '请检查当前步骤填写内容';
        wx.showToast({ title: msg, icon: 'none', duration: 2500 });
        return;
      }

      // 每步保存草稿
      await this.saveDraft();
      this.setData({ step: step + 1 });
    } catch (e) {
      wx.showToast({ title: e.message || '校验失败,请重试', icon: 'none' });
    } finally {
      this.setData({ stepValidating: false });
    }
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

  // 步骤2：俱乐部名称输入 → 调后端生成缩写
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
      const data = await request.post('/clubs/abbr', { name });
      // 后端返回的 abbreviation 已是大写;前端只渲染
      const abbr = data.abbreviation || '';
      const occupied = data.conflict === true || data.occupied === true;
      const alternatives = (data.alternatives || []).filter(Boolean).slice(0, 3);
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
    // 选定备选缩写后视为不再占用(占用状态以服务端为准,这里只是 UX 反馈)
    this.setData({ abbreviation: abbr, abbrOccupied: false, abbrAlternatives: [] });
  },

  onAbbrHelp() {
    wx.navigateTo({ url: '/pages/club/abbr-help/abbr-help' });
  },

  // 通用输入:仅 setData,不做任何业务校验/计算
  onInput(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ [field]: e.detail.value });
    // 身份证输入后异步调后端校验(后端返回 valid/age_text 等)
    if (field === 'idCard') {
      this.validateIdCardAsync(e.detail.value);
    }
  },

  // 办理方式单选切换（self / agent）
  onSelectHandleType(e) {
    const value = e.currentTarget.dataset.value;
    if (!value) return;
    this.setData({ handleType: value });
  },

  // 身份证异步校验 → 调后端 /thin/id-card/validate
  // 后端返回: valid / age / is_under_16 / age_text / message
  validateIdCardAsync(idCard) {
    if (!idCard || idCard.length !== 18) {
      this.setData({ idCardValid: false, idCardAgeText: '' });
      return;
    }
    request.post('/thin/id-card/validate', { id_card: idCard }).then((data) => {
      this.setData({
        idCardValid: data.valid === true,
        idCardAgeText: data.age_text || ''
      });
    }).catch(() => {
      this.setData({ idCardValid: false, idCardAgeText: '' });
    });
  },

  // 对公账户输入 → 仅 setData,脱敏由后端 submit 接口返回 masked 字段
  // (前端不再实现 maskAccount 算法)

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

  // 切换活体认证收费同意状态
  toggleFeeAgree() {
    this.setData({ feeAgreed: !this.data.feeAgreed });
  },

  // 活体认证(实际认证逻辑由后端 / 微信 SDK 完成,前端只触发并渲染状态)
  onLivenessCheck() {
    if (!this.data.feeAgreed) {
      wx.showToast({ title: '请先阅读并同意收费说明', icon: 'none' });
      return;
    }
    wx.showToast({ title: '活体认证中...', icon: 'loading' });
    request.post('/clubs/liveness-check', { club_type: this.data.clubType }).then((data) => {
      const status = data?.status || 0;
      this.setData({ livenessStatus: status });
      if (status === 1) {
        wx.showToast({ title: '认证通过', icon: 'success' });
      } else {
        wx.showToast({ title: data?.message || '认证失败,请重试', icon: 'none' });
      }
    }).catch(() => {
      wx.showToast({ title: '认证失败,请重试', icon: 'none' });
    });
  },

  // B4：上传合同 - 上传时调用 /clubs/upload-pdf 由后端做真伪/签名校验
  onUploadContract() {
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf'],
      success: (res) => {
        const file = res.tempFiles[0];
        // 后端 PDF 校验(签名/水印/页数等业务规则全部在后端)
        wx.showLoading({ title: '校验PDF中...' });
        wx.uploadFile({
          url: request.baseURL + '/clubs/upload-pdf',
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
                // 失败时 toast 提示具体原因(后端下发)
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

  // B7：企业代办合同上传 - 同样调用 /clubs/upload-pdf 校验
  onUploadAgentAuth() {
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf'],
      success: (res) => {
        const file = res.tempFiles[0];
        wx.showLoading({ title: '校验代办合同...' });
        wx.uploadFile({
          url: request.baseURL + '/clubs/upload-pdf',
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

  // 下载合同模板(后端 /clubs/contract-template 下发)
  onDownloadTemplate() {
    wx.showToast({ title: '下载合同模板中...', icon: 'loading' });
    request.get('/clubs/contract-template', { club_type: this.data.clubType }).then((data) => {
      const url = data?.url;
      if (!url) {
        wx.showToast({ title: '模板未就绪,请稍后重试', icon: 'none' });
        return;
      }
      wx.downloadFile({
        url: request.baseURL + url,
        header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
        success: (res) => {
          wx.openDocument({ filePath: res.tempFilePath, fileType: 'pdf' });
        },
        fail: () => {
          wx.showToast({ title: '下载失败,请重试', icon: 'none' });
        }
      });
    }).catch(() => {
      wx.showToast({ title: '下载失败,请重试', icon: 'none' });
    });
  },

  // 收集当前步骤表单数据(供 validate-step / submit 使用)
  collectStepPayload() {
    const d = this.data;
    const payload = {
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
      contract_file: d.contractFile
    };
    if (d.isEnterprise) {
      Object.assign(payload, {
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
    }
    return payload;
  },

  // 步骤5：提交(所有业务校验由后端 /clubs/submit 内部完成)
  // 后端返回 corporate_account_masked 等渲染字段
  async onSubmit() {
    if (this.data.submitting) return;
    this.setData({ submitting: true });

    const postData = {
      club_type: this.data.clubType,
      ...this.collectStepPayload()
    };

    try {
      const data = await request.post('/clubs/submit', postData);
      // 后端返回 masked 字段用于展示(可选)
      const masked = data?.corporate_account_masked;
      if (masked) {
        this.setData({ corporateAccountMasked: masked });
      }
      wx.showToast({ title: data?.message || '入驻申请已提交', icon: 'success' });
      // 提交成功后清空草稿
      try { await request.post('/clubs/draft', { club_type: this.data.clubType, clear: 1 }); } catch (e) {}
      setTimeout(() => wx.navigateBack(), 1500);
    } catch (e) {
      wx.showToast({ title: e.message || '提交失败', icon: 'none' });
    } finally {
      this.setData({ submitting: false });
    }
  },

  getBaseUrl() {
    return request.baseURL;
  }
});
