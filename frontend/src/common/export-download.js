// exportBaseName returns a portable last path component for an original filename.
export function exportBaseName(source) {
  const name = (source.originalName || source.name).split(/[\\/]/).pop();
  const cleaned = name
    .replace(/[^\p{L}\p{M}\p{N}\p{P}\p{S} ]/gu, "")
    .replace(/[~\\:|"?*<>{}]/g, "")
    .replace(/\.{2,}/g, ".")
    .trim();
  return cleaned && cleaned !== "." ? cleaned : "photo";
}

// splitExtension separates the filename stem and last extension.
function splitExtension(name) {
  const dot = name.lastIndexOf(".");
  return dot > 0 ? [name.slice(0, dot), name.slice(dot)] : [name, ""];
}

// buildExportDownloads snapshots selected outputs in group, original, preview, and size order.
export function buildExportDownloads(groups, selectedIds) {
  const selected = new Set(selectedIds);
  const usedNames = new Set();
  const downloads = [];
  const order = { original: 0, preview: 1, thumbnail: 2, avc: 3, "video-part": 3 };

  for (const group of groups) {
    const originalName = exportBaseName(group.original);
    const [originalStem] = splitExtension(originalName);
    const stem = originalStem.trim();
    const options = group.options.filter((option) => selected.has(option.id)).sort((a, b) => {
      return order[a.kind] - order[b.kind] || (a.kind === "thumbnail" ? (a.width || 0) * (a.height || 0) - (b.width || 0) * (b.height || 0) : 0);
    });
    for (const option of options) {
      const source = Object.freeze({ ...option.source });
      const sourceExtension = splitExtension(source.name.split(/[\\/]/).pop())[1] || `.${source.format}`;
      const outputStem = option.kind === "original" ? splitExtension(exportBaseName(source))[0].trim() : stem;
      const outputSuffix = { preview: "preview", avc: "avc", "video-part": "video" }[option.kind] || option.size;
      const name = exportBaseName({ name: option.kind === "original" ? `${outputStem}${sourceExtension}` : `${outputStem}_${outputSuffix}.${option.format}` });
      const [base, extension] = splitExtension(name);
      let filename = name;
      let suffix = 2;
      while (usedNames.has(filename.toLowerCase())) {
        filename = `${base}_${suffix++}${extension}`;
      }
      usedNames.add(filename.toLowerCase());
      downloads.push(Object.freeze({ ...option, source, filename }));
    }
  }
  return Object.freeze(downloads);
}

// exportDownloadUrl builds an attachment URL with the current session's source-specific token.
export function exportDownloadUrl(item, config) {
  const hash = encodeURIComponent(item.source.hash);
  if (item.kind === "thumbnail") {
    return `${config.apiUri}/t/${hash}/${encodeURIComponent(config.previewToken)}/${encodeURIComponent(item.size)}?download=1&name=original`;
  }
  const format = ["avc", "video-part"].includes(item.kind) ? "&format=avc" : "";
  return `${config.apiUri}/dl/${hash}?t=${encodeURIComponent(config.downloadToken)}&name=original${format}`;
}

// downloadDirect hands a server attachment to the browser without buffering it in JavaScript.
export function downloadDirect(url, filename) {
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.hidden = true;
  document.body.appendChild(link);
  link.click();
  link.remove();
}

// ExportQueue spaces browser handoffs and retries failed outputs without repeating successful ones.
export class ExportQueue {
  // constructor captures immutable outputs and live authorization callbacks.
  constructor(items, { refresh, allowed, config, transfer = downloadDirect, progress = () => {}, change = () => {} }) {
    this.items = items;
    this.refresh = refresh;
    this.allowed = allowed;
    this.config = config;
    this.transfer = transfer;
    this.progress = progress;
    this.change = change;
    this.states = new Map(items.map((item) => [item.id, { state: "ready" }]));
    this.cancelled = false;
    this.started = false;
    this.handed = 0;
    this.lastHandoff = null;
    this.timer = null;
    this.resume = null;
    this.running = null;
    this.retrying = new Set();
  }

  // active checks cancellation and the current dialog permission.
  active() {
    return !this.cancelled && this.allowed();
  }

  // update publishes output state only while the dialog remains active.
  update(item, state) {
    if (this.active()) {
      this.states.set(item.id, { state });
      this.change(Object.fromEntries(this.states));
    }
  }

  // cancel releases a pending interval and prevents further browser handoffs.
  cancel() {
    this.cancelled = true;
    clearTimeout(this.timer);
    this.resume?.();
    this.resume = null;
  }

  // wait pauses between browser handoffs until the interval expires or cancellation occurs.
  wait(milliseconds) {
    return new Promise((resolve) => {
      this.resume = resolve;
      this.timer = setTimeout(() => {
        this.resume = null;
        resolve();
      }, milliseconds);
    });
  }

  // pump transfers ready outputs in fixed order after refreshing the current session.
  async pump(refresh) {
    try {
      if (refresh) {
        await this.refresh();
      }
    } catch {
      for (const item of this.items) {
        if (this.states.get(item.id).state === "ready") {
          this.update(item, "error");
        }
      }
      return this.handed;
    }
    while (this.active()) {
      const ready = this.items.find((item) => this.states.get(item.id).state === "ready");
      if (!ready) {
        break;
      }
      const remaining = this.lastHandoff === null ? 0 : 250 - (Date.now() - this.lastHandoff);
      if (remaining > 0) {
        await this.wait(remaining);
        continue;
      }
      try {
        this.transfer(exportDownloadUrl(ready, this.config), ready.filename);
        this.lastHandoff = Date.now();
        this.update(ready, "handed");
        this.progress(++this.handed);
      } catch {
        this.update(ready, "error");
      }
    }
    return this.handed;
  }

  // execute serializes runners so a retry also runs after a finishing handoff loop.
  execute(refresh = true) {
    this.running = this.running ? this.running.then(() => this.pump(refresh)) : this.pump(refresh);
    return this.running;
  }

  // run starts the fixed list once and publishes its initial states.
  run() {
    if (this.started || !this.active()) {
      return Promise.resolve(this.handed);
    }
    this.started = true;
    this.change(Object.fromEntries(this.states));
    return this.execute();
  }

  // retry refreshes authorization and queues one failed output behind any active runner.
  async retry(item) {
    if (!this.active() || this.states.get(item.id)?.state !== "error" || this.retrying.has(item.id)) {
      return this.handed;
    }
    this.retrying.add(item.id);
    try {
      await this.refresh();
      if (this.active()) {
        this.update(item, "ready");
        return await this.execute(false);
      }
      return this.handed;
    } catch {
      this.update(item, "error");
      return this.handed;
    } finally {
      this.retrying.delete(item.id);
    }
  }
}
