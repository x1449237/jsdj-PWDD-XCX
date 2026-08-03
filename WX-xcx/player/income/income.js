/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    availableBalance: '0.00',
    frozenBalance: '0.00',
    incomeList: [],
    monthFilter: [],
    monthIndex: -1,
    selectedMonth: '',
    page: 1,
    pageSize: 20,
    hasMore: true,
    loading: true,
    loadingMore: false,
    showRuleModal: false
  },

  onLoad() {
    this.loadBalance();
    this.loadIncomeList();
  },

  onShow() {
    this.loadBalance();
  },

  /* ========== 余额 ========== */
  async loadBalance() {
    try {
      const res = await request.get('/player/wallet/balance');
      this.setData({
        availableBalance: res.available_balance_text || res.availableBalance || '0.00',
        frozenBalance: res.frozen_balance_text || res.frozenBalance || '0.00'
      });
    } catch (err) {
      // 忽略错误
    }
  },

  /* ========== 收入明细（后端返回预分组 groups 与 months） ========== */
  async loadIncomeList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/player/income/list', {
        page: this.data.page,
        pageSize: this.data.pageSize,
        month: this.data.selectedMonth
      });

      const groups = (res.groups || []).map(group => this.formatIncomeGroup(group));

      this.setData({
        incomeList: groups,
        // 月份筛选选项由后端下发 month_filter_options(含"全部"选项),前端不硬编码
        monthFilter: res.month_filter_options || [],
        hasMore: res.hasMore !== false,
        loading: false
      });
    } catch (err) {
      this.setData({ loading: false });
    }
  },

  async loadMoreIncome() {
    if (!this.data.hasMore || this.data.loadingMore) return;
    this.setData({ loadingMore: true });
    try {
      const nextPage = this.data.page + 1;
      const res = await request.get('/player/income/list', {
        page: nextPage,
        pageSize: this.data.pageSize,
        month: this.data.selectedMonth
      });

      const newGroups = (res.groups || []).map(group => this.formatIncomeGroup(group));
      const merged = [...this.data.incomeList];
      newGroups.forEach(newGroup => {
        const existing = merged.find(g => g.month === newGroup.month);
        if (existing) {
          existing.items = [...(existing.items || []), ...(newGroup.items || [])];
        } else {
          merged.push(newGroup);
        }
      });

      this.setData({
        incomeList: merged,
        page: nextPage,
        hasMore: res.hasMore !== false,
        loadingMore: false
      });
    } catch (err) {
      this.setData({ loadingMore: false });
    }
  },

  onReachBottom() {
    this.loadMoreIncome();
  },

  formatIncomeGroup(group) {
    return {
      ...group,
      month: group.month || '',
      monthTotal: group.month_total_text || group.monthTotal || '',
      items: (group.items || []).map(item => this.formatIncomeItem(item))
    };
  },

  formatIncomeItem(item) {
    return {
      ...item,
      typeText: item.type_text || '',
      typeIconClass: item.type_icon_class || '',
      typeClass: item.type_class || '',
      amountText: item.amount_text || '',
      timeText: item.time_text || '',
      statusText: item.status_text || '',
      orderNo: item.order_no || item.orderNo || ''
    };
  },

  /* ========== 月份筛选 ========== */
  onMonthChange(e) {
    const index = parseInt(e.detail.value);
    const monthItem = this.data.monthFilter[index] || {};
    this.setData({
      monthIndex: index,
      selectedMonth: monthItem.value || '',
      page: 1,
      incomeList: []
    });
    this.loadIncomeList();
  },

  /* ========== 提现 ========== */
  // 是否可提现由后端 can_withdraw 字段决定,前端不解析余额数值
  goWithdraw() {
    request.get('/player/wallet/withdraw-check').then((res) => {
      if (res && res.can_withdraw === true) {
        wx.navigateTo({ url: '/package-wallet/withdraw/withdraw' });
      } else {
        wx.showToast({ title: res.message || '暂不可提现', icon: 'none' });
      }
    }).catch(() => {
      wx.showToast({ title: '查询失败', icon: 'none' });
    });
  },

  /* ========== 提现规则 ========== */
  showWithdrawRule() {
    this.setData({ showRuleModal: true });
  },

  closeRuleModal() {
    this.setData({ showRuleModal: false });
  },

  noop() {}
});
