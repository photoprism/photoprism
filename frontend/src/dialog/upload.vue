<template>
  <v-dialog :value="show" fullscreen hide-overlay scrollable
            lazy persistent class="p-upload-dialog" @keydown.esc="cancel">
    <v-card color="application">
      <v-toolbar dark flat color="navigation" :dense="$vuetify.breakpoint.smAndDown">
        <v-btn icon dark @click.stop="cancel">
          <v-icon>close</v-icon>
        </v-btn>
        <v-toolbar-title>
          <translate key="Upload">Upload</translate>
        </v-toolbar-title>
      </v-toolbar>
      <v-container grid-list-xs ext-xs-left fluid>
        <v-form ref="form" class="p-photo-upload" lazy-validation dense @submit.prevent="submit">
          <input ref="upload" type="file" multiple class="d-none input-upload" @change.stop="onUpload()">

          <v-container fluid>
            <p class="subheading">
              <v-combobox v-if="total === 0" v-model="selectedAlbums" flat solo hide-details chips
                          deletable-chips multiple color="secondary-dark"
                          class="my-0 input-albums"
                          :items="albums"
                          item-value="UID"
                          item-text="Title"
                          :allow-overflow="false"
                          :label="$gettext('Select albums or create a new one')"
                          return-object
              >
                <template #no-data>
                  <v-list-tile>
                    <v-list-tile-content>
                      <v-list-tile-title>
                        <translate
                            key="Press enter to create a new album.">Press enter to create a new album.</translate>
                      </v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                </template>
                <template #selection="data">
                  <v-chip
                      :key="JSON.stringify(data.item)"
                      :selected="data.selected"
                      :disabled="data.disabled"
                      class="v-chip--select-multi"
                      @input="data.parent.selectItem(data.item)"
                  >
                    <v-icon class="pr-1">bookmark</v-icon>
                    {{ data.item.Title ? data.item.Title : data.item | truncate(40) }}
                  </v-chip>
                </template>
              </v-combobox>
              <span v-else-if="failed"><translate key="Upload failed">Upload failed</translate></span>
              <span v-else-if="total > 0 && completed < 100">
                <translate :translate-params="{n: current, t: total}">Uploading %{n} of %{t}…</translate>
              </span>
              <span v-else-if="indexing"><translate key="Upload complete">Upload complete. Indexing…</translate></span>
              <span v-else-if="completed === 100"><translate key="Done">Done.</translate></span>
            </p>

            <v-progress-linear v-model="completed" color="secondary-dark"
                               :indeterminate="indexing"></v-progress-linear>

            <p v-if="safe" class="body-1">
              <translate>Please don't upload photos containing offensive content.</translate>
              <translate>Uploads that may contain such images will be rejected automatically.</translate>
            </p>

            <p v-if="review" class="body-1">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </p>

            <v-btn
                :disabled="busy"
                color="primary-button"
                class="white--text ml-0 mt-2 action-upload"
                depressed
                @click.stop="onUploadDialog()"
            >
              <translate key="Upload">Upload</translate>
              <v-icon :right="!rtl" :left="rtl"  dark>cloud_upload</v-icon>
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

export default {
  name: 'PTabUpload',
  props: {
    show: Boolean,
  },
  data() {
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
      completed: 0,
      token: "",
      review: this.$config.feature("review"),
      safe: !this.$config.get("uploadNSFW"),
      rtl: this.$rtl,
    };
  },
  watch: {
    show: function () {
      this.reset();
      this.review = this.$config.feature("review");
      this.safe = !this.$config.get("uploadNSFW");
      this.findAlbums("");
    }
  },
  methods: {
    findAlbums(q) {
      if (this.loading) {
        return;
      }

      this.loading = true;

      const params = {
        q: q,
        count: 2000,
        offset: 0,
        type: "album"
      };

      Album.search(params).then(response => {
        this.loading = false;
        this.albums = response.models;
      }).catch(() => this.loading = false);
    },
    cancel() {
      if (this.busy) {
        Notify.info(this.$gettext("Uploading photos…"));
        return;
      }

      this.$emit('cancel');
    },
    confirm() {
      if (this.busy) {
        Notify.info(this.$gettext("Uploading photos…"));
        return;
      }

      this.$emit('confirm');
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
      this.completed = 0;
      this.token = "";
    },
    onUploadDialog() {
      this.$refs.upload.click();
    },
    onUpload() {
      if (this.busy) {
        return;
      }

      this.selected = this.$refs.upload.files;
      this.total = this.selected.length;

      if (this.total < 1) {
        return;
      }

      this.token = Util.generateToken();
      this.selected = this.$refs.upload.files;
      this.total = this.selected.length;
      this.busy = true;
      this.indexing = false;
      this.failed = false;
      this.current = 0;
      this.completed = 0;
      this.uploads = [];

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

          formData.append('files', file);

          await Api.post(`users/${userUid}/upload/${ctx.token}`,
            formData,
            {
              headers: {
                'Content-Type': 'multipart/form-data'
              }
            }
          ).then(() => {
            ctx.completed = Math.round((ctx.current / ctx.total) * 100);
          }).catch(() => {
            ctx.completed = Math.round((ctx.current / ctx.total) * 100);
          });
        }
      }

      performUpload(this).then(() => {
        this.indexing = true;
        const ctx = this;
        Api.put(`users/${userUid}/upload/${ctx.token}`,{
          albums: addToAlbums,
        }).then(() => {
          ctx.reset();
          Notify.success(ctx.$gettext("Upload complete"));
          ctx.$emit('confirm');
        }).catch(() => {
          ctx.reset();
          Notify.error(ctx.$gettext("Upload failed"));
        });
      });
    },
  },
};
</script>
