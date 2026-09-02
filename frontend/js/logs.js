// ============================================================
// Nimbo · Visualización de logs con filtros
// ============================================================

const LogsPanel = (() => {
  let entries = [];        // últimos logs cargados
  let currentFilter = 'ALL';

  const el = {
    list: document.getElementById('logs-list'),
    refresh: document.getElementById('btn-logs-refresh'),
  };

  function init() {
    el.refresh.addEventListener('click', load);
    document.querySelectorAll('.log-filter').forEach((btn) => {
      btn.addEventListener('click', () => setFilter(btn.dataset.level));
    });
  }

  async function open() {
    openModal('logs-modal');
    await load();
  }

  // Trae los últimos logs y los renderiza.
  async function load() {
    try {
      const data = await API.getLogs(100);
      entries = data.logs || [];
      render();
    } catch (_) {
      Notify.show('error', 'No se pudieron cargar los logs');
    }
  }

  function setFilter(level) {
    currentFilter = level;
    document.querySelectorAll('.log-filter').forEach((b) => {
      b.classList.toggle('active', b.dataset.level === level);
    });
    render();
  }

  function render() {
    el.list.innerHTML = '';

    const filtered = currentFilter === 'ALL'
      ? entries
      : entries.filter((e) => (e.level || 'INFO') === currentFilter);

    if (filtered.length === 0) {
      const li = document.createElement('li');
      li.className = 'log-empty';
      li.textContent = 'No hay logs pa mostrar con ese filtro';
      el.list.appendChild(li);
      return;
    }

    filtered.forEach((e) => {
      const level = e.level || 'INFO';
      const li = document.createElement('li');
      li.className = 'log-line log-' + level.toLowerCase();
      li.innerHTML =
        '<span class="log-time">' + Utils.esc(e.timestamp) + '</span>' +
        '<span class="log-level">' + Utils.esc(level) + '</span>' +
        '<span class="log-event">' + Utils.esc(e.event) + '</span>' +
        '<span class="log-msg">' + Utils.esc(e.message) + '</span>';
      el.list.appendChild(li);
    });
  }

  return { init, open };
})();
