// ============================================================
// Nimbo · Panel de configuración
// ============================================================

const ConfigPanel = (() => {
  const el = {
    storage: document.getElementById('cfg-storage'),
    maxSize: document.getElementById('cfg-maxsize'),
    port: document.getElementById('cfg-port'),
    save: document.getElementById('btn-config-save'),
  };

  function init() {
    el.save.addEventListener('click', save);
    // Guardar con Enter dentro del formulario.
    document.getElementById('config-form').addEventListener('submit', (e) => {
      e.preventDefault();
      save();
    });
  }

  // Carga la configuración actual y abre el modal.
  async function open() {
    try {
      const cfg = await API.getConfig();
      el.storage.value = cfg.storagePath || '';
      el.maxSize.value = Math.round((cfg.maxFileSize || 0) / (1024 * 1024));
      el.port.value = cfg.port || '';
      openModal('config-modal');
    } catch (_) {
      Notify.show('error', 'No se pudo cargar la configuración');
    }
  }

  async function save() {
    const storage = el.storage.value.trim();
    const mb = parseFloat(el.maxSize.value);
    const port = el.port.value.trim();

    if (!storage || !port || !(mb > 0)) {
      Notify.show('error', 'Revisa los campos: todos son obligatorios');
      return;
    }

    const payload = {
      storagePath: storage,
      maxFileSize: Math.round(mb * 1024 * 1024),
      port,
    };

    try {
      const res = await API.saveConfig(payload);
      if (res.success) {
        Notify.show('success', res.message || 'Configuración guardada');
        // Actualizar el espacio disponible mostrado en la barra de estado.
        if (App.state.config) App.state.config.maxFileSize = payload.maxFileSize;
        Files.updateCapacity(App.state.config);
        closeModal('config-modal');
      } else {
        Notify.show('error', res.message || 'No se pudo guardar la configuración');
      }
    } catch (_) {
      Notify.show('error', 'No se pudo guardar la configuración');
    }
  }

  return { init, open };
})();
