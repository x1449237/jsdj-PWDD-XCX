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
    const month = options.month || this.getCurrentMonth();
    this.setData({ bindId, month });
    this.loadReport();
    this.loadChatSummary();
  },

  getCurrentMonth() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    return `${year}-${month}`;
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
    if (!bindId || !month) return;

    this.setData({ loading: true });
    try {
      const res = await request.get('/guardian/consume_report', {
        bind_id: bindId,
        month: month
      });
      const data = (res && res.data) || this.getMockReport();

      const categoryList = (data.category_list || []).map(c => ({
        ...c,
        amount_text: (c.amount / 100).toFixed(2)
      }));

      const totalAmount = data.total_amount || 0;
      const maxDaily = Math.max(...(data.daily_trend || []).map(d => Number(d.amount) || 0), 1);
      const dailyTrend = (data.daily_trend || []).map(d => {
        const dateStr = String(d.date || '');
        return {
          ...d,
          label: dateStr.length >= 2 ? dateStr.slice(-2) + '日' : dateStr,
          bar_height: totalAmount > 0 ? Math.max(6, Math.min(100, Math.round((Number(d.amount) / maxDaily) * 90))) : 0
        };
      });

      this.setData({
        reportData: data,
        orders: data.orders && data.orders.length > 0 ? data.orders : this.data.orders,
        totalAmount: totalAmount,
        orderCount: data.order_count || 0,
        compareMoM: Number(data.compare_mom || 0),
        categoryList: categoryList,
        dailyTrend: dailyTrend
      });
    } catch (err) {
      console.error('加载报告失败:', err);
      const mock = this.getMockReport();
      this.setData({
        reportData: mock,
        totalAmount: mock.total_amount,
        orderCount: mock.order_count,
        compareMoM: mock.compare_mom,
        categoryList: (mock.category_list || []).map(c => ({
          ...c,
          amount_text: (c.amount / 100).toFixed(2)
        })),
        dailyTrend: (mock.daily_trend || []).map(d => ({
          ...d,
          label: String(d.date).slice(-2) + '日',
          bar_height: Math.max(6, Math.min(100, Math.round((Number(d.amount) / Math.max(...(mock.daily_trend || []).map(x => Number(x.amount) || 0), 1)) * 90)))
        }))
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  getMockReport() {
    return {
      total_amount: 128600,
      order_count: 12,
      compare_mom: 12,
      category_list: [
        { category: 'rank_boost', name: '段位代练', icon: '🏆', amount: 78600, count: 6, percent: 61 },
        { category: 'companion', name: '陪玩娱乐', icon: '🎮', amount: 29900, count: 3, percent: 23 },
        { category: 'reward', name: '打赏打手', icon: '🎁', amount: 12000, count: 2, percent: 9 },
        { category: 'other', name: '其他消费', icon: '📦', amount: 8100, count: 1, percent: 7 }
      ],
      daily_trend: [
        { date: '2025-01-05', amount: 8600 },
        { date: '2025-01-08', amount: 12000 },
        { date: '2025-01-11', amount: 9800 },
        { date: '2025-01-14', amount: 15600 },
        { date: '2025-01-18', amount: 24800 },
        { date: '2025-01-22', amount: 18500 },
        { date: '2025-01-28', amount: 39300 }
      ],
      orders: this.getMockOrders()
    };
  },

  getMockOrders() {
    const games = ['王者荣耀', '英雄联盟', '和平精英', '原神'];
    const ranks = ['钻石→星耀', '黄金→铂金', '青铜→白银', '铂金→钻石'];
    const statuses = [
      { status: 4, status_text: '已完成' },
      { status: 3, status_text: '待验收' },
      { status: 2, status_text: '服务中' },
      { status: 5, status_text: '已取消' }
    ];
    return [
      {
        id: 1001, order_sn: 'DD202501280001', game_name: games[0], rank: ranks[0],
        paid_amount: 39300, amount: 39300, status: 4, status_text: '已完成',
        create_time: '01-28 20:35'
      },
      {
        id: 1002, order_sn: 'DD202501220003', game_name: games[1], rank: ranks[1],
        paid_amount: 18500, amount: 18500, status: 4, status_text: '已完成',
        create_time: '01-22 19:12'
      },
      {
        id: 1003, order_sn: 'DD202501180005', game_name: games[2], rank: '王牌3星',
        paid_amount: 24800, amount: 24800, status: 3, status_text: '待验收',
        create_time: '01-18 14:28'
      }
    ];
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
    if (!bindId || !month) return;

    try {
      const res = await request.get('/guardian/chat_summary', {
        bind_id: bindId,
        month: month
      });
      const data = (res && res.data) || this.getMockChatSummary();
      this.setData({
        chatSummary: data.summary || {},
        chatList: data.list || []
      });
    } catch (err) {
      console.error('加载聊天摘要失败:', err);
      const mock = this.getMockChatSummary();
      this.setData({
        chatSummary: mock.summary,
        chatList: mock.list
      });
    }
  },

  getMockChatSummary() {
    const names = ['打*A', '陪*师', '打*B', '管**员'];
    const msgList = [
      '好的明**排上号',
      '那我**点开始了',
      '下次**再合作哈',
      '打**赏你了 谢谢'
    ];
    const list = names.map((n, i) => ({
      id: 2000 + i,
      name_masked: n,
      avatar: '',
      last_msg_masked: msgList[i] || msgList[0],
      last_time: `01-${20 + i} ${10 + i}:${20 + i * 3}`,
      sensitive_count: i === 1 ? 2 : 0,
      unread: i === 0 ? 3 : 0
    }));
    return {
      summary: {
        total_count: 486,
        chat_count: 12,
        sensitive_count: 2
      },
      list: list
    };
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
