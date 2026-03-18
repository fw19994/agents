(function () {
  const listEl = document.getElementById('session-list');
  const btnNew = document.getElementById('btn-new-session');
  if (!listEl) return;

  function renderList(sessions, currentId) {
    currentId = currentId || window.currentSessionId || null;
    listEl.innerHTML = '';
    if (!sessions || !sessions.length) {
      listEl.innerHTML = '<p class="text-slate-400 text-xs p-2">暂无会话</p>';
      return;
    }
    sessions.forEach(function (s) {
      const isActive = s.id === currentId;
      const div = document.createElement('div');
      div.className = 'session-item flex items-center gap-2 rounded-lg px-2 py-2 text-left text-sm cursor-pointer ' + (isActive ? 'bg-indigo-50 text-indigo-700' : 'hover:bg-slate-50 text-slate-700');
      div.dataset.sessionId = s.id;
      const title = (s.title || '新会话').trim() || '新会话';
      const t = document.createElement('span');
      t.className = 'flex-1 truncate';
      t.textContent = title;
      div.appendChild(t);
      const del = document.createElement('button');
      del.type = 'button';
      del.className = 'opacity-0 hover:opacity-100 text-slate-400 hover:text-red-500 p-0.5';
      del.innerHTML = '<i class="fas fa-trash-alt text-xs"></i>';
      del.title = '删除';
      div.appendChild(del);
      listEl.appendChild(div);

      div.addEventListener('click', function (e) {
        if (e.target.closest('button')) return;
        selectSession(s.id);
      });
      del.addEventListener('click', function (e) {
        e.stopPropagation();
        deleteOne(s.id);
      });
    });
  }

  function refreshList() {
    if (typeof listSessions !== 'function') return;
    listSessions(50).then(function (sessions) {
      renderList(sessions);
    });
  }

  function selectSession(id) {
    window.currentSessionId = id;
    if (typeof onCurrentSessionChange === 'function') onCurrentSessionChange(id);
    if (typeof getSessionDetail !== 'function') return;
    getSessionDetail(id).then(function (data) {
      if (!data || !data.messages) return;
      if (typeof window.renderHistory === 'function') window.renderHistory(data.messages);
      refreshList();
    });
  }

  function deleteOne(id) {
    if (typeof deleteSession !== 'function') return;
    deleteSession(id).then(function (ok) {
      if (!ok) return;
      if (window.currentSessionId === id) {
        window.currentSessionId = null;
        if (typeof window.clearMessagesAndWelcome === 'function') window.clearMessagesAndWelcome();
      }
      refreshList();
    });
  }

  if (btnNew) {
    btnNew.addEventListener('click', function () {
      window.currentSessionId = null;
      if (typeof onCurrentSessionChange === 'function') onCurrentSessionChange(null);
      if (typeof window.clearMessagesAndWelcome === 'function') window.clearMessagesAndWelcome();
      refreshList();
    });
  }

  window.refreshSessionList = refreshList;
  window.selectSession = selectSession;
  window.onCurrentSessionChange = function () {
    refreshList();
  };

  refreshList();
})();
