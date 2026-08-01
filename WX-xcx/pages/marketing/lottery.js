const request = require('../../../utils/request');
const app = getApp();

const DEFAULT_ACTIVITY = {
  id: 1,
  name: '订单完成抽奖',
  cost_type: 'free',
  cost_value: 0,
  daily_limit: 3,
  start_time: new Date().toISOString(),
  end_time: new Date(Date.now() + 86400000 * 30).toISOString()
};

const DEFAULT_PRIZES = [
  { id: 1, name: '50元优惠券', probability: 0.05, is_win: true },
  { id: 2, name: '20元优惠券', probability: 0.15, is_win: true },
  { id: 3, name: '谢谢参与', probability: 0.35, is_win: false },
  { id: 4, name: '5元优惠券', probability: 0.20, is_win: true },
  { id: 5, name: '谢谢参与', probability: 0.15, is_win: false },
  { id: 6, name: '免单券', probability: 0.02, is_win: true },
  { id: 7, name: '2元优惠券', probability: 0.05, is_win: true },
  { id: 8, name: '积分+100', probability: 0.03, is_win: true }
];

Page({
  data: {
    isLogin: false,
    activity: DEFAULT_ACTIVITY,
    prizes: DEFAULT_PRIZES,
    isDrawing: false,
    result: null,
    showResult: false,
    rotateDegree: 0,
    canDraw: 0,
    records: [],
    orderId: '',
    source: ''
  },

  onLoad(options) {
    const { orderId = '', source = '' } = options || {};
    this.setData({ orderId, source });
    this.checkLogin();
    this.loadActivity();
  },

  onShow() {
    this.checkLogin();
  },

  checkLogin() {
    const isLogin = !!(app.globalData && app.globalData.isLogin);
    this.setData({ isLogin });
  },

  loadActivity() {
    request.get('/lottery/activities').then((res) => {
      const activity = res.data?.activity || DEFAULT_ACTIVITY;
      const prizes = res.data?.prizes?.length ? res.data.prizes : DEFAULT_PRIZES;
      this.setData({ activity, prizes });
      if (this.data.isLogin) {
        this.loadRecords();
      }
    }).catch(() => {
      this.setData({
        activity: DEFAULT_ACTIVITY,
        prizes: DEFAULT_PRIZES
      });
    });
  },

  loadRecords() {
    request.get('/lottery/records', { limit: 10 }).then((res) => {
      this.setData({
        records: res.data?.list || []
      });
    }).catch(() => {
      this.setData({
        records: [
          { id: 1, prize_name: '20元优惠券', draw_time: '刚刚' },
          { id: 2, prize_name: '5元优惠券', draw_time: '10分钟前' }
        ]
      });
    });
  },

  drawMock() {
    const prizes = this.data.prizes;
    const rand = Math.random();
    let acc = 0;
    let hitIndex = 0;
    for (let i = 0; i < prizes.length; i++) {
      acc += (prizes[i].probability || (1 / prizes.length));
      if (rand <= acc) {
        hitIndex = i;
        break;
      }
    }
    return {
      is_win: prizes[hitIndex].is_win !== false,
      prize_id: prizes[hitIndex].id,
      prize_name: prizes[hitIndex].name,
      _hitIndex: hitIndex
    };
  },

  onDraw() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    if (this.data.isDrawing) return;
    if (!this.data.activity) {
      wx.showToast({ title: '活动未开始', icon: 'none' });
      return;
    }

    this.setData({ isDrawing: true });
    const { orderId, source } = this.data;

    request.post('/lottery/draw', {
      activity_id: this.data.activity.id,
      order_id: orderId,
      source
    }).then((res) => {
      const result = res.data || {};
      const hitIndex = this.data.prizes.findIndex(p => p.id === result.prize_id);
      const prizeIndex = hitIndex >= 0 ? hitIndex : (result._hitIndex != null ? result._hitIndex : 0);
      const baseDegree = 360 * 5;
      const eachDegree = 360 / this.data.prizes.length;
      const targetDegree = baseDegree + (prizeIndex * eachDegree) + (eachDegree / 2);

      this.setData({
        rotateDegree: this.data.rotateDegree + targetDegree,
        result: {
          ...result,
          prize_name: result.prize_name || this.data.prizes[prizeIndex]?.name || '谢谢参与',
          is_win: result.is_win !== false && result.prize_name !== '谢谢参与'
        }
      });

      setTimeout(() => {
        this.setData({
          isDrawing: false,
          showResult: true
        });
        this.loadRecords();
      }, 3000);
    }).catch(() => {
      const mockResult = this.drawMock();
      const prizeIndex = mockResult._hitIndex;
      const baseDegree = 360 * 5;
      const eachDegree = 360 / this.data.prizes.length;
      const targetDegree = baseDegree + (prizeIndex * eachDegree) + (eachDegree / 2);

      this.setData({
        rotateDegree: this.data.rotateDegree + targetDegree,
        result: {
          is_win: mockResult.is_win,
          prize_id: mockResult.id,
          prize_name: mockResult.prize_name
        }
      });

      setTimeout(() => {
        this.setData({
          isDrawing: false,
          showResult: true
        });
      }, 3000);
    });
  },

  closeResult() {
    this.setData({ showResult: false, result: null });
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
