// ============================================================
// Nimbo · Cliente API y utilidades compartidas
// ============================================================

// Utilidades compartidas por todos los módulos.
const Utils = {
  // Escapa texto para insertarlo de forma segura en HTML (evita XSS).
  esc(s) {
    const div = document.createElement('div');
    div.textContent = String(s ?? '');
    return div.innerHTML;
  },

  // Escapa un valor para usarlo dentro de atributos HTML.
  escAttr(s) {
    return Utils.esc(s).replace(/"/g, '&quot;');
  },

  // Formatea un tamaño en bytes a una unidad legible.
  formatSize(bytes) {
    if (!Number.isFinite(bytes) || bytes < 0) return '—';
    if (bytes < 1024) return bytes + ' B';
    const units = ['KB', 'MB', 'GB', 'TB'];
    let v = bytes;
    let u = -1;
    do { v /= 1024; u++; } while (v >= 1024 && u < units.length - 1);
    return v.toFixed(v >= 100 ? 0 : 1) + ' ' + units[u];
  },

  // Icono emoji según la extensión del archivo.
  iconFor(name) {
    const ext = String(name).split('.').pop().toLowerCase();
    if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) return '🖼️';
    if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return '🗜️';
    if (['mp3', 'wav', 'ogg', 'flac', 'm4a'].includes(ext)) return '🎵';
    if (['mp4', 'mkv', 'avi', 'mov', 'webm'].includes(ext)) return '🎬';
    if (['pdf'].includes(ext)) return '📕';
    if (['doc', 'docx', 'odt', 'txt', 'md'].includes(ext)) return '📄';
    if (['xls', 'xlsx', 'csv', 'ods'].includes(ext)) return '📊';
    if (['ppt', 'pptx', 'odp'].includes(ext)) return '📽️';
    return '📄';
  },
};

// Cliente HTTP de la API REST de Nimbo.
const API = {
  // GET /api/files
  async listFiles() {
    const res = await fetch('/api/files');
    if (!res.ok) throw new Error('error al listar archivos');
    return res.json();
  },

  // GET /api/config
  async getConfig() {
    const res = await fetch('/api/config');
    if (!res.ok) throw new Error('error al leer la configuración');
    return res.json();
  },

  // POST /api/config
  async saveConfig(data) {
    const res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return res.json();
  },

  // DELETE /api/delete/{name}
  async deleteFile(name) {
    const res = await fetch('/api/delete/' + encodeURIComponent(name), { method: 'DELETE' });
    return res.json();
  },

  // GET /api/logs?limit=n
  async getLogs(limit = 100) {
    const res = await fetch('/api/logs?limit=' + limit);
    if (!res.ok) throw new Error('error al leer los logs');
    return res.json();
  },

  // POST /api/download-zip → descarga un ZIP con los archivos seleccionados.
  async downloadZip(files) {
    const res = await fetch('/api/download-zip', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ files }),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.message || 'No se pudo crear el ZIP');
    }
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'nimbo.zip';
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  },

  // URL de descarga individual (se usa en un <a> nativo).
  downloadUrl(name) {
    return '/api/download/' + encodeURIComponent(name);
  },
};
