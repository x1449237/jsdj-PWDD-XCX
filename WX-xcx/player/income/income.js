const request = require('../../utils/request');
const util = require('../../utils/util');

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
        availableBalance: util.fenToYuan(res.availableBalance || 0),
        frozenBalance: util.fenToYuan(res.frozenBalance || 0)
      });
    } catch (err) {
      // 忽略错误
    }
  },

  /* ========== 收入明细 ========== */
  async loadIncomeList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/player/income/list', {
        page: this.data.page,
        pageSize: this.data.pageSize,
        month: this.data.selectedMonth
      });

      const groupedList = this.groupByMonth(res.list || []);
      const months = this.extractMonths(groupedList);

      this.setData({
        incomeList: groupedList,
        monthFilter: ['全部月份', ...months],
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

      const newItems = res.list || [];
      const mergedList = this.mergeGroupedList(this.data.incomeList, newItems);

      this.setData({
        incomeList: mergedList,
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

  groupByMonth(rawList) {
    const grouped = {};
    rawList.forEach(item => {
      const date = new Date(item.createTime);
      const month = `${date.getFullYear()}年${date.getMonth() + 1}月`;
      if (!grouped[month]) {
        grouped[month] = { month, items: [], monthTotal: 0 };
      }
      const formatted = this.formatIncomeItem(item);
      grouped[month].items.push(formatted);
      grouped[month].monthTotal += (item.amount || 0);
    });

    const result = Object.values(grouped);
    result.forEach(group => {
      group.monthTotal = util.fenToYuan(group.monthTotal);
      group.items.sort((a, b) => new Date(b.createTime) - new Date(a.createTime));
    });
    result.sort((a, b) => {
      const [ay, am] = a.month.replace(/[年月]/g, '-').split('-').filter(Boolean);
      const [by, bm] = b.month.replace(/[年月]/g, '-').split('-').filter(Boolean);
      return new Date(by, bm - 1) - new Date(ay, am - 1);
    });
    return result;
  },

  mergeGroupedList(existing, newItems) {
    const merged = [...existing];
    const newGrouped = this.groupByMonth(newItems);
    newGrouped.forEach(newGroup => {
      const existingGroup = merged.find(g => g.month === newGroup.month);
      if (existingGroup) {
        const existingIds = new Set(existingGroup.items.map(i => i.id));
        const uniqueNew = newGroup.items.filter(i => !existingIds.has(i.id));
        existingGroup.items = [...existingGroup.items, ...uniqueNew];
        const totalFen = existingGroup.items.reduce((sum, i) => sum + (i.amountFen || 0), 0);
        existingGroup.monthTotal = util.fenToYuan(totalFen);
      } else {
        merged.push(newGroup);
      }
    });
    merged.sort((a, b) => {
      const [ay, am] = a.month.replace(/[年月]/g, '-').split('-').filter(Boolean);
      const [by, bm] = b.month.replace(/[年月]/g, '-').split('-').filter(Boolean);
      return new Date(by, bm - 1) - new Date(ay, am - 1);
    });
    return merged;
  },

  extractMonths(groupedList) {
    return groupedList.map(g => g.month);
  },

  formatIncomeItem(item) {
    const typeMap = {
      1: { text: '服务收入', iconClass: 'service', class: 'type-service' },
      2: { text: '打赏', iconClass: 'tip', class: 'type-tip' },
      3: { text: '佣金', iconClass: 'commission', class: 'type-commission' }
    };
    const typeInfo = typeMap[item.type] || { text: '其他', iconClass: 'other', class: 'type-service' };

    return {
      ...item,
      amountFen: item.amount || 0,
      amountText: util.fenToYuan(Math.abs(item.amount || 0)),
      typeText: typeInfo.text,
      typeIconClass: typeInfo.iconClass,
      typeClass: typeInfo.class,
      timeText: util.formatTime(item.createTime, 'MM-DD HH:mm'),
      statusText: item.status === 1 ? '已到账' : (item.status === 0 ? '冻结中' : '')
    };
  },

  /* ========== 月份筛选 ========== */
  onMonthChange(e) {
    const index = parseInt(e.detail.value);
    const month = index === 0 ? '' : this.data.monthFilter[index];
    this.setData({
      monthIndex: index,
      selectedMonth: month,
      page: 1,
      incomeList: []
    });
    this.loadIncomeList();
  },

  /* ========== 提现 ========== */
  goWithdraw() {
    const available = parseFloat(this.data.availableBalance);
    if (available <= 0) {
      wx.showToast({ title: '暂无可提现余额', icon: 'none' });
      return;
    }
    wx.navigateTo({
      url: '/package-wallet/withdraw/withdraw'
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