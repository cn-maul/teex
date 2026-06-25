/**
 * 公共刷题工具函数
 */

// 题型标签映射
export const TYPE_LABELS = {
  single: '单选题',
  multi: '多选题',
  judge: '判断题',
  fill: '填空题',
}

/**
 * 获取题型中文标签
 * @param {string} type - 题型标识 (single/multi/judge/fill)
 * @returns {string}
 */
export function getTypeLabel(type) {
  return TYPE_LABELS[type] || '单选题'
}

/**
 * 解析选项 JSON 字符串
 * @param {string} optionsStr - JSON 数组字符串
 * @returns {string[]}
 */
export function parseOptions(optionsStr) {
  if (!optionsStr) return []
  try {
    const parsed = JSON.parse(optionsStr)
    if (!Array.isArray(parsed)) {
      console.warn('[parseOptions] options is not an array:', optionsStr)
      return []
    }
    return parsed
  } catch (e) {
    console.warn('[parseOptions] Failed to parse options:', optionsStr, e.message)
    return []
  }
}

/**
 * 获取选项字母（如 "A"）
 * 支持多种格式: "A. xxx", "A、xxx", "A) xxx", "（A）xxx", "A xxx"
 * @param {string} option - 选项字符串
 * @returns {string}
 */
export function getOptionLetter(option) {
  if (!option) return ''
  // Match a leading ASCII letter followed by a common delimiter or whitespace
  const match = option.trim().match(/^([A-Za-z])\s*[.、)）:\s]/)
  if (match) return match[1].toUpperCase()
  // Fallback: option might be just a bare letter
  const trimmed = option.trim()
  if (/^[A-Za-z]$/.test(trimmed)) return trimmed.toUpperCase()
  return trimmed.charAt(0)
}

/**
 * 获取选项文本内容（去掉字母前缀）
 * @param {string} option - 选项字符串
 * @returns {string}
 */
export function getOptionText(option) {
  if (!option || option.length < 2) return option || ''
  return option.substring(1).replace(/^[\s.、\s]+/, '').trim()
}
