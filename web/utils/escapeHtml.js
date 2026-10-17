/**
 * HTML 转义与关键词高亮工具。
 *
 * 背景：B站 标题/UP 主名等字段会经 v-html 渲染，未转义即存储型 XSS
 * （标题含 <img onerror=...> 会原样进 DOM）。所有走 v-html 的文本
 * 必须先过 escapeHtml，再包裹受控的 <span>。
 */

/**
 * 转义 HTML 特殊字符
 * @param {*} text 任意值（null/undefined 返回空串）
 * @returns {string} 转义后的安全文本
 */
export const escapeHtml = (text) => {
  if (text === null || text === undefined) return '';
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
};

/**
 * 转义正则元字符，防止用户输入的正则关键词抛 SyntaxError
 * @param {string} str 原始字符串
 * @returns {string} 可安全传入 new RegExp 的字符串
 */
export const escapeRegExp = (str) => str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

/**
 * 关键词高亮：先整体转义文本，再按转义后的关键词匹配并包裹 <span>，
 * 保证最终 HTML 里只有受控的标签。无关键词时返回纯转义文本。
 *
 * @param {string} text 原始文本
 * @param {string} keyword 关键词（空白分隔多个）
 * @param {string} [highlightClass='text-accent'] 高亮 span 的 class
 * @returns {string} 可安全用于 v-html 的 HTML 字符串
 */
export const highlightText = (text, keyword, highlightClass = 'text-accent') => {
  const escaped = escapeHtml(text);
  if (!keyword) return escaped;

  const keywords = String(keyword)
    .split(/\s+/)
    .filter((k) => k);
  let result = escaped;
  for (const kw of keywords) {
    // 关键词同样转义后再匹配，与转义后的文本保持一致
    const regex = new RegExp(escapeRegExp(escapeHtml(kw)), 'gi');
    result = result.replace(regex, (match) => `<span class="${highlightClass}">${match}</span>`);
  }
  return result;
};
