/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段（amount_text/trend_height/month 等）
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数。
 * 约定：request.get(url, data) 直接 resolve 出内层 data.data，res 即数据对象本身。
 */
const request = require('../../../utils/request');

Page({
  data: {
    bindId: 0,
    month: '',
    activeTab: 0,
    reportData: null,
    orders: [],
    totalAmount: 0,
    orderCount: 0,
    compareMoM: 0,
    categoryList: [],
    dailyTrend: [],
    chatSummary: {
      total_count: 0,
      chat_count: 0,
      sensitive_count: 0
    },
    chatList: [],
    loading: false,
    hasMoreOrders: true,
    orderPage: 1,
    orderLimit: 20
  },

  onLoad(options) {
    const bindId = options.bind_id ? parseInt(options.bind_id) : 0;
    const month = options.month || '';
    this.setData({ bindId, month });
    this.loadReport();
    this.loadChatSummary();
  },

  onMonthChange(e) {
    const month = e.detail.value;
    this.setData({
      month,
      orders: [],
      orderPage: 1,
      hasMoreOrders: true
    });
    this.loadReport();
    this.loadChatSummary();
  },

  onTabChange(e) {
    const index = parseInt(e.currentTarget.dataset.index || 0);
    this.setData({ activeTab: index });

    if (index === 1 && this.data.orders.length === 0) {
      this.loadOrderList(true);
    }
  },

  async loadReport() {
    const { bindId, month } = this.data;
    if (!bindId) return;

    this.setData({ loading: true });
    try {
      const res = await request.get('/guardian/consume_report', {
        bind_id: bindId,
        month: month
      });
      const data = res || {};

      this.setData({
        reportData: data,
        month: data.month || month,
        orders: data.orders && data.orders.length > 0 ? data.orders : this.data.orders,
        totalAmount: data.total_amount || 0,
        orderCount: data.order_count || 0,
        compareMoM: data.compare_mom || 0,
        categoryList: data.category_list || [],
        dailyTrend: data.daily_trend || []
      });
    } catch (err) {
      console.error('加载报告失败:', err);
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({
        reportData: null,
        totalAmount: 0,
        orderCount: 0,
        compareMoM: 0,
        categoryList: [],
        dailyTrend: []
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  loadOrderList(reset) {
    if (reset) {
      this.setData({
        orders: [],
        orderPage: 1,
        hasMoreOrders: true
      });
    }
    if (this.data.loading || !this.data.hasMoreOrders) return;
    this.loadReport();
  },

  async loadChatSummary() {
    const { bindId, month } = this.data;
    if (!bindId) return;

    try {
      const res = await request.get('/guardian/chat_summary', {
        bind_id: bindId,
        month: month
      });
      const data = res || {};
      this.setData({
        chatSummary: data.summary || {},
        chatList: data.list || []
      });
    } catch (err) {
      console.error('加载聊天摘要失败:', err);
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({
        chatSummary: {},
        chatList: []
      });
    }
  },

  onOrderTap(e) {
    const orderId = e.currentTarget.dataset.id;
    if (!orderId) return;
    wx.navigateTo({
      url: `/order-flow/detail/detail?orderId=${orderId}&id=${orderId}`,
      fail: () => {
        wx.navigateTo({
          url: `/order-flow/detail/detail?id=${orderId}`
        });
      }
    });
  },

  onChatTap(e) {
    const chatId = e.currentTarget.dataset.id;
    if (!chatId) return;
    wx.showModal({
      title: '聊天详情',
      content: '聊天内容已脱敏处理。仅展示消息频次、时间分布，及系统识别的敏感关键词摘要。',
      showCancel: false,
      confirmText: '我知道了'
    });
  },

  onReachBottom() {
    if (this.data.activeTab === 1) {
      this.loadOrderList(false);
    }
  },

  onPullDownRefresh() {
    this.setData({
      orders: [],
      orderPage: 1,
      hasMoreOrders: true
    });
    this.loadReport().finally(() => {
      wx.stopPullDownRefresh();
    });
  }
});
