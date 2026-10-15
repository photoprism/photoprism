import { exportMediaUrl, prepareExportMedia } from "common/export-media";

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
  const filename = encodeURIComponent(item.filename);
  const hash = encodeURIComponent(item.source.hash);
  if (["avc", "video-part"].includes(item.kind)) {
    return `${exportMediaUrl(item.source, config, "download", item.kind)}&filename=${filename}`;
  }
  if (item.kind === "thumbnail") {
    return `${config.contentUri}/t/${hash}/${encodeURIComponent(config.previewToken)}/${encodeURIComponent(item.size)}?download=1&filename=${filename}`;
  }
  return `${config.apiUri}/dl/${hash}?t=${encodeURIComponent(config.downloadToken)}&filename=${filename}`;
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

// ExportQueue prepares one output at a time while spacing ready browser handoffs.
export class ExportQueue {
  // constructor captures immutable outputs and live authorization callbacks.
  constructor(items, { refresh, allowed, config, transfer = downloadDirect, prepare = prepareExportMedia, progress = () => {}, change = () => {} }) {
    this.items = items;
    this.refresh = refresh;
    this.allowed = allowed;
    this.config = config;
    this.transfer = transfer;
    this.prepare = prepare;
    this.progress = progress;
    this.change = change;
    this.states = new Map(items.map((item) => [item.id, { state: ["avc", "video-part"].includes(item.kind) && !item.ready ? "waiting" : "ready" }]));
    this.cancelled = false;
    this.started = false;
    this.handed = 0;
    this.lastHandoff = null;
    this.timer = null;
    this.resume = null;
    this.preparing = null;
    this.running = null;
    this.retrying = new Set();
  }

  // active checks cancellation and the current dialog permission.
  active() {
    return !this.cancelled && this.allowed();
  }

  // update publishes output state only while the dialog remains active.
  update(item, state, result = null) {
    if (!this.active()) {
      return;
    }
    this.states.set(item.id, { state, result });
    this.change(Object.fromEntries(this.states));
    this.wake();
  }

  // wake resumes a waiting queue after preparation or cancellation.
  wake() {
    clearTimeout(this.timer);
    this.resume?.();
    this.resume = null;
  }

  // cancel stops client work while a started server preparation can publish its sidecar.
  cancel() {
    this.cancelled = true;
    this.wake();
  }

  // wait pauses until a preparation completes or the handoff interval expires.
  wait(milliseconds) {
    return new Promise((resolve) => {
      this.resume = resolve;
      if (milliseconds !== undefined) {
        this.timer = setTimeout(() => {
          this.resume = null;
          resolve();
        }, milliseconds);
      }
    });
  }

  // prepareOne refreshes tokens and records a single preparation result or failure.
  async prepareOne(item) {
    try {
      if (!this.active()) {
        return;
      }
      this.update(item, "preparing");
      await this.refresh();
      if (!this.active()) {
        return;
      }
      const result = await this.prepare(item, this.config);
      this.update(item, "ready", result);
    } catch {
      this.update(item, "error");
    } finally {
      this.preparing = null;
      this.wake();
    }
  }

  // pump transfers ready outputs in fixed order without waiting for conversions.
  async pump(refresh = true) {
    try {
      if (refresh) {
        await this.refresh();
      }
    } catch {
      for (const item of this.items) {
        if (["waiting", "ready"].includes(this.states.get(item.id).state)) {
          this.update(item, "error");
        }
      }
      return this.handed;
    }
    while (this.active()) {
      const next = this.items.find((item) => this.states.get(item.id).state === "waiting");
      if (!this.preparing && next) {
        this.preparing = Promise.resolve().then(() => this.prepareOne(next));
      }
      const ready = this.items.find((item) => this.states.get(item.id).state === "ready");
      if (ready) {
        const remaining = this.lastHandoff === null ? 0 : 250 - (Date.now() - this.lastHandoff);
        if (remaining > 0) {
          await this.wait(remaining);
          continue;
        }
        if (!this.active()) {
          break;
        }
        try {
          this.transfer(exportDownloadUrl(ready, this.config), ready.filename);
          this.lastHandoff = Date.now();
          this.update(ready, "handed", this.states.get(ready.id).result);
          this.progress(++this.handed);
        } catch {
          this.update(ready, "error");
        }
      } else if (this.preparing) {
        await this.wait();
      } else {
        break;
      }
    }
    return this.handed;
  }

  // execute coalesces queue runners including retries submitted during another preparation.
  execute(refresh = true) {
    if (!this.running) {
      this.running = this.pump(refresh).finally(() => {
        this.running = null;
      });
    }
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

  // retry schedules only an errored output and never repeats successful handoffs.
  async retry(item) {
    if (!this.active() || this.states.get(item.id)?.state !== "error" || this.retrying.has(item.id)) {
      return this.handed;
    }
    this.retrying.add(item.id);
    try {
      await this.refresh();
      if (!this.active()) {
        return this.handed;
      }
      this.update(item, ["avc", "video-part"].includes(item.kind) ? "waiting" : "ready");
      return this.execute(false);
    } catch {
      this.update(item, "error");
      return this.handed;
    } finally {
      this.retrying.delete(item.id);
    }
  }

  // manual hands a ready output over once or repeats an explicit download of a handed output.
  async manual(item) {
    const initial = this.states.get(item.id)?.state;
    if (!this.active() || !["ready", "handed"].includes(initial)) {
      return;
    }
    await this.refresh();
    const current = this.states.get(item.id);
    if (!this.active() || (initial === "ready" && current.state !== "ready")) {
      return;
    }
    this.transfer(exportDownloadUrl(item, this.config), item.filename);
    this.lastHandoff = Date.now();
    if (current.state === "ready") {
      this.update(item, "handed", current.result);
      this.progress(++this.handed);
    }
  }
}
