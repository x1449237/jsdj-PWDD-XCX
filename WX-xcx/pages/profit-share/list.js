const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    totalIncome: '0.00',
    totalSettled: '0.00',
    totalPending: '0.00',
    totalTax: '0.00',
    profitList: [],
    monthFilter: [],
    monthIndex: -1,
    selectedMonth: '',
    roleFilter: [
      { label: '全部', value: '' },
      { label: '打手', value: 1 },
      { label: '俱乐部', value: 2 },
      { label: '分销商', value: 3 }
    ],
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
        totalIncome: util.fenToYuan(res.totalIncome || 0),
        totalSettled: util.fenToYuan(res.totalSettled || 0),
        totalPending: util.fenToYuan(res.totalPending || 0),
        totalTax: util.fenToYuan(res.totalTax || 0)
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

      const groupedList = this.groupByMonth(res.list || []);
      const months = this.extractMonths(groupedList);

      this.setData({
        profitList: groupedList,
        monthFilter: ['全部月份', ...months],
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

      const newItems = res.list || [];
      const mergedList = this.mergeGroupedList(this.data.profitList, newItems);

      this.setData({
        profitList: mergedList,
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

  groupByMonth(rawList) {
    const grouped = {};
    rawList.forEach(item => {
      const date = new Date(item.createTime);
      const month = `${date.getFullYear()}年${date.getMonth() + 1}月`;
      if (!grouped[month]) {
        grouped[month] = { month, items: [], monthTotal: 0 };
      }
      const formatted = this.formatProfitItem(item);
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

  formatProfitItem(item) {
    const roleMap = {
      1: { text: '打手', class: 'role-player' },
      2: { text: '俱乐部', class: 'role-club' },
      3: { text: '分销商', class: 'role-distributor' },
      4: { text: '平台', class: 'role-platform' }
    };
    const roleInfo = roleMap[item.role] || { text: '未知', class: 'role-player' };

    const statusMap = {
      0: { text: '待结算', class: 'status-pending' },
      1: { text: '已结算', class: 'status-settled' },
      2: { text: '已冻结', class: 'status-frozen' },
      3: { text: '已退款', class: 'status-refund' }
    };
    const statusInfo = statusMap[item.status] || { text: '未知', class: 'status-pending' };

    return {
      ...item,
      amountFen: item.amount || 0,
      amountText: util.fenToYuan(Math.abs(item.amount || 0)),
      ratioText: item.ratio ? item.ratio + '%' : '',
      roleText: roleInfo.text,
      roleClass: roleInfo.class,
      statusText: statusInfo.text,
      statusClass: statusInfo.class,
      timeText: util.formatTime(item.createTime, 'MM-DD HH:mm'),
      taxText: item.taxAmount ? util.fenToYuan(item.taxAmount) : '0.00'
    };
  },

  onMonthChange(e) {
    const index = parseInt(e.detail.value);
    const month = index === 0 ? '' : this.data.monthFilter[index];
    this.setData({
      monthIndex: index,
      selectedMonth: month,
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
