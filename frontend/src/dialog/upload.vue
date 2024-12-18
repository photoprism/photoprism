<template>
  <v-dialog :model-value="show" fullscreen :scrim="false" scrollable persistent class="p-upload-dialog" @keydown.esc="cancel">
    <v-card color="background">
      <v-toolbar flat color="navigation" :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
        <v-btn icon @click.stop="cancel">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>
          <translate key="Upload">Upload</translate>
        </v-toolbar-title>
      </v-toolbar>
      <v-container grid-list-xs ext-xs-left fluid>
        <v-form ref="form" class="p-photo-upload" validate-on="blur" @submit.prevent="submit">
          <input ref="upload" type="file" multiple class="d-none input-upload" @change.stop="onUpload()" />

          <v-container fluid>
            <p class="text-body-2 pb-2">
              <!-- TODO: check property allow-overflow TEST -->
              <v-combobox
                v-if="total === 0"
                v-model="selectedAlbums"
                hide-details
                chips
                closable-chips
                multiple
                class="my-0 input-albums"
                :items="albums"
                item-title="Title"
                item-value="UID"
                :label="$gettext('Select albums or create a new one')"
                return-object
              >
                <template #no-data>
                  <v-list-item>
                    <v-list-item-title>
                      <translate key="Press enter to create a new album.">Press enter to create a new album.</translate>
                    </v-list-item-title>
                  </v-list-item>
                </template>
                <template #chip="data">
                  <v-chip
                      :key="JSON.stringify(data.item)"
                      :model-value="data.selected"
                      :disabled="data.disabled"
                      class="bg-highlight rounded-xl"
                      @click:close="removeSelection(data.index)"
                  >
                    <v-icon class="pr-1">mdi-bookmark</v-icon>
                    <!-- TODO: change this filter -->
                    <!-- {{ data.item.Title ? data.item.Title : data.item | truncate(40) }} -->
                    {{ data.item.title ? data.item.title : data.item }}
                  </v-chip>
                </template>
              </v-combobox>
              <span v-else-if="failed"><translate key="Upload failed">Upload failed</translate></span>
              <span v-else-if="total > 0 && completedTotal < 100">
                <translate :translate-params="{ n: current, t: total }">Uploading %{n} of %{t}…</translate>
              </span>
              <span v-else-if="indexing"><translate key="Upload complete">Upload complete. Indexing…</translate></span>
              <span v-else-if="completedTotal === 100"><translate key="Done">Done.</translate></span>
            </p>

            <v-progress-linear v-model="completedTotal" :indeterminate="indexing" class="py-1" :height="21">
              <p class="px-2 ma-0 text-end opacity-85"
                ><span v-if="eta">{{ eta }}</span></p
              >
            </v-progress-linear>

            <p v-if="isDemo" class="text-body-2 py-2">
              <translate :translate-params="{ n: fileLimit }">You can upload up to %{n} files for test purposes.</translate>
              <translate>Please do not upload any private, unlawful or offensive pictures. </translate>
            </p>
            <p v-else-if="rejectNSFW" class="text-body-2 py-2">
              <translate>Please don't upload photos containing offensive content.</translate>
              <translate>Uploads that may contain such images will be rejected automatically.</translate>
            </p>

            <p v-if="featReview" class="text-body-2 py-2">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </p>

            <v-btn :disabled="busy" color="highlight" class="text-white ml-0 mt-2 action-upload" variant="flat" @click.stop="onUploadDialog()">
              <translate key="Upload">Upload</translate>
              <v-icon end>mdi-download</v-icon>
            </v-btn>
          </v-container>
        </v-form>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import Album from "model/album";
import Util from "common/util";
import { Duration } from "luxon";

export default {
  name: "PUploadDialog",
  props: {
    show: Boolean,
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
      loading: false,
      uploads: [],
      busy: false,
      indexing: false,
      failed: false,
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
      rtl: this.$rtl,
    };
  },
  watch: {
    show: function () {
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
      this.findAlbums("");
    },
  },
  methods: {
    removeSelection(index) {
      this.selectedAlbums.splice(index, 1);
    },
    findAlbums(q) {
      if (this.loading) {
        return;
      }

      this.loading = true;

      const params = {
        q: q,
        count: 2000,
        offset: 0,
        type: "album",
      };

      Album.search(params)
        .then((response) => {
          this.loading = false;
          this.albums = response.models;
        })
        .catch(() => (this.loading = false));
    },
    cancel() {
      if (this.busy) {
        Notify.info(this.$gettext("Uploading photos…"));
        return;
      }

      this.$emit("cancel");
    },
    confirm() {
      if (this.busy) {
        Notify.info(this.$gettext("Uploading photos…"));
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
            this.eta = dur.toHuman();
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
        Notify.error(this.$gettext("Too many files selected"));
        return;
      }

      this.selected = files;
      this.total = files.length;

      // No files selected?
      if (!this.selected || this.total < 1) {
        return;
      }

      this.uploads = [];
      this.token = Util.generateToken();
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

      Notify.info(this.$gettext("Uploading photos…"));

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

          await Api.post(`users/${userUid}/upload/${ctx.token}`, formData, {
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
          Notify.error(this.$gettext("Upload failed"));
          return;
        }

        this.indexing = true;
        this.eta = "";

        const ctx = this;
        Api.put(`users/${userUid}/upload/${ctx.token}`, {
          albums: addToAlbums,
        })
          .then(() => {
            ctx.reset();
            Notify.success(ctx.$gettext("Upload complete"));
            ctx.$emit("confirm");
          })
          .catch(() => {
            ctx.reset();
            Notify.error(ctx.$gettext("Upload failed"));
          });
      });
    },
  },
};
</script>
