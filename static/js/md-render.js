/**
 * 助手回复：Markdown → 安全 HTML（marked + DOMPurify）
 */
(function (global) {
  function escapeHtml(s) {
    if (s == null || s === '') return '';
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  var MD_TAGS = [
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'p', 'ul', 'ol', 'li', 'strong', 'b', 'em', 'i',
    'code', 'pre', 'blockquote', 'a', 'br', 'hr',
    'table', 'thead', 'tbody', 'tr', 'th', 'td',
    'del', 'span',
  ];

  function renderAssistantMarkdown(text) {
    text = text || '';
    if (typeof marked !== 'undefined' && typeof DOMPurify !== 'undefined') {
      try {
        if (marked.setOptions) {
          marked.setOptions({ breaks: true, gfm: true });
        }
        var html = typeof marked.parse === 'function' ? marked.parse(text) : marked(text);
        return DOMPurify.sanitize(html, {
          ALLOWED_TAGS: MD_TAGS,
          ALLOWED_ATTR: ['href', 'title', 'class', 'colspan', 'rowspan', 'target', 'rel'],
        });
      } catch (e) {
        console.warn('markdown render', e);
      }
    }
    return '<p class="md-fallback">' + escapeHtml(text).replace(/\n/g, '<br>') + '</p>';
  }

  global.escapeHtml = escapeHtml;
  global.renderAssistantMarkdown = renderAssistantMarkdown;
})(typeof window !== 'undefined' ? window : this);
