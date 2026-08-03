/**
 * validator.js — 小程序前端零逻辑架构
 *
 * 架构铁律：所有业务校验逻辑由 Go 后端执行,前端禁止做任何校验。
 * 本文件仅保留非空提示(UX 反馈),所有格式/规则校验调用后端 /thin/* API。
 * 后端返回权威校验结果,前端只渲染。
 */
const request = require('./request');

// 非空提示(仅 UI 反馈,非业务校验)
const checkNotEmpty = (val, fieldName) => {
  if (!val || (typeof val === 'string' && !val.trim())) {
    return { valid: false, message: '请输入' + fieldName };
  }
  return { valid: true, message: '' };
};

// 手机号校验 → 调后端 API
const validatePhone = async (phone) => {
  if (!phone) return { valid: false, message: '请输入手机号' };
  try {
    const res = await request.post('/thin/phone/validate', { phone });
    return { valid: !!res.valid, message: res.valid ? '' : '请输入正确的手机号' };
  } catch (e) {
    return { valid: false, message: '校验失败,请重试' };
  }
};

// 身份证校验 → 调后端 API(含校验位算法+年龄计算)
const validateIdCard = async (idCard) => {
  if (!idCard) return { valid: false, message: '请输入身份证号' };
  try {
    const res = await request.post('/thin/id-card/validate', { id_card: idCard });
    return { valid: !!res.valid, message: res.valid ? '' : (res.message || '身份证号不正确') };
  } catch (e) {
    return { valid: false, message: '校验失败,请重试' };
  }
};

// 密码强度 → 调后端 API
const validatePassword = async (password) => {
  if (!password) return { valid: false, message: '请输入密码' };
  try {
    const res = await request.post('/thin/password/strength', { password });
    return { valid: !!res.valid, message: res.valid ? '' : res.message, strength: res.score, level: res.level };
  } catch (e) {
    return { valid: false, message: '校验失败,请重试' };
  }
};

// 金额校验 → 调后端 API
const validateAmount = async (amount) => {
  if (!amount && amount !== 0) return { valid: false, message: '请输入金额' };
  try {
    const res = await request.post('/thin/amount/validate', { amount: String(amount) });
    return { valid: !!res.valid, message: res.valid ? '' : (res.message || '金额不正确') };
  } catch (e) {
    return { valid: false, message: '校验失败,请重试' };
  }
};

// 短信验证码非空(实际有效性由后端在提交时校验)
const validateSmsCode = (code) => {
  return checkNotEmpty(code, '验证码');
};

// 邀请码非空(实际有效性由后端在提交时校验)
const validateInviteCode = (code) => {
  return checkNotEmpty(code, '邀请码');
};

// 真实姓名非空(实际格式由后端在提交时校验)
const validateRealName = (name) => {
  return checkNotEmpty(name, '真实姓名');
};

module.exports = {
  checkNotEmpty,
  validatePhone,
  validateIdCard,
  validatePassword,
  validateAmount,
  validateSmsCode,
  validateInviteCode,
  validateRealName
};
