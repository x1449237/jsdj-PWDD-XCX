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
    detail: null,
    statusMap: {
      pending: '待审核',
      approved: '已通过',
      rejected: '已驳回',
      revoked: '已下架'
    },
    statusColorMap: {
      pending: '#e6a23c',
      approved: '#67c23a',
      rejected: '#f56c6c',
      revoked: '#909399'
    }
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

  // 提交新建/编辑
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
        amount: Math.round(Number(f.amount) * 100), // 元转分
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

  // 查看详情
  async onDetail(e) {
    const id = e.currentTarget.dataset.id;
    try {
      const res = await request.get('/club/fine-rules/detail', { id });
      const detail = res.data || {};
      detail.status_text = this.data.statusMap[detail.status] || '未知';
      detail.status_color = this.data.statusColorMap[detail.status] || '#909399';
      detail.amount_yuan = detail.amount_yuan !== undefined
        ? detail.amount_yuan
        : (detail.amount !== undefined ? (detail.amount / 100).toFixed(2) : '0.00');
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
        amount: d.amount_yuan !== undefined ? String(d.amount_yuan) : '',
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
