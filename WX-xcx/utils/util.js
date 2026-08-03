/**
 * util.js — 小程序前端零逻辑架构
 *
 * 架构铁律：金额换算、状态映射、脱敏等业务逻辑全部由 Go 后端完成。
 * 后端接口直接返回 _text/_color/_display 等渲染字段,前端只渲染。
 * 本文件仅保留纯 UI 工具函数(防抖/节流/延时),不包含任何业务逻辑。
 */

// 防抖(纯 UI 工具)
const debounce = (fn, delay = 300) => {
  let timer = null;
  return function (...args) {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      fn.apply(this, args);
      timer = null;
    }, delay);
  };
};

// 节流(纯 UI 工具)
const throttle = (fn, delay = 300) => {
  let lastTime = 0;
  return function (...args) {
    const now = Date.now();
    if (now - lastTime >= delay) {
      lastTime = now;
      fn.apply(this, args);
    }
  };
};

// 延时(纯 UI 工具)
const sleep = (ms) => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

// 数值限幅(纯 UI 工具,业务边界校验仍由后端执行)
const clamp = (value, min, max) => {
  return Math.min(Math.max(value, min), max);
};

// 临时客户端 ID(仅用于前端消息 temp_id,业务主键由后端生成)
const generateId = () => {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 10);
  return `${timestamp}${random}`;
};

// ====================================================================
// 以下函数已废弃 — 业务逻辑已迁移至 Go 后端,后端接口直接返回渲染字段
// 前端调用方应直接使用后端返回的 xxx_text / xxx_color / xxx_display 字段
// ====================================================================
//
// 【已废弃】formatTime      → 后端接口返回 time_text 字段
// 【已废弃】formatRelativeTime → 后端接口返回 relative_time_text 字段
// 【已废弃】formatMoney      → 后端接口返回 amount_text 字段(元字符串,含千分位)
// 【已废弃】fenToYuan        → 后端接口直接返回元字符串,前端不再换算
// 【已废弃】yuanToFen        → 前端提交元字符串,后端调 ParseYuanToFen 转换
// 【已废弃】maskPhone        → 后端接口返回已脱敏的 phone_masked 字段
// 【已废弃】maskIdCard       → 后端接口返回已脱敏的 id_card_masked 字段
// 【已废弃】maskName         → 后端接口返回已脱敏的 name_masked 字段
// 【已废弃】getOrderStatusText → 后端接口返回 status_text 字段
// 【已废弃】getOrderStatusColor → 后端接口返回 status_color 字段
//
// 如需临时兼容,以下为空实现(直接返回原始值,渲染以后端字段为准):
const formatTime = () => '';
const formatRelativeTime = () => '';
const formatMoney = (fen) => fen != null ? String(fen) : '';
const fenToYuan = (fen) => fen != null ? String(fen) : '';
const yuanToFen = (yuan) => yuan != null ? String(yuan) : '';
const maskPhone = (phone) => phone || '';
const maskIdCard = (idCard) => idCard || '';
const maskName = (name) => name || '';
const getOrderStatusText = () => '';
const getOrderStatusColor = () => '';

module.exports = {
  // 纯 UI 工具(保留)
  debounce,
  throttle,
  sleep,
  clamp,
  generateId,
  // 已废弃兼容桩(后端应直接返回渲染字段,前端不应调用)
  formatTime,
  formatRelativeTime,
  formatMoney,
  fenToYuan,
  yuanToFen,
  maskPhone,
  maskIdCard,
  maskName,
  getOrderStatusText,
  getOrderStatusColor,
};
