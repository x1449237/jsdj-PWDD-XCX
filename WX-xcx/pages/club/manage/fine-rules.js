/**
 * 架构规则：前端零业务逻辑。
 * 本页面仅负责：调用后端 API、setData 渲染后端返回字段、纯 UI 反馈。
 * 状态文案/颜色由后端返回 status_text / status_color；
 * 金额以元为单位字符串直接提交，展示使用后端 amount_text，不做元分换算。
 */
const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    list: [],
    page: 1,
    limit: 20,
    total: 0,
    loading: false,
    noMore: false,
    status: '',
    // 新建/编辑弹窗
    formVisible: false,
    formLoading: false,
    form: {
      id: 0,
      rule_name: '',
      amount: '',
      scene: '',
      content: ''
    },
    // 详情弹窗
    detailVisible: false,
    detail: null
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadList(true);
  },

  onPullDownRefresh() {
    this.loadList(true);
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList(false);
    }
  },

  async loadList(refresh) {
    if (this.data.loading) return;
    this.setData({ loading: true });
    try {
      const page = refresh ? 1 : this.data.page;
      const res = await request.get('/club/fine-rules/list', {
        club_id: this.data.clubId,
        page,
        limit: this.data.limit,
        status: this.data.status
      });
      const list = refresh ? (res.list || []) : [...this.data.list, ...(res.list || [])];
      this.setData({
        list,
        total: res.total || 0,
        page: page + 1,
        noMore: list.length >= (res.total || 0),
        loading: false
      });
    } catch (e) {
      wx.showToast({ title: e.message || '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }
    if (refresh) {
      wx.stopPullDownRefresh();
    }
  },

  onStatusChange(e) {
    const status = e.currentTarget.dataset.status;
    this.setData({ status, page: 1, list: [], noMore: false });
    this.loadList(true);
  },

  // 打开新建弹窗
  onOpenCreate() {
    this.setData({
      formVisible: true,
      form: {
        id: 0,
        rule_name: '',
        amount: '',
        scene: '',
        content: ''
      }
    });
  },

  // 关闭弹窗
  onCloseForm() {
    this.setData({ formVisible: false });
  },

  onFormInput(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ ['form.' + field]: e.detail.value });
  },

  // 提交新建/编辑（金额以元字符串提交，由后端处理）
  async onSubmitForm() {
    const f = this.data.form;
    if (!f.rule_name.trim()) {
      wx.showToast({ title: '请填写规则名称', icon: 'none' });
      return;
    }
    if (!f.amount || Number(f.amount) <= 0) {
      wx.showToast({ title: '请填写有效罚款金额', icon: 'none' });
      return;
    }
    if (!f.scene.trim()) {
      wx.showToast({ title: '请填写适用场景', icon: 'none' });
      return;
    }
    this.setData({ formLoading: true });
    try {
      await request.post('/club/fine-rules/submit', {
        id: f.id || 0,
        club_id: this.data.clubId,
        rule_name: f.rule_name.trim(),
        amount: f.amount,
        scene: f.scene.trim(),
        content: f.content.trim()
      });
      wx.showToast({ title: f.id ? '已更新' : '已提交备案', icon: 'success' });
      this.setData({ formVisible: false, formLoading: false });
      this.loadList(true);
    } catch (e) {
      wx.showToast({ title: e.message || '提交失败', icon: 'none' });
      this.setData({ formLoading: false });
    }
  },

  // 查看详情（文案/颜色/金额均使用后端返回字段）
  async onDetail(e) {
    const id = e.currentTarget.dataset.id;
    try {
      const res = await request.get('/club/fine-rules/detail', { id });
      const detail = res.data || res || {};
      this.setData({ detail, detailVisible: true });
    } catch (e) {
      wx.showToast({ title: e.message || '加载详情失败', icon: 'none' });
    }
  },

  onCloseDetail() {
    this.setData({ detailVisible: false });
  },

  // 编辑（仅 pending/rejected 可编辑）
  onEditFromDetail() {
    const d = this.data.detail;
    if (!d) return;
    this.setData({
      detailVisible: false,
      formVisible: true,
      form: {
        id: d.id,
        rule_name: d.rule_name || '',
        amount: d.amount_text !== undefined ? String(d.amount_text) : '',
        scene: d.scene || '',
        content: d.content || ''
      }
    });
  },

  // 主动下架（仅 approved 可下架）
  async onRevoke(e) {
    const id = (e.currentTarget.dataset.id) || (this.data.detail && this.data.detail.id);
    if (!id) return;
    wx.showModal({
      title: '下架罚款规则',
      content: '确定下架该罚款规则吗？下架后将不再对成员生效。',
      success: async (res) => {
        if (!res.confirm) return;
        try {
          await request.post('/club/fine-rules/revoke', { id, club_id: this.data.clubId });
          wx.showToast({ title: '已下架', icon: 'success' });
          this.setData({ detailVisible: false });
          this.loadList(true);
        } catch (e) {
          wx.showToast({ title: e.message || '操作失败', icon: 'none' });
        }
      }
    });
  }
});
