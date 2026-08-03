/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
Component({
  properties: {
    order: {
      type: Object,
      value: {
        order_no: '',
        game_name: '',
        game_icon: '',
        service_type: '',
        rank: '',
        user_avatar: '',
        user_nickname: '',
        user_level: '',
        user_rate: 0,
        amount: 0,
        duration: 0,
        create_time: '',
        status: 'pending'
      },
      observer: function (newVal) {
        this.updateStatusInfo(newVal);
      }
    },
    showActions: {
      type: Boolean,
      value: true
    },
    extraClass: {
      type: String,
      value: ''
    }
  },

  data: {
    orderStatusText: '',
    statusTagClass: ''
  },

  lifetimes: {
    attached() {
      this.updateStatusInfo(this.properties.order);
    }
  },

  methods: {
    updateStatusInfo(order) {
      if (!order) return;
      this.setData({
        orderStatusText: order.status_text || '',
        statusTagClass: order.status_tag_class || ''
      });
    },

    onTap() {
      this.triggerEvent('tap', { order: this.properties.order });
    },

    onCancel() {
      this.triggerEvent('cancel', { order: this.properties.order });
    },

    onPay() {
      this.triggerEvent('pay', { order: this.properties.order });
    },

    onContact() {
      this.triggerEvent('contact', { order: this.properties.order });
    },

    onEvaluate() {
      this.triggerEvent('evaluate', { order: this.properties.order });
    },

    onViewAppeal() {
      this.triggerEvent('viewappeal', { order: this.properties.order });
    }
  }
});
