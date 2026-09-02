// ============================================================
// Nimbo · Lista de archivos y acciones
// ============================================================

const Files = (() => {
  const state = {
    files: [],          // todos los archivos del servidor
    filtered: [],       // archivos tras aplicar la búsqueda
    selected: new Set(), // nombres seleccionados
    search: '',
  };

  const el = {
    tbody: document.getElementById('file-tbody'),
    empty: document.getElementById('empty-state'),
    search: document.getElementById('search-input'),
    selectAll: document.getElementById('select-all'),
    count: document.getElementById('file-count'),
    used: document.getElementById('used-space'),
    max: document.getElementById('max-space'),
    btnZip: document.getElementById('btn-download-zip'),
  };

  function init() {
    el.search.addEventListener('input', () => setSearch(el.search.value));
    el.selectAll.addEventListener('change', toggleSelectAll);
    el.tbody.addEventListener('click', onTbodyClick);
    el.tbody.addEventListener('change', onTbodyChange);
    refresh();
  }

  // Carga la lista de archivos desde el servidor.
  function refresh(silent = false) {
    API.listFiles()
      .then((data) => {
        state.files = data.files || [];
        applyFilter();
      })
      .catch(() => {
        if (!silent) Notify.show('error', 'No se pudo cargar la lista de archivos');
      });
  }

  function setSearch(value) {
    state.search = value.trim().toLowerCase();
    applyFilter();
  }

  function applyFilter() {
    if (!state.search) {
      state.filtered = state.files;
    } else {
      state.filtered = state.files.filter((f) =>
        f.name.toLowerCase().includes(state.search)
      );
    }
    // Quitar de la selección archivos que ya no se ven.
    const visible = new Set(state.filtered.map((f) => f.name));
    for (const name of state.selected) {
      if (!visible.has(name)) state.selected.delete(name);
    }
    render();
  }

  function render() {
    el.tbody.innerHTML = '';
    const files = state.filtered;

    if (files.length === 0) {
      el.empty.classList.remove('hidden');
    } else {
      el.empty.classList.add('hidden');
      files.forEach((f) => {
        const checked = state.selected.has(f.name);
        const tr = document.createElement('tr');
        tr.innerHTML =
          '<td><input type="checkbox" data-name="' + Utils.escAttr(f.name) + '"' + (checked ? ' checked' : '') + '></td>' +
          '<td class="file-name"><span class="file-icon">' + Utils.iconFor(f.name) + '</span>' +
          '<span class="file-label" title="' + Utils.escAttr(f.name) + '">' + Utils.esc(f.name) + '</span></td>' +
          '<td class="file-size">' + Utils.formatSize(f.size) + '</td>' +
          '<td class="file-date">' + Utils.esc(f.modTime || '') + '</td>' +
          '<td class="file-actions">' +
          '<button class="icon-btn" data-action="download" data-name="' + Utils.escAttr(f.name) + '" title="Descargar">⬇️</button>' +
          '<button class="icon-btn danger" data-action="delete" data-name="' + Utils.escAttr(f.name) + '" title="Eliminar">🗑️</button>' +
          '</td>';
        el.tbody.appendChild(tr);
      });
    }

    updateStatusBar();
    updateSelectAll();
    updateZipButton();
  }

  function updateStatusBar() {
    const total = state.files.length;
    const used = state.files.reduce((acc, f) => acc + (f.size || 0), 0);
    el.count.textContent = total === 1 ? '1 archivo' : total + ' archivos';
    el.used.textContent = Utils.formatSize(used) + ' en uso';
  }

  // Muestra el espacio disponible según la configuración cargada.
  function updateCapacity(cfg) {
    if (cfg && cfg.maxFileSize) {
      el.max.textContent = Utils.formatSize(cfg.maxFileSize) + ' de espacio';
    }
  }

  function updateSelectAll() {
    const files = state.filtered;
    const totalSelected = files.filter((f) => state.selected.has(f.name)).length;
    el.selectAll.checked = files.length > 0 && totalSelected === files.length;
    el.selectAll.indeterminate = totalSelected > 0 && totalSelected < files.length;
  }

  function updateZipButton() {
    el.btnZip.disabled = state.selected.size === 0;
  }

  // Clics dentro de la tabla (descargar / eliminar).
  function onTbodyClick(e) {
    const btn = e.target.closest('button[data-action]');
    if (!btn) return;
    const name = btn.dataset.name;

    if (btn.dataset.action === 'download') {
      downloadOne(name);
    } else if (btn.dataset.action === 'delete') {
      removeFile(name);
    }
  }

  // Cambios de checkbox dentro de la tabla.
  function onTbodyChange(e) {
    const cb = e.target.closest('input[data-name]');
    if (!cb) return;
    if (cb.checked) {
      state.selected.add(cb.dataset.name);
    } else {
      state.selected.delete(cb.dataset.name);
    }
    updateSelectAll();
    updateZipButton();
  }

  function toggleSelectAll() {
    const files = state.filtered;
    const allSelected = files.length > 0 && files.every((f) => state.selected.has(f.name));
    files.forEach((f) => {
      if (allSelected) state.selected.delete(f.name);
      else state.selected.add(f.name);
    });
    render();
  }

  // Descarga individual: un <a> nativo con el nombre del archivo.
  function downloadOne(name) {
    const a = document.createElement('a');
    a.href = API.downloadUrl(name);
    a.setAttribute('download', name);
    document.body.appendChild(a);
    a.click();
    a.remove();
  }

  // Descarga en ZIP de todos los archivos seleccionados.
  function downloadZip() {
    const files = [...state.selected];
    if (files.length === 0) {
      Notify.show('warning', 'Selecciona al menos un archivo pa bajar el ZIP');
      return;
    }
    Notify.show('info', 'Preparando el ZIP...');
    API.downloadZip(files)
      .then(() => Notify.show('success', 'ZIP listo, se está bajando'))
      .catch((err) => Notify.show('error', err.message || 'No se pudo crear el ZIP'));
  }

  // Elimina un archivo con confirmación.
  function removeFile(name) {
    const ok = confirm('¿Seguro que querís eliminar «' + name + '»?\nEsta acción no se puede deshacer.');
    if (!ok) return;
    API.deleteFile(name)
      .then((data) => {
        if (data.success) {
          state.selected.delete(name);
          Notify.show('success', 'Se eliminó «' + name + '»');
          refresh();
        } else {
          Notify.show('error', data.message || 'No se pudo eliminar el archivo');
        }
      })
      .catch(() => Notify.show('error', 'No se pudo eliminar el archivo'));
  }

  return { init, refresh, updateCapacity, downloadZip };
})();
