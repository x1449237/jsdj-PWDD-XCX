const extApi = require('../../utils/ext-api');
const app = getApp();

const CHANGE_TYPE_MAP = {
  income: { text: '收入', tagClass: 'tag-success' },
  expense: { text: '支出', tagClass: 'tag-primary' }
};

const formatMoney = (v) => ((Number(v) || 0) / 100).toFixed(2);
const formatTime = (v) => v ? String(v).replace('T', ' ').slice(0, 16) : '-';

Page({
  data: {
    tabs: ['全部', '收入', '支出'],
    tabIndex: 0,
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
      return;
    }
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, list: [], noMore: false });
    this.loadList();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList();
    }
  },

  onTabTap(e) {
    const index = Number(e.currentTarget.dataset.index);
    if (index === this.data.tabIndex) return;
    this.setData({ tabIndex: index, page: 1, list: [], noMore: false });
    this.loadList();
  },

  loadList() {
    if (this.data.loading || this.data.noMore) return;
    this.setData({ loading: true });
    const params = { page: this.data.page, size: this.data.pageSize };
    const ct = ['', 'income', 'expense'][this.data.tabIndex];
    if (ct) params.change_type = ct;

    extApi.listWalletLogs(params).then((res) => {
      const list = (res.list || []).map((item) => {
        const isIncome = item.change_type === 'income';
        const conf = CHANGE_TYPE_MAP[item.change_type] || { text: item.change_type || '其他', tagClass: 'tag-warning' };
        const amt = Math.abs(Number(item.amount) || 0) / 100;
        return {
          ...item,
          typeText: conf.text,
          tagClass: conf.tagClass,
          amountText: (isIncome ? '+' : '-') + amt.toFixed(2),
          amountColor: isIncome ? '#07c160' : '#e94560',
          balanceText: formatMoney(item.balance_after),
          createdText: formatTime(item.created_at)
        };
      });
      this.setData({
        list: this.data.list.concat(list),
        page: this.data.page + 1,
        noMore: list.length < this.data.pageSize,
        loading: false
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  }
});
