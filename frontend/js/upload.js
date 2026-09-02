// ============================================================
// Nimbo · Subida de archivos con barra de progreso
// ============================================================

const Upload = (() => {
  const state = {
    items: [],      // cola de archivos
    active: false,  // hay una subida en curso
    cancelled: false,
    xhr: null,      // petición activa (para poder cancelarla)
  };

  const el = {
    modal: document.getElementById('upload-modal'),
    drop: document.getElementById('drop-zone'),
    input: document.getElementById('file-input'),
    queue: document.getElementById('upload-queue'),
    start: document.getElementById('btn-upload-start'),
    cancel: document.getElementById('btn-upload-cancel'),
  };

  function init() {
    // Drag & drop.
    el.drop.addEventListener('click', () => el.input.click());
    el.drop.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); el.input.click(); }
    });
    el.drop.addEventListener('dragover', (e) => {
      e.preventDefault();
      el.drop.classList.add('drag-over');
    });
    el.drop.addEventListener('dragleave', () => el.drop.classList.remove('drag-over'));
    el.drop.addEventListener('drop', (e) => {
      e.preventDefault();
      el.drop.classList.remove('drag-over');
      if (e.dataTransfer.files.length) addFiles(e.dataTransfer.files);
    });
    el.input.addEventListener('change', () => {
      if (el.input.files.length) addFiles(el.input.files);
      el.input.value = '';
    });
    el.start.addEventListener('click', start);
    el.cancel.addEventListener('click', cancel);
  }

  function open() {
    state.items = [];
    state.active = false;
    state.cancelled = false;
    render();
    showModal();
  }

  // Agrega archivos a la cola validando su tamaño.
  function addFiles(fileList) {
    const maxSize = (App.state.config && App.state.config.maxFileSize) || Infinity;
    for (const file of fileList) {
      if (file.size > maxSize) {
        state.items.push({
          name: file.name,
          size: file.size,
          progress: 0,
          status: 'error',
          error: 'Supera el tamaño máximo permitido',
          file: null,
        });
        Notify.show('error', '«' + file.name + '» supera el tamaño máximo');
        continue;
      }
      state.items.push({
        name: file.name,
        size: file.size,
        progress: 0,
        status: 'pending',
        error: null,
        file,
      });
    }
    render();
  }

  // Procesa la cola de forma secuencial para no saturar el servidor.
  async function start() {
    if (state.active) return;
    if (state.items.length === 0) {
      Notify.show('warning', 'Agrega archivos primero');
      return;
    }
    state.active = true;
    state.cancelled = false;
    el.start.disabled = true;

    for (const item of state.items) {
      if (item.status === 'error' || item.status === 'done' || item.status === 'cancelled') continue;
      await uploadItem(item);
      if (state.cancelled) break;
    }

    state.active = false;
    el.start.disabled = false;

    const done = state.items.filter((i) => i.status === 'done').length;
    const failed = state.items.filter((i) => i.status === 'error').length;

    if (done > 0) {
      Notify.show('success', 'Se subieron ' + done + ' archivo(s) con éxito');
    }
    if (failed > 0) {
      Notify.show('error', failed + ' archivo(s) no se pudieron subir');
    }

    Files.refresh();
    if (done > 0) setTimeout(close, 1200);
  }

  // Sube un archivo usando XMLHttpRequest (el único con evento de progreso).
  function uploadItem(item) {
    return new Promise((resolve) => {
      item.status = 'uploading';
      render();

      const form = new FormData();
      form.append('file', item.file);

      const xhr = new XMLHttpRequest();
      state.xhr = xhr;
      xhr.open('POST', '/api/upload');

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          item.progress = Math.round((e.loaded / e.total) * 100);
          updateItemProgress(item);
        }
      };

      xhr.onload = () => {
        let data = { success: false };
        try { data = JSON.parse(xhr.responseText); } catch (_) { /* respuesta no JSON */ }

        item.progress = 100;
        if (data.success && data.files && data.files.length) {
          item.status = 'done';
        } else {
          item.status = 'error';
          item.error = (data.errors && data.errors[0] && data.errors[0].msg) || 'No se pudo subir';
        }
        updateItemProgress(item);
        render();
        resolve();
      };

      xhr.onerror = () => {
        item.status = 'error';
        item.error = 'Error de red';
        render();
        resolve();
      };

      xhr.onabort = () => {
        item.status = 'cancelled';
        render();
        resolve();
      };

      xhr.send(form);
    });
  }

  // Cancela la subida en curso.
  function cancel() {
    if (!state.active && state.items.length === 0) {
      close();
      return;
    }
    state.cancelled = true;
    if (state.xhr) {
      try { state.xhr.abort(); } catch (_) { /* noop */ }
    }
    state.items.forEach((i) => {
      if (i.status === 'pending') i.status = 'cancelled';
    });
    state.active = false;
    el.start.disabled = false;
    render();
  }

  // ---- Render ----

  function render() {
    el.queue.innerHTML = '';
    el.start.disabled = state.items.length === 0 || state.active;

    state.items.forEach((item) => {
      const li = document.createElement('li');
      li.className = 'queue-item';

      const head = document.createElement('div');
      head.className = 'q-head';

      const name = document.createElement('span');
      name.className = 'q-name';
      name.textContent = item.name;

      const status = document.createElement('span');
      status.className = 'q-status';
      status.textContent = statusText(item);

      head.appendChild(name);
      head.appendChild(status);
      li.appendChild(head);

      const bar = document.createElement('div');
      bar.className = 'q-bar';
      const fill = document.createElement('div');
      fill.className = 'q-bar-fill ' + (item.status === 'done' ? 'done' : item.status === 'error' ? 'error' : '');
      fill.style.width = item.progress + '%';
      bar.appendChild(fill);
      li.appendChild(bar);

      item.el = { status, fill };
      el.queue.appendChild(li);
    });
  }

  function statusText(item) {
    switch (item.status) {
      case 'done': return '✅ Listo';
      case 'error': return '❌ ' + (item.error || 'Error');
      case 'cancelled': return '⏹️ Cancelado';
      case 'uploading': return item.progress + '% · ' + Utils.formatSize(item.size);
      default: return 'Esperando · ' + Utils.formatSize(item.size);
    }
  }

  // Actualiza solo la barra y el texto del estado sin re-renderizar todo.
  function updateItemProgress(item) {
    if (item.el) {
      item.el.fill.style.width = item.progress + '%';
      item.el.status.textContent = statusText(item);
    }
  }

  function showModal() { document.getElementById('upload-modal').classList.add('modal-open'); }
  function close() { document.getElementById('upload-modal').classList.remove('modal-open'); }

  return { init, open, close };
})();
