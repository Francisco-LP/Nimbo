// ============================================================
// Nimbo · Sistema de notificaciones (toasts)
// ============================================================

const Notify = (() => {
  const TYPES = {
    success: { icon: '✅', cls: 'toast-success' },
    error: { icon: '❌', cls: 'toast-error' },
    warning: { icon: '⚠️', cls: 'toast-warning' },
    info: { icon: 'ℹ️', cls: 'toast-info' },
  };

  const MAX_VISIBLE = 3;
  const DURATION = 5000; // auto-cierre en 5 segundos

  // Muestra un toast. Si ya hay 3 visibles, se elimina el más antiguo.
  function show(type, message) {
    const t = TYPES[type] || TYPES.info;
    const container = document.getElementById('toast-container');

    while (container.children.length >= MAX_VISIBLE) {
      container.firstElementChild.remove();
    }

    const toast = document.createElement('div');
    toast.className = 'toast ' + t.cls;
    toast.innerHTML =
      '<span class="toast-icon"></span>' +
      '<span class="toast-msg"></span>' +
      '<button class="toast-close" title="Cerrar">✕</button>';
    toast.querySelector('.toast-icon').textContent = t.icon;
    toast.querySelector('.toast-msg').textContent = message;

    const timer = setTimeout(() => remove(toast), DURATION);
    toast.querySelector('.toast-close').addEventListener('click', () => {
      clearTimeout(timer);
      remove(toast);
    });

    container.appendChild(toast);
    requestAnimationFrame(() => toast.classList.add('toast-visible'));
  }

  // Elimina un toast con una pequeña transición de salida.
  function remove(toast) {
    if (toast.dataset.removing) return;
    toast.dataset.removing = '1';
    toast.classList.remove('toast-visible');
    setTimeout(() => toast.remove(), 300);
  }

  return { show };
})();
