/**
 * 微信小程序订阅消息授权组件
 * 用于向用户请求订阅消息授权，支持单次和多次订阅
 */
Component({
  properties: {
    // 模板ID列表，逗号分隔
    tmplIds: {
      type: String,
      value: ''
    },
    // 是否自动触发
    auto: {
      type: Boolean,
      value: false
    },
    // 触发场景
    scene: {
      type: String,
      value: ''
    }
  },

  data: {
    subscribed: false,
    showTip: false
  },

  lifetimes: {
    attached() {
      if (this.properties.auto) {
        this.requestSubscribe();
      }
    }
  },

  methods: {
    /**
     * 请求订阅消息授权
     */
    requestSubscribe() {
      const tmplIds = this.properties.tmplIds;
      if (!tmplIds) {
        console.warn('[subscribe-msg] 未配置模板ID');
        return;
      }

      const ids = tmplIds.split(',').map(id => id.trim()).filter(id => id);
      if (ids.length === 0) {
        return;
      }

      const that = this;

      wx.requestSubscribeMessage({
        tmplIds: ids,
        success(res) {
          const accepted = [];
          const rejected = [];
          ids.forEach(id => {
            if (res[id] === 'accept') {
              accepted.push(id);
            } else {
              rejected.push(id);
            }
          });

          console.log('[subscribe-msg] 授权结果:', { accepted, rejected });

          that.setData({ subscribed: accepted.length > 0 });

          // 上报授权结果到后端
          that.reportSubscribeResult(accepted, rejected);

          // 触发事件
          that.triggerEvent('subscribed', {
            accepted,
            rejected,
            result: res
          });
        },
        fail(err) {
          console.error('[subscribe-msg] 授权失败:', err);

          // 触发失败事件
          that.triggerEvent('fail', { error: err });
        }
      });
    },

    /**
     * 上报订阅结果到后端
     */
    reportSubscribeResult(accepted, rejected) {
      const app = getApp();
      const request = require('../../utils/request');

      request.post('/subscribe/report', {
        accepted: accepted,
        rejected: rejected,
        scene: this.properties.scene
      }).catch(() => {
        // 静默失败
      });
    },

    /**
     * 显示授权提示
     */
    showSubscribeTip() {
      this.setData({ showTip: true });
    },

    /**
     * 隐藏授权提示
     */
    hideSubscribeTip() {
      this.setData({ showTip: false });
    }
  }
});