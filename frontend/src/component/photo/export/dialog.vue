<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    theme="lightbox"
    :fullscreen="$vuetify.display.smAndDown"
    max-width="560"
    scrollable
    persistent
    class="p-dialog p-photo-export-dialog"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
    @keydown.esc.exact.stop.prevent="close"
    @keydown.stop
  >
    <v-card ref="content" tabindex="-1" color="surface" :rounded="$vuetify.display.smAndDown ? 0 : undefined" class="export-card">
      <v-card-title class="export-title d-flex align-center justify-space-between">
        <span>{{ $gettext("Export") }}</span>
        <v-btn variant="text" icon="mdi-close" height="44" width="44" :title="$gettext('Close')" :aria-label="$gettext('Close')" class="action-dismiss" @click.stop="close"></v-btn>
      </v-card-title>
      <v-divider></v-divider>
      <v-card-text class="export-content">
        <div v-if="loading" role="status" class="text-center pa-4">
          <v-progress-circular indeterminate aria-hidden="true"></v-progress-circular>
          <p>{{ $gettext("Loading…") }}</p>
        </div>
        <v-alert v-else-if="loadError" type="error" variant="tonal">
          {{ $gettext("Failed to load files. Please try again.") }}
          <v-btn class="action-retry mt-2" @click="loadFiles">{{ $gettext("Retry") }}</v-btn>
        </v-alert>
        <div v-else-if="started" class="export-progress">
          <div v-for="item in activeOutputs" :key="item.id" class="export-output" :data-option-id="item.id">
            <div class="export-output-heading">
              <v-icon v-if="states[item.id]?.state === 'error'" icon="mdi-alert-circle-outline" color="error" size="24" aria-hidden="true"></v-icon>
              <v-progress-circular v-else indeterminate size="24" width="2" color="on-surface" aria-hidden="true"></v-progress-circular>
              <span class="export-file">
                <span class="export-filename text-body-2 font-weight-medium">{{ item.filename }}</span>
                <span class="text-caption text-medium-emphasis">{{ optionMetadata(item) }}</span>
                <span class="text-caption" :class="{ 'text-error': states[item.id]?.state === 'error' }" role="status">{{ stateLabel(item) }}</span>
              </span>
              <v-btn v-if="states[item.id]?.state === 'error'" class="action-retry-output" variant="text" color="on-surface" size="small" :aria-label="$gettext('Retry %{filename}', { filename: item.filename })" @click="retry(item)">{{ $gettext("Retry") }}</v-btn>
            </div>
          </div>
        </div>
        <p v-else-if="!groups.length" class="export-empty">{{ $gettext("No supported files available for export.") }}</p>
        <template v-else>
          <v-expansion-panels v-model="expanded" multiple variant="accordion" class="export-groups">
            <v-expansion-panel v-for="group in groups" :key="group.id" :value="group.id" class="export-group rounded-lg">
              <v-expansion-panel-title class="export-group-title">
                <v-img v-if="group.thumbnailSource" :src="thumbnailUrl(group)" width="44" height="44" cover alt="" class="export-thumbnail rounded">
                  <template #error><v-icon :icon="mediaIcon(group)" size="28"></v-icon></template>
                </v-img>
                <v-icon v-else :icon="mediaIcon(group)" size="28" class="export-thumbnail"></v-icon>
                <span class="export-file">
                  <span class="export-filename text-body-2 font-weight-medium">{{ fileName(group.original) }}</span>
                  <span class="export-file-type text-caption">{{ mediaLabel(group) }}</span>
                  <span v-if="fileMetadata(group.original)" class="export-file-metadata text-caption">{{ fileMetadata(group.original) }}</span>
                </span>
                <v-chip v-if="selectedCount(group)" :closable="false" variant="tonal" size="x-small" color="on-surface" :title="$gettext('Selected')" class="export-selected ms-2 me-2">{{ $gettext("%{n} Selected", { n: selectedCount(group) }) }}</v-chip>
              </v-expansion-panel-title>
              <v-expansion-panel-text class="export-group-options">
                <p v-if="group.previewMissing" class="text-body-2 text-error mb-3">{{ $gettext("No preview image available.") }}</p>
                <div v-for="section in optionSections(group)" :key="section.id" class="export-options-section" :class="{ 'export-resized-jpegs': section.id === 'resized' }" :role="section.label ? 'group' : undefined" :aria-label="section.label || undefined">
                  <v-divider v-if="section.label"></v-divider>
                  <p v-if="section.label" class="export-section-title text-caption text-medium-emphasis">{{ section.label }}</p>
                  <div v-for="option in section.options" :key="option.id" class="export-option" :data-option-id="option.id">
                    <v-checkbox v-model="selected" :value="option.id" :disabled="started" hide-details density="compact" color="on-surface">
                      <template #label>
                        <span class="export-option-label">
                          <span class="export-option-description">
                            <span class="text-body-2 font-weight-medium">{{ optionLabel(option) }}</span>
                            <span class="text-caption">{{ optionMetadata(option) }}</span>
                            <span v-if="needsConversion(option)" class="export-conversion-note text-caption text-medium-emphasis">{{ $gettext("Download starts after transcoding.") }}</span>
                          </span>
                          <span v-if="optionBytes(option)" class="export-option-size text-caption">{{ optionBytes(option) }}</span>
                        </span>
                      </template>
                    </v-checkbox>
                  </div>
                </div>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
          <p v-if="selected.length > 1" class="export-hint text-caption">{{ $gettext("Your browser may ask you to allow multiple downloads.") }}</p>
        </template>
      </v-card-text>
      <v-divider></v-divider>
      <div v-if="transferError" class="export-error text-body-2 text-error" role="alert">
        {{ $gettext("Some files could not be exported. Please try again.") }}
      </div>
      <v-card-actions class="action-buttons export-actions">
        <v-btn v-if="started || !$vuetify.display.smAndDown" variant="flat" color="button" height="44" :block="started && $vuetify.display.smAndDown" class="action-close action-cancel" @click.stop="close">{{ started && errors && !pending ? $gettext("Close") : $gettext("Cancel") }}</v-btn>
        <v-btn v-if="!started" variant="flat" color="highlight" height="44" :block="$vuetify.display.smAndDown" class="action-export action-confirm" :disabled="loading || !selected.length || !permitted" @click.stop="confirm">{{ exportLabel }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import { markRaw } from "vue";
import Photo from "model/photo";
import File from "model/file";
import { canExport } from "common/export";
import { buildMediaExportGroups } from "common/export-media";
import { buildExportDownloads, exportBaseName, ExportQueue } from "common/export-download";

export default {
  name: "PPhotoExportDialog",
  props: {
    visible: { type: Boolean, default: false },
    uid: { type: String, default: "" },
  },
  emits: ["close", "leave"],
  // data initializes the dialog selection and output-state snapshots.
  data() {
    return {
      loading: false,
      closing: false,
      loadError: false,
      transferError: false,
      groups: [],
      expanded: [],
      selected: [],
      downloads: [],
      states: {},
      started: false,
      complete: false,
      handed: 0,
      generation: 0,
      queue: null,
    };
  },
  computed: {
    // exportLabel names the action with the number of selected output files.
    exportLabel() {
      return this.selected.length === 1 ? this.$gettext("Export 1 File") : this.$gettext("Export %{n} Files", { n: this.selected.length });
    },
    // activeOutputs lists unfinished files while successful handoffs leave the progress view.
    activeOutputs() {
      return this.downloads.filter((item) => this.states[item.id]?.state !== "handed");
    },
    // pending counts outputs waiting for browser handoff.
    pending() {
      return Object.values(this.states).filter((item) => item.state === "ready").length;
    },
    // errors counts individually retryable outputs.
    errors() {
      return Object.values(this.states).filter((item) => item.state === "error").length;
    },
    // permitted reads the server-filtered Export setting independently of Download.
    permitted() {
      return canExport(this.$config);
    },
  },
  watch: {
    // visible resets selections on opening and invalidates all pending work on closing.
    visible: {
      immediate: true,
      flush: "sync",
      handler(value) {
        if (value) {
          this.reset();
          this.loadFiles();
        } else {
          this.stop();
        }
      },
    },
    // uid closes the export when the Lightbox changes photos.
    uid() {
      if (this.visible) {
        this.close();
      }
    },
    // permitted stops pending transfers when the setting or authorization is revoked.
    permitted(value) {
      if (!value && this.visible) {
        this.close();
      }
    },
  },
  // beforeUnmount releases the focus scope when parent closure skips the leave transition.
  beforeUnmount() {
    this.stop();
    this.$view.leave(this);
  },
  methods: {
    // afterEnter installs the shared dialog focus scope.
    afterEnter() {
      this.$view.enter(this);
    },
    // afterLeave restores the Lightbox focus scope after the overlay disappears.
    afterLeave() {
      this.$view.leave(this);
      this.$emit("leave");
    },
    // stop invalidates late detail responses and cancels pending browser handoffs.
    stop() {
      this.generation++;
      this.queue?.cancel();
    },
    // reset restores the initial selection state for each opening.
    reset() {
      this.stop();
      this.loading = false;
      this.closing = false;
      this.loadError = false;
      this.transferError = false;
      this.groups = [];
      this.expanded = [];
      this.selected = [];
      this.downloads = [];
      this.states = {};
      this.started = false;
      this.complete = false;
      this.handed = 0;
      this.queue = null;
    },
    // close stops pending work while downloads already handed to the browser continue.
    close() {
      if (this.closing || !this.visible) {
        return;
      }
      this.closing = true;
      this.stop();
      this.$emit("close");
    },
    // loadFiles fetches full stack details and discards replies for a closed or replaced dialog.
    async loadFiles() {
      if (!this.visible || !this.uid || !this.permitted) {
        this.close();
        return;
      }
      const generation = ++this.generation;
      const uid = this.uid;
      this.loading = true;
      this.loadError = false;
      try {
        const photo = await new Photo().find(uid);
        if (generation !== this.generation || !this.visible || uid !== this.uid || !this.permitted) {
          return;
        }
        this.groups = buildMediaExportGroups(photo.Files, this.$config.values.thumbs, this.$config.values.thumbLimit);
        this.expanded = this.groups.length ? [this.groups[0].id] : [];
        this.selected = [];
      } catch {
        if (generation === this.generation && this.visible) {
          this.loadError = true;
        }
      } finally {
        if (generation === this.generation) {
          this.loading = false;
        }
      }
    },
    // fileName returns the original's last filename component.
    fileName(source) {
      return exportBaseName(source);
    },
    // selectedCount counts output files selected in a single original group.
    selectedCount(group) {
      return group.options.filter((option) => this.selected.includes(option.id)).length;
    },
    // thumbnailUrl uses the matched raster file rather than the stack's primary image.
    thumbnailUrl(group) {
      return new File({ Hash: group.thumbnailSource.hash }).thumbnailUrl("tile_224");
    },
    // mediaIcon supplies the existing media symbol when a raster preview is unavailable.
    mediaIcon(group) {
      return { video: "mdi-video", live: "mdi-motion-play", animated: "mdi-motion-play", document: "mdi-file-document", vector: "mdi-svg", raw: "mdi-file-image" }[group.original.mediaType] || "mdi-image";
    },
    // mediaLabel names the original format and media type.
    mediaLabel(group) {
      const labels = { video: this.$gettext("Video"), live: this.$gettext("Live Photo"), animated: this.$gettext("Animated"), document: this.$gettext("Document"), vector: this.$gettext("Vector"), raw: this.$gettext("RAW"), image: this.$gettext("Image") };
      return [group.original.format.toUpperCase(), labels[group.original.mediaType]].filter(Boolean).join(" · ");
    },
    // fileMetadata formats only known original dimensions, codec, duration, and size.
    fileMetadata(source) {
      return [source.width && source.height ? `${source.width} × ${source.height}` : "", source.codec, source.duration ? this.$util.formatDuration(source.duration) : "", source.bytes !== null ? this.$util.formatBytes(source.bytes) : ""].filter(Boolean).join(" · ");
    },
    // displayOptions orders presentation without changing the download queue's ordering.
    displayOptions(group) {
      const order = { original: 0, avc: 1, "video-part": 1, preview: 2, thumbnail: 3 };
      return [...group.options].sort((a, b) => order[a.kind] - order[b.kind]);
    },
    // optionSections groups JPEG sizes after the original, video, and preview choices.
    optionSections(group) {
      const options = this.displayOptions(group);
      return [
        { id: "files", label: "", options: options.filter((option) => option.kind !== "thumbnail") },
        { id: "resized", label: this.$gettext("Resized JPEGs"), options: options.filter((option) => option.kind === "thumbnail") },
      ].filter((section) => section.options.length > 0);
    },
    // needsConversion identifies a video output without an indexed AVC derivative.
    needsConversion(option) {
      return ["avc", "video-part"].includes(option.kind) && !option.ready;
    },
    // optionLabel translates the output choice independently of its metadata.
    optionLabel(option) {
      const labels = { Original: this.$gettext("Original"), "Full-Size Preview": this.$gettext("Full-Size Preview"), "Full-Size Preview (First Page)": this.$gettext("Full-Size Preview (First Page)"), "Full Size": this.$gettext("Full Size"), "Full HD": this.$gettext("Full HD"), "MP4 AVC": this.$gettext("Compatible Video"), "Video Part": this.$gettext("Video Part") };
      return labels[option.label] || option.label;
    },
    // optionMetadata formats indexed dimensions, output format, and codec.
    optionMetadata(option) {
      const width = option.width;
      const height = option.height;
      const codec = option.codec || (["original", "preview"].includes(option.kind) ? option.source.codec : "");
      return [width && height ? `${width} × ${height}` : "", option.format.toUpperCase(), codec].filter(Boolean).join(" · ");
    },
    // optionBytes reports indexed file sizes without estimating generated outputs.
    optionBytes(option) {
      const bytes = ["avc", "video-part"].includes(option.kind) ? option.bytes : option.kind === "thumbnail" ? null : option.source.bytes;
      return bytes !== null && bytes !== undefined ? this.$util.formatBytes(bytes) : "";
    },
    // stateLabel translates the current output state.
    stateLabel(item) {
      const labels = { ready: this.$gettext("Pending"), error: this.$gettext("Failed") };
      return labels[this.states[item.id]?.state] || "";
    },
    // retry restarts one failed output with current authorization and tokens.
    async retry(item) {
      if (this.visible && this.permitted) {
        await this.queue?.retry(item);
      }
    },
    // refreshConfig updates tokens and settings, and fails when the server answers with the public config of a session it no longer accepts.
    async refreshConfig() {
      await this.$config.update();
      if (this.$config.values?.mode === "public") {
        throw new Error("session ended");
      }
    },
    // confirm freezes the selected output list and starts a single cancelable queue.
    async confirm() {
      if (this.started || this.loading || !this.selected.length || !this.permitted || !this.visible) {
        return;
      }
      const generation = this.generation;
      this.downloads = buildExportDownloads(this.groups, this.selected);
      this.started = true;
      this.$nextTick(() => {
        if (generation === this.generation && this.visible) {
          this.$view.focus(this.$refs.content, ".action-cancel");
        }
      });
      this.queue = markRaw(new ExportQueue(this.downloads, {
        refresh: () => this.refreshConfig(),
        allowed: () => generation === this.generation && this.visible && this.permitted,
        config: this.$config,
        progress: (count) => {
          this.handed = count;
          this.complete = count === this.downloads.length;
          if (this.complete) {
            this.close();
          }
        },
        change: (states) => {
          this.states = states;
          this.transferError = Object.values(states).some((item) => item.state === "error");
        },
      }));
      try {
        await this.queue.run();
        if (generation === this.generation && this.visible) {
          this.complete = this.handed === this.downloads.length;
          this.transferError = !this.complete;
        }
      } catch {
        if (generation === this.generation && this.visible) {
          this.transferError = true;
        }
      }
    },
  },
};
</script>
<style scoped>
.export-card {
  min-height: 0;
}

.export-title,
.export-error,
.export-actions {
  flex-shrink: 0;
}

.export-title {
  margin: 0;
  padding: 16px 24px;
}

.export-content {
  min-height: 0;
  overflow-y: auto;
  padding: 16px 24px 20px;
  margin: 0;
}

.export-groups {
  gap: 12px;
}

.export-group {
  overflow: hidden;
}

.export-group-title {
  padding: 16px;
  align-items: center;
}

.export-group-options :deep(.v-expansion-panel-text__wrapper) {
  padding: 0 16px 16px;
}

.export-resized-jpegs {
  margin-top: 12px;
}

.export-section-title {
  margin: 12px 0 4px;
}

.export-thumbnail {
  flex: 0 0 44px;
  height: 44px;
  margin-inline-end: 12px;
}

.export-file {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.export-filename,
.export-file-type,
.export-file-metadata,
.export-option-description > span {
  display: block;
  overflow-wrap: anywhere;
}

.export-option :deep(.v-selection-control__wrapper) {
  flex: 0 0 44px;
  width: 44px;
}

.export-option :deep(.v-label) {
  width: 100%;
  min-width: 0;
  margin-inline-start: 12px;
}

.export-option-label {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding-block: 8px;
}

.export-option-description {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.export-option-size {
  flex-shrink: 0;
}

.export-output {
  padding-block: 12px;
}

.export-output-heading {
  display: flex;
  align-items: center;
  gap: 12px;
}

.export-hint {
  margin: 16px 0 0;
}

.export-error {
  padding: 16px 24px 0;
}

.export-actions {
  margin: 0;
  padding: 16px 24px;
}

.export-actions .action-confirm {
  min-width: 160px;
}

@media (max-width: 959px) {
  .export-title,
  .export-content,
  .export-actions {
    padding: 16px;
  }

  .export-group-title {
    padding: 12px;
  }

  .export-group-options :deep(.v-expansion-panel-text__wrapper) {
    padding: 0 12px 12px;
  }

  .export-error {
    padding: 16px 16px 0;
  }
}
</style>
