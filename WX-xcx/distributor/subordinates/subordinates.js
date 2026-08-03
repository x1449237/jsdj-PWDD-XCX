/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    activeTab: 1,
    searchKeyword: '',
    subList: [],
    level1Total: 0,
    level2Total: 0,
    page: 1,
    pageSize: 20,
    hasMore: true,
    loading: true,
    loadingMore: false
  },

  onLoad() {
    this.loadSubList();
    this.loadCounts();
  },

  async loadCounts() {
    try {
      const res = await request.get('/distributor/subordinates/count');
      this.setData({
        level1Total: res.level1Count || 0,
        level2Total: res.level2Count || 0
      });
    } catch (err) {
      // 忽略
    }
  },

  async loadSubList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/distributor/subordinates', {
        level: this.data.activeTab,
        page: this.data.page,
        pageSize: this.data.pageSize,
        keyword: this.data.searchKeyword
      });

      const list = (res.list || []).map(item => this.formatSubItem(item));
      this.setData({
        subList: list,
        hasMore: res.hasMore !== false,
        loading: false
      });
    } catch (err) {
      this.setData({ loading: false });
    }
  },

  async onLoadMore() {
    if (!this.data.hasMore || this.data.loadingMore) return;
    this.setData({ loadingMore: true });
    try {
      const nextPage = this.data.page + 1;
      const res = await request.get('/distributor/subordinates', {
        level: this.data.activeTab,
        page: nextPage,
        pageSize: this.data.pageSize,
        keyword: this.data.searchKeyword
      });

      const list = (res.list || []).map(item => this.formatSubItem(item));
      this.setData({
        subList: [...this.data.subList, ...list],
        page: nextPage,
        hasMore: res.hasMore !== false,
        loadingMore: false
      });
    } catch (err) {
      this.setData({ loadingMore: false });
    }
  },

  formatSubItem(item) {
    return {
      ...item,
      avatar: item.avatar || '/assets/images/default-avatar.png',
      nickname: item.nickname || '未知用户',
      level: item.level || this.data.activeTab,
      registerTimeText: item.register_time_text || item.registerTimeText || '',
      orderCount: item.orderCount || 0,
      isRealName: item.isRealName || false,
      parentName: item.parent_name_masked || item.parentName || ''
    };
  },

  /* ========== Tab 切换 ========== */
  switchTab(e) {
    const tab = parseInt(e.currentTarget.dataset.tab);
    if (tab === this.data.activeTab) return;
    this.setData({
      activeTab: tab,
      page: 1,
      subList: [],
      hasMore: true
    });
    this.loadSubList();
  },

  /* ========== 搜索 ========== */
  onSearchInput(e) {
    this.setData({ searchKeyword: e.detail.value });
  },

  onSearch() {
    this.setData({ page: 1, subList: [], hasMore: true });
    this.loadSubList();
  },

  clearSearch() {
    this.setData({ searchKeyword: '', page: 1, subList: [], hasMore: true });
    this.loadSubList();
  }
});
