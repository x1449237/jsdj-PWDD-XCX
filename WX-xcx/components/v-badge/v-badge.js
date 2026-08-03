// 后端下发配置映射: 禁止前端自行定义V标样式(需求212)
// 本地只做兜底默认值,最终样式按 globalData.badgeRenderConfigs 覆盖
const FALLBACK_CONFIG = {
  golden_v: { badge_key: 'golden_v', badge_name: '平台官方', tooltip_text: '平台官方认证', size_ratio_vs_font: 1.0, display_priority: 100, attach_to_club_entity_only: 0 },
  blue_v:   { badge_key: 'blue_v',   badge_name: '企业认证俱乐部', tooltip_text: '企业认证俱乐部', size_ratio_vs_font: 0.95, display_priority: 80, attach_to_club_entity_only: 1 },
  green_v:  { badge_key: 'green_v',  badge_name: '个人认证俱乐部', tooltip_text: '个人认证俱乐部', size_ratio_vs_font: 0.95, display_priority: 70, attach_to_club_entity_only: 1 }
};

Component({
  properties: {
    badgeType: {
      type: String,
      value: 'golden_v'
      // golden_v | blue_v | green_v
      // up_bronze | up_advanced | up_high | up_elite | up_master | up_supreme
    },
    badgeText: {
      type: String,
      value: ''
    },
    // 俱乐部V标是否置灰(临时封禁 202)
    grayDisabled: {
      type: Boolean,
      value: false
    },
    // 是否黑名单(214 黑名单完全不渲染)
    isBlacklisted: {
      type: Boolean,
      value: false
    }
  },

  data: {
    badgeClass: '',
    badgeLabel: 'V',
    isUpMaster: false,
    badgeStyleStr: '',
    tooltipText: '',
    hideBadge: false
  },

  lifetimes: {
    attached() {
      this.updateBadgeStyle();
    }
  },

  observers: {
    'badgeType, grayDisabled, isBlacklisted': function() {
      this.updateBadgeStyle();
    }
  },

  methods: {
    // 读取后端下发的V标配置(没有则兜底默认)
    _getBadgeConfig(type) {
      const app = getApp();
      const cfgMap = (app && app.globalData && app.globalData.badgeRenderConfigs) || {};
      if (cfgMap && cfgMap[type]) return cfgMap[type];
      return FALLBACK_CONFIG[type] || FALLBACK_CONFIG.golden_v;
    },

    updateBadgeStyle() {
      const type = this.properties.badgeType;
      let badgeClass = '';
      let badgeLabel = 'V';
      let isUpMaster = false;
      let hideBadge = !!this.properties.isBlacklisted;

      // UP 主类型
      if (type && type.startsWith('up_')) {
        isUpMaster = true;
        badgeLabel = 'UP';
        switch (type) {
          case 'up_bronze':   badgeClass = 'badge-up-bronze'; break;
          case 'up_advanced': badgeClass = 'badge-up-advanced'; break;
          case 'up_high':     badgeClass = 'badge-up-high'; break;
          case 'up_elite':    badgeClass = 'badge-up-elite'; break;
          case 'up_master':   badgeClass = 'badge-up-master'; break;
          case 'up_supreme':  badgeClass = 'badge-up-supreme'; break;
          default:            badgeClass = 'badge-up-bronze';
        }
        this.setData({ badgeClass, badgeLabel, isUpMaster, hideBadge, badgeStyleStr: '', tooltipText: 'UP等级徽章' });
        return;
      }

      // V 标: 按后端下发渲染
      const cfg = this._getBadgeConfig(type);
      const ratio = parseFloat(cfg.size_ratio_vs_font || 1);
      const baseClass = (type === 'golden_v') ? 'badge-golden'
                       : (type === 'blue_v')   ? 'badge-blue'
                       : (type === 'green_v')  ? 'badge-green' : 'badge-golden';
      badgeClass = baseClass;
      const styles = [];
      if (ratio !== 1) {
        styles.push('transform: scale(' + ratio.toFixed(2) + ')');
      }
      // 临时封禁 V 标置灰(202)
      if (this.properties.grayDisabled) {
        styles.push('filter: grayscale(100%); opacity: 0.5');
      }
      this.setData({
        badgeClass,
        badgeLabel: 'V',
        isUpMaster: false,
        hideBadge,
        badgeStyleStr: styles.join(';'),
        tooltipText: cfg.tooltip_text || ''
      });
    },

    // 点击 V 标弹悬浮提示(204)
    onClickBadge() {
      if (this.data.hideBadge || !this.data.tooltipText) return;
      wx.showToast({ title: this.data.tooltipText, icon: 'none', duration: 2000 });
    }
  }
});