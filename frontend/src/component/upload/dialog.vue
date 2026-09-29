<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :fullscreen="$vuetify.display.mdAndDown"
    scrim
    scrollable
    persistent
    class="p-upload-dialog v-dialog--upload"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
    @keydown.esc.exact="onClose"
  >
    <v-form ref="form" class="p-photo-upload" validate-on="invalid-input" tabindex="-1" @submit.prevent="onSubmit">
      <v-card :tile="$vuetify.display.mdAndDown">
        <v-toolbar flat color="navigation" class="mb-4" density="comfortable">
          <v-toolbar-title>
            {{ title }}
          </v-toolbar-title>
          <v-btn icon class="action-close" :aria-label="$gettext('Close')" @click.stop="onClose">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text class="flex-grow-0">
          <div class="form-container">
            <div class="form-header">
              <span v-if="failed">{{ $gettext(`Upload failed`) }}</span>
              <span v-else-if="total > 0 && completedTotal < 100">
                {{ $gettext(`Uploading %{n} of %{t}…`, { n: current, t: total }) }}
              </span>
              <span v-else-if="indexing">{{ $gettext(`Upload complete. Indexing…`) }}</span>
              <span v-else-if="completedTotal === 100">{{ $gettext(`Done.`) }}</span>
              <span v-else-if="insufficientStorage"
                >{{ $gettext(`Insufficient storage.`) }} {{ $gettext(`Increase storage size or delete files to continue.`) }}</span
              >
              <span v-else-if="$vuetify.display.mdAndDown">{{ $gettext(`Select the files to upload…`) }}</span>
              <span v-else>{{ $gettext(`Select or drop files to upload…`) }}</span>
            </div>
            <div class="form-body">
              <div class="form-controls">
                <v-file-upload
                  :model-value="selected"
                  :filter-by-type="accept"
                  :disabled="busy || insufficientStorage"
                  :multiple="true"
                  :title="$vuetify.display.mdAndDown ? $gettext('Browse') : ''"
                  :density="$vuetify.display.mdAndDown ? 'compact' : 'comfortable'"
                  :icon="$vuetify.display.mdAndDown ? 'mdi-cloud-upload' : 'mdi-cloud-upload'"
                  hide-details
                  show-size
                  clearable
                  class="mt-1 mb-0 pb-0 input-file-upload"
                  @update:model-value="onFilesSelected"
                />
                <v-combobox
                  v-model="selectedAlbums"
                  v-model:menu="albumsMenu"
                  :disabled="busy || loading || total > 0 || insufficientStorage"
                  :placeholder="$gettext('Select or create albums')"
                  :items="albums"
                  item-title="Title"
                  item-value="UID"
                  hide-details
                  chips
                  closable-chips
                  return-object
                  multiple
                  class="input-albums"
                  @update:menu="onAlbumsMenuUpdate"
                  @keydown.enter.stop="onAlbumsEnter"
                >
                  <template #no-data>
                    <v-list-item>
                      <v-list-item-title>
                        {{ $gettext(`Press enter to create a new album.`) }}
                      </v-list-item-title>
                    </v-list-item>
                  </template>
                  <template #chip="chip">
                    <v-chip
                      :model-value="chip.selected"
                      :disabled="chip.disabled"
                      prepend-icon="mdi-bookmark"
                      class="text-truncate"
                      @click:close="removeSelection(chip.index)"
                    >
                      {{ chip.internalItem.title }}
                    </v-chip>
                  </template>
                </v-combobox>
                <v-progress-linear
                  :model-value="completedTotal"
                  :indeterminate="indexing"
                  :height="16"
                  color="surface-variant"
                  class="v-progress-linear--upload"
                >
                  <span v-if="eta" class="eta text-caption opacity-80">{{ eta }}</span>
                </v-progress-linear>
              </div>
              <div class="form-text">
                <p v-if="isDemo">
                  {{ $gettext(`You can upload up to %{n} files for test purposes.`, { n: fileLimit }) }}
                  {{ $gettext(`Please do not upload any private, unlawful or offensive pictures.`) }}
                </p>
                <p v-else-if="rejectNSFW">
                  {{ $gettext(`Please don't upload photos containing offensive content.`) }}
                  {{ $gettext(`Uploads that may contain such images will be rejected automatically.`) }}
                </p>
                <p v-if="featReview">
                  {{ $gettext(`Non-photographic and low-quality images require a review before they appear in search results.`) }}
                </p>
              </div>
            </div>
          </div>
        </v-card-text>
        <v-card-actions class="action-buttons mt-1">
          <v-btn :disabled="busy" variant="flat" color="button" class="action-close" @click.stop="onClose">
            {{ $gettext(`Close`) }}
          </v-btn>
          <v-btn v-if="retryable" :disabled="busy" variant="flat" color="highlight" class="action-retry" @click.stop="onRetry()">
            {{ $gettext(`Retry`) }}
          </v-btn>
          <v-btn
            v-else
            :disabled="busy || insufficientStorage || !hasFiles"
            variant="flat"
            color="highlight"
            class="action-select action-upload"
            @click.stop="onUpload()"
          >
            {{ $gettext(`Upload`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import Album from "model/album";
import { createAlbumSelectionWatcher } from "common/albums";
import { $gettext } from "common/gettext";
import { Duration } from "luxon";

export default {
  name: "PUploadDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    data: {
      type: Object,
      default: () => {},
    },
  },
  emits: ["close", "confirm"],
  data() {
    const isDemo = this.$config.get("demo");
    return {
      accept: this.$config.get("uploadAllow"),
      albums: [],
      selectedAlbums: [],
      albumsMenu: false,
      suppressAlbumsMenuOpen: false,
      selected: [],
      uploads: [],
      busy: false,
      loading: false,
      indexing: false,
      failed: false,
      insufficientStorage: this.$config.insufficientStorage(),
      current: 0,
      total: 0,
      totalSize: 0,
      totalFailed: 0,
      completedSize: 0,
      completedTotal: 0,
      started: 0,
      remainingTime: -1,
      eta: "",
      token: "",
      retryable: false,
      processAlbums: [],
      isDemo: isDemo,
      fileLimit: isDemo ? 3 : 0,
      rejectNSFW: !this.$config.get("uploadNSFW"),
      featReview: this.$config.feature("review"),
      rtl: this.$isRtl,
    };
  },
  computed: {
    title() {
      return this.$gettext(`Upload`);
    },
    hasFiles() {
      return Array.isArray(this.selected) && this.selected.length > 0;
    },
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.reset();
        this.isDemo = this.$config.get("demo");
        this.fileLimit = this.isDemo ? 3 : 0;
        this.rejectNSFW = !this.$config.get("uploadNSFW");
        this.featReview = this.$config.feature("review");

        // Set currently selected albums.
        if (this.data && Array.isArray(this.data.albums)) {
          this.selectedAlbums = this.data.albums;
        } else {
          this.selectedAlbums = [];
        }

        // Fetch albums from backend.
        this.load("");
      } else {
        this.reset();
      }
    },
    selectedAlbums: createAlbumSelectionWatcher("albums"),
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    removeSelection(index) {
      this.selectedAlbums.splice(index, 1);
    },
    onLoad() {
      this.loading = true;
    },
    onLoaded() {
      this.loading = false;
    },
    onAlbumsEnter() {
      this.suppressAlbumsMenuOpen = true;
      this.albumsMenu = false;
      window.setTimeout(() => {
        this.suppressAlbumsMenuOpen = false;
      }, 250);
    },
    onAlbumsMenuUpdate(val) {
      if (val && this.suppressAlbumsMenuOpen) {
        this.albumsMenu = false;
        return;
      }
      this.albumsMenu = val;
    },
    load(q) {
      if (this.loading) {
        return;
      }

      this.onLoad();

      const params = {
        q: q,
        count: 2000,
        offset: 0,
        type: "album",
      };

      Album.search(params)
        .then((response) => {
          this.albums = response.models;
        })
        .finally(() => {
          this.onLoaded();
        });
    },
    onClose() {
      if (this.busy) {
        $notify.info(this.$gettext("Uploading photos…"));
        return;
      }

      this.$emit("close");
    },
    confirm() {
      if (this.busy) {
        $notify.info(this.$gettext("Uploading photos…"));
        return;
      }

      this.$emit("confirm");
    },
    onSubmit() {
      // DO NOTHING
    },
    reset() {
      this.busy = false;
      this.selected = [];
      this.uploads = [];
      this.clearProgress();
      this.albumsMenu = false;
      this.suppressAlbumsMenuOpen = false;
    },
    // clearProgress ends the current upload, including a retry it offered, but keeps the selected files and albums.
    clearProgress() {
      this.indexing = false;
      this.failed = false;
      this.current = 0;
      this.total = 0;
      this.totalSize = 0;
      this.totalFailed = 0;
      this.completedSize = 0;
      this.completedTotal = 0;
      this.started = 0;
      this.remainingTime = -1;
      this.eta = "";
      this.token = "";
      this.retryable = false;
      this.processAlbums = [];
    },
    onFilesSelected(newFiles) {
      const newArr = Array.isArray(newFiles) ? newFiles : newFiles ? [newFiles] : [];

      // Changing the selection after a failed upload starts over with a new upload.
      if (this.retryable) {
        this.clearProgress();
      }

      const existing = Array.isArray(this.selected) ? this.selected : [];

      // Clear: empty array from the clearable button or a reset.
      if (newArr.length === 0) {
        this.selected = [];
        return;
      }

      // Remove: every file in the new set is already present by reference →
      // this is a single-item removal emitted by VFileUploadItem's × button.
      if (newArr.every((f) => existing.includes(f))) {
        this.selected = newArr;
        return;
      }

      // Browse / drop: merge with existing selection, skip duplicates
      // identified by name + size + lastModified so re-selecting the same
      // file on a second browse pass does not add a second entry.
      const merged = [...existing];
      for (const f of newArr) {
        if (!merged.some((e) => e.name === f.name && e.size === f.size && e.lastModified === f.lastModified)) {
          merged.push(f);
        }
      }
      this.selected = merged;
    },
    onUploadProgress(ev) {
      if (!ev || !ev.loaded || !ev.total) {
        return;
      }

      const { loaded, total } = ev;

      // Update upload status.
      if (loaded > 0 && total > 0 && loaded < total) {
        const currentSize = loaded + this.completedSize;
        const elapsedTime = Date.now() - this.started;
        this.completedTotal = Math.floor((currentSize * 100) / this.totalSize);

        // Show estimate after 10 seconds.
        if (elapsedTime >= 10000) {
          const rate = currentSize / elapsedTime;
          const ms = this.totalSize / rate - elapsedTime;
          this.remainingTime = Math.ceil(ms * 0.001);
          if (this.remainingTime > 0) {
            const dur = Duration.fromObject({
              minutes: Math.floor(this.remainingTime / 60),
              seconds: this.remainingTime % 60,
            });
            this.eta = dur.toHuman({ unitDisplay: "short" });
          } else {
            this.eta = "";
          }
        }
      }
    },
    onUploadComplete(file) {
      if (!file || !file.size || file.size < 0) {
        return;
      }

      this.completedSize += file.size;
      if (this.totalSize > 0) {
        this.completedTotal = Math.floor((this.completedSize * 100) / this.totalSize);
      }
    },
    onUpload() {
      if (this.busy) {
        return;
      }

      // Too many files selected for upload?
      if (this.isDemo && this.selected && this.selected.length > this.fileLimit) {
        $notify.error(this.$gettext("Too many files selected"));
        return;
      }

      // No files selected?
      if (!this.selected || this.selected.length < 1) {
        return;
      }

      this.uploads = [];
      this.token = this.$util.generateToken();
      this.retryable = false;
      this.busy = true;
      this.indexing = false;
      this.failed = false;
      this.current = 0;
      this.total = this.selected.length;
      this.totalFailed = 0;
      this.totalSize = 0;
      this.completedSize = 0;
      this.completedTotal = 0;
      this.started = Date.now();
      this.eta = "";
      this.remainingTime = -1;

      // Calculate total upload size.
      for (let i = 0; i < this.selected.length; i++) {
        let file = this.selected[i];
        this.totalSize += file.size;
      }

      let userUid = this.$session.getUserUID();

      $notify.info(this.$gettext("Uploading photos…"));

      let addToAlbums = [];

      if (this.selectedAlbums && this.selectedAlbums.length > 0) {
        this.selectedAlbums.forEach((a) => {
          if (typeof a === "string") {
            addToAlbums.push(a);
          } else if (a instanceof Album && a.UID) {
            addToAlbums.push(a.UID);
          } else if (typeof a === "object" && a?.UID) {
            addToAlbums.push(a.UID);
          }
        });
      }

      // Deduplicate album UIDs
      addToAlbums = [...new Set(addToAlbums)];

      async function performUpload(ctx) {
        for (let i = 0; i < ctx.selected.length; i++) {
          let file = ctx.selected[i];
          let formData = new FormData();

          ctx.current = i + 1;

          formData.append("files", file);

          await $api
            .post(`users/${userUid}/upload/${ctx.token}`, formData, {
              headers: {
                "Content-Type": "multipart/form-data",
              },
              onUploadProgress: ctx.onUploadProgress,
            })
            .then(() => {
              ctx.onUploadComplete(file);
            })
            .catch(() => {
              ctx.totalFailed++;
              ctx.onUploadComplete(file);
            });
        }
      }

      performUpload(this).then(() => {
        if (this.totalFailed >= this.total) {
          this.reset();
          $notify.error(this.$gettext("Upload failed"));
          return;
        }

        this.eta = "";
        this.processAlbums = addToAlbums;
        this.processUpload();
      });
    },
    // processUpload asks the server to import the files uploaded with the current token. If the import
    // did not run or complete, the files stay staged, so the token is kept and the user can retry.
    processUpload() {
      const userUid = this.$session.getUserUID();
      const token = this.token;

      this.busy = true;
      this.indexing = true;
      this.failed = false;
      this.retryable = false;

      return $api
        .put(`users/${userUid}/upload/${token}`, {
          albums: this.processAlbums,
        })
        .then(() => {
          if (token !== this.token) {
            return;
          }

          this.reset();
          $notify.success($gettext("Upload complete"));
          this.$emit("confirm");
        })
        .catch((err) => {
          const status = err?.response?.status;

          if (token !== this.token) {
            return;
          } else if ([500, 503, 507].includes(status)) {
            this.busy = false;
            this.indexing = false;
            this.failed = true;
            this.retryable = true;
          } else {
            this.reset();
          }

          $notify.error($gettext("Upload failed"));
        });
    },
    // onRetry sends the processing request again for the files that are still staged.
    onRetry() {
      if (this.busy || !this.retryable || !this.token) {
        return;
      }

      this.processUpload();
    },
  },
};
</script>
