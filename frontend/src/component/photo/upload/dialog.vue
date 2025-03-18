<template>
  <v-dialog
    :model-value="visible"
    :fullscreen="$vuetify.display.mdAndDown"
    scrim
    scrollable
    persistent
    class="p-photo-upload-dialog v-dialog--upload"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" class="p-photo-upload" validate-on="invalid-input" tabindex="1" @submit.prevent="submit">
      <input ref="upload" type="file" multiple class="d-none input-upload" @change.stop="onUpload()" />
      <v-card :tile="$vuetify.display.mdAndDown">
        <v-toolbar
          v-if="$vuetify.display.mdAndDown"
          flat
          color="navigation"
          class="mb-4"
          :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
        >
          <v-btn icon @click.stop="close">
            <v-icon>mdi-close</v-icon>
          </v-btn>
          <v-toolbar-title>
            {{ title }}
          </v-toolbar-title>
        </v-toolbar>
        <v-card-title v-else class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-cloud-upload</v-icon>
          <h6 class="text-h6">{{ title }}</h6>
        </v-card-title>
        <v-card-text class="flex-grow-0">
          <div class="form-container">
            <div class="form-header">
              <span v-if="failed">{{ $gettext(`Upload failed`) }}</span>
              <span v-else-if="total > 0 && completedTotal < 100">
                {{ $gettext(`Uploading %{n} of %{t}…`, { n: current, t: total }) }}
              </span>
              <span v-else-if="indexing">{{ $gettext(`Upload complete. Indexing…`) }}</span>
              <span v-else-if="completedTotal === 100">{{ $gettext(`Done.`) }}</span>
              <span v-else-if="filesQuotaReached"
                >{{ $gettext(`Insufficient storage.`) }}
                {{ $gettext(`Increase storage size or delete files to continue.`) }}</span
              >
              <span v-else>{{ $gettext(`Select the files to upload…`) }}</span>
            </div>
            <div class="form-body">
              <div class="form-controls">
                <v-combobox
                  v-model="selectedAlbums"
                  :disabled="busy || loading || total > 0 || filesQuotaReached"
                  :loading="loading"
                  hide-details
                  chips
                  closable-chips
                  multiple
                  class="input-albums"
                  :items="albums"
                  item-title="Title"
                  item-value="UID"
                  :placeholder="$gettext('Select or create an album')"
                  return-object
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
                      {{ chip.item.title ? chip.item.title : chip.item }}
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
                  {{ $gettext(`Please do not upload any private, unlawful or offensive pictures. `) }}
                </p>
                <p v-else-if="rejectNSFW">
                  {{ $gettext(`Please don't upload photos containing offensive content.`) }}
                  {{ $gettext(`Uploads that may contain such images will be rejected automatically.`) }}
                </p>
                <p v-if="featReview">
                  {{
                    $gettext(
                      `Non-photographic and low-quality images require a review before they appear in search results.`
                    )
                  }}
                </p>
              </div>
            </div>
          </div>
        </v-card-text>
        <v-card-actions class="action-buttons mt-1">
          <v-btn :disabled="busy" variant="flat" color="button" class="action-close" @click.stop="close">
            {{ $gettext(`Close`) }}
          </v-btn>
          <v-btn
            :disabled="busy || filesQuotaReached"
            variant="flat"
            color="highlight"
            class="action-select action-upload"
            @click.stop="onUploadDialog()"
          >
            {{ $gettext(`Browse`) }}
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
import { Duration } from "luxon";

export default {
  name: "PPhotoUploadDialog",
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
  data() {
    const isDemo = this.$config.get("demo");
    return {
      albums: [],
      selectedAlbums: [],
      selected: [],
      uploads: [],
      busy: false,
      loading: false,
      indexing: false,
      failed: false,
      filesQuotaReached: this.$config.filesQuotaReached(),
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
    close() {
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
    submit() {
      // DO NOTHING
    },
    reset() {
      this.busy = false;
      this.selected = [];
      this.uploads = [];
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
    },
    onUploadDialog() {
      this.$refs.upload.click();
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

      const files = this.$refs.upload.files;

      // Too many files selected for upload?
      if (this.isDemo && files && files.length > this.fileLimit) {
        $notify.error(this.$gettext("Too many files selected"));
        return;
      }

      this.selected = files;
      this.total = files.length;

      // No files selected?
      if (!this.selected || this.total < 1) {
        return;
      }

      this.uploads = [];
      this.token = this.$util.generateToken();
      this.selected = this.$refs.upload.files;
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
          }
        });
      }

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

        this.indexing = true;
        this.eta = "";

        const ctx = this;
        $api
          .put(`users/${userUid}/upload/${ctx.token}`, {
            albums: addToAlbums,
          })
          .then(() => {
            ctx.reset();
            $notify.success(ctx.$gettext("Upload complete"));
            ctx.$emit("confirm");
          })
          .catch(() => {
            ctx.reset();
            $notify.error(ctx.$gettext("Upload failed"));
          });
      });
    },
  },
};
</script>
