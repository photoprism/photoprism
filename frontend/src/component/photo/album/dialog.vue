<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="390"
    class="p-dialog p-photo-album-dialog"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" accept-charset="UTF-8" tabindex="1" @submit.prevent="confirm">
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon icon="mdi-bookmark" size="28" color="primary"></v-icon>
          <h6 class="text-h6">{{ $gettext(`Add to album`) }}</h6>
        </v-card-title>
        <v-card-text>
          <v-combobox
            ref="input"
            v-model="album"
            autocomplete="off"
            :placeholder="$gettext('Select or create an album')"
            :items="items"
            :disabled="loading"
            :loading="loading"
            hide-no-data
            hide-details
            return-object
            item-title="Title"
            item-value="UID"
            class="input-album"
            @keyup.enter.native="confirm"
          >
          </v-combobox>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn
            :disabled="!album"
            variant="flat"
            color="highlight"
            class="action-confirm text-white"
            @click.stop="confirm"
          >
            {{ $gettext(`Confirm`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import Album from "model/album";

// TODO: Handle cases where users have more than 10000 albums.
const MaxResults = 10000;

export default {
  name: "PPhotoAlbumDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      loading: false,
      newAlbum: null,
      album: null,
      albums: [],
      items: [],
      labels: {
        addToAlbum: this.$gettext("Add to album"),
        createAlbum: this.$gettext("Create album"),
      },
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.reset();
        this.load("");
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
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.loading) {
        return;
      }

      if (typeof this.album === "object" && this.album?.UID) {
        this.loading = true;
        this.$emit("confirm", this.album?.UID);
      } else if (typeof this.album === "string" && this.album.length > 0) {
        this.loading = true;
        let newAlbum = new Album({ Title: this.album, UID: "", Favorite: false });

        newAlbum
          .save()
          .then((a) => {
            this.album = a;
            this.$emit("confirm", a.UID);
          })
          .catch(() => {
            this.loading = false;
          });
      }
    },
    onLoad() {
      this.loading = true;
      this.$nextTick(() => {
        this.$refs.input.focus();
      });
    },
    onLoaded() {
      this.loading = false;
      this.$nextTick(() => {
        this.$refs.input.focus();
      });
    },
    reset() {
      this.loading = false;
      this.newAlbum = null;
      this.albums = [];
      this.items = [];
    },
    load(q) {
      if (this.loading) {
        return;
      }

      this.onLoad();

      const params = {
        q: q,
        count: MaxResults,
        offset: 0,
        type: "album",
      };

      Album.search(params)
        .then((response) => {
          this.albums = response.models;
          this.items = [...this.albums];
        })
        .finally(() => {
          this.onLoaded();
        });
    },
  },
};
</script>
