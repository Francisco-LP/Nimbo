// ============================================================
// Nimbo · Inicialización y estado global
// ============================================================

const App = {
  state: {
    config: null,
  },

  init() {
    this.loadConfig();
    Files.init();
    Upload.init();
    ConfigPanel.init();
    LogsPanel.init();
    this.wireEvents();

    // Actualización automática de la lista cada 30 segundos.
    setInterval(() => Files.refresh(true), 30000);
  },

  // Carga la configuración (necesaria para validar tamaños y mostrar espacio).
  async loadConfig() {
    try {
      this.state.config = await API.getConfig();
      Files.updateCapacity(this.state.config);
    } catch (_) {
      Notify.show('warning', 'No se pudo leer la configuración del servidor');
    }
  },

  wireEvents() {
    document.getElementById('btn-config').addEventListener('click', () => ConfigPanel.open());
    document.getElementById('btn-logs').addEventListener('click', () => LogsPanel.open());
    document.getElementById('btn-upload').addEventListener('click', () => Upload.open());
    document.getElementById('btn-download-zip').addEventListener('click', () => Files.downloadZip());

    // Cierre de modales: botón ✕, clic en el fondo o tecla Escape.
    document.querySelectorAll('[data-close]').forEach((btn) => {
      btn.addEventListener('click', () => closeModal(btn.dataset.close));
    });
    document.querySelectorAll('.modal-overlay').forEach((m) => {
      m.addEventListener('click', (e) => {
        if (e.target === m) closeModal(m.id);
      });
    });
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') closeAllModals();
    });
  },
};

// ---- Abrir / cerrar modales (funciones globales) ----

function openModal(id) {
  document.getElementById(id).classList.add('modal-open');
}

function closeModal(id) {
  document.getElementById(id).classList.remove('modal-open');
}

function closeAllModals() {
  document.querySelectorAll('.modal-overlay').forEach((m) => {
    m.classList.remove('modal-open');
  });
}

// Arranque cuando el DOM está listo.
document.addEventListener('DOMContentLoaded', () => App.init());
