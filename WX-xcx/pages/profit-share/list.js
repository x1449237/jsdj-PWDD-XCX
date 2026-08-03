/**
 * 架构规则：前端零业务逻辑。
 * 本页面仅负责：调用后端 API、setData 渲染后端返回字段、纯 UI 反馈。
 * 角色/状态文案、金额、时间均由后端返回：
 *   role_text / role_class / status_text / status_class / amount_text / time_text / tax_text / ratio_text
 * 月份分组、月度小计、排序均由后端预聚合后返回 groups（含 month_total_text）。
 */
const request = require('../../utils/request');

Page({
  data: {
    totalIncome: '',
    totalSettled: '',
    totalPending: '',
    totalTax: '',
    profitList: [],
    monthFilter: [],
    monthIndex: -1,
    selectedMonth: '',
    // 角色筛选选项由后端下发,前端不硬编码角色 label/value
    roleFilter: [],
    roleIndex: 0,
    selectedRole: '',
    page: 1,
    pageSize: 20,
    hasMore: true,
    loading: true,
    loadingMore: false,
    showTaxModal: false
  },

  onLoad() {
    this.loadSummary();
    this.loadProfitList();
  },

  onShow() {
    this.loadSummary();
  },

  async loadSummary() {
    try {
      const res = await request.get('/profit_share/summary');
      this.setData({
        totalIncome: res.total_income_text || '',
        totalSettled: res.total_settled_text || '',
        totalPending: res.total_pending_text || '',
        totalTax: res.total_tax_text || '',
        // 后端下发角色筛选选项 [{label, value}]
        roleFilter: res.role_filter_options || []
      });
    } catch (err) {
      // 忽略错误
    }
  },

  async loadProfitList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/profit_share/list', {
        page: this.data.page,
        pageSize: this.data.pageSize,
        month: this.data.selectedMonth,
        role: this.data.selectedRole
      });

      const groups = this.mapGroups(res.groups || []);

      this.setData({
        profitList: groups,
        // 月份筛选选项由后端下发 month_filter_options(含"全部"选项),前端不硬编码
        monthFilter: res.month_filter_options || [],
        hasMore: res.hasMore !== false,
        loading: false
      });
    } catch (err) {
      this.setData({ loading: false });
    }
  },

  async loadMoreProfit() {
    if (!this.data.hasMore || this.data.loadingMore) return;
    this.setData({ loadingMore: true });
    try {
      const nextPage = this.data.page + 1;
      const res = await request.get('/profit_share/list', {
        page: nextPage,
        pageSize: this.data.pageSize,
        month: this.data.selectedMonth,
        role: this.data.selectedRole
      });

      const newGroups = this.mapGroups(res.groups || []);
      const merged = this.mergeGroupedList(this.data.profitList, newGroups);

      this.setData({
        profitList: merged,
        page: nextPage,
        hasMore: res.hasMore !== false,
        loadingMore: false
      });
    } catch (err) {
      this.setData({ loadingMore: false });
    }
  },

  onReachBottom() {
    this.loadMoreProfit();
  },

  // 仅做结构映射：后端返回的预聚合分组转为前端渲染字段，不做金额换算/日期解析/排序
  mapGroups(groups) {
    return groups.map(group => ({
      month: group.month || '',
      monthTotal: group.month_total_text || '',
      items: (group.items || []).map(item => this.mapProfitItem(item))
    }));
  },

  mapProfitItem(item) {
    return {
      ...item,
      amountText: item.amount_text || '',
      ratioText: item.ratio_text || '',
      roleText: item.role_text || '',
      roleClass: item.role_class || '',
      statusText: item.status_text || '',
      statusClass: item.status_class || '',
      timeText: item.time_text || '',
      taxText: item.tax_text || ''
    };
  },

  // 分页合并：仅按月份名拼接 items，不重算小计、不排序（均由后端负责）
  mergeGroupedList(existing, newGroups) {
    const merged = [...existing];
    newGroups.forEach(newGroup => {
      const existingGroup = merged.find(g => g.month === newGroup.month);
      if (existingGroup) {
        const existingIds = new Set(existingGroup.items.map(i => i.id));
        const uniqueNew = newGroup.items.filter(i => !existingIds.has(i.id));
        existingGroup.items = [...existingGroup.items, ...uniqueNew];
      } else {
        merged.push(newGroup);
      }
    });
    return merged;
  },

  onMonthChange(e) {
    const index = parseInt(e.detail.value);
    const monthItem = this.data.monthFilter[index] || {};
    this.setData({
      monthIndex: index,
      selectedMonth: monthItem.value || '',
      page: 1,
      profitList: []
    });
    this.loadProfitList();
  },

  onRoleChange(e) {
    const index = parseInt(e.detail.value);
    const role = this.data.roleFilter[index].value;
    this.setData({
      roleIndex: index,
      selectedRole: role,
      page: 1,
      profitList: []
    });
    this.loadProfitList();
  },

  showTaxRule() {
    this.setData({ showTaxModal: true });
  },

  closeTaxModal() {
    this.setData({ showTaxModal: false });
  },

  goToOrder(e) {
    const orderNo = e.currentTarget.dataset.orderno;
    if (!orderNo) return;
    wx.navigateTo({
      url: `/order-flow/detail/detail?orderNo=${orderNo}`
    });
  },

  noop() {}
});
