<template>
  <v-dialog :model-value="show" persistent max-width="390" class="p-photo-album-dialog" @keydown.esc="cancel">
    <v-form ref="form" validate-on="blur" accept-charset="UTF-8" @submit.prevent="confirm">
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon icon="mdi-image-album" size="28" color="primary"></v-icon>
          <h6 class="text-h6"><translate key="Add to album">Add to album</translate></h6>
        </v-card-title>
        <v-card-text>
          <v-combobox
            ref="input"
            v-model="album"
            autocomplete="off"
            :placeholder="$gettext('Select albums or create a new one')"
            :items="items"
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
        <v-card-actions>
          <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
            <translate>Cancel</translate>
          </v-btn>
          <v-btn :disabled="!album" variant="flat" color="highlight" class="action-confirm text-white" @click.stop="confirm">
            <translate>Confirm</translate>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import Album from "model/album";

// Todo: Handle cases where users have more than 10000 albums.
const MaxResults = 10000;

export default {
  name: "PPhotoAlbumDialog",
  props: {
    show: Boolean,
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
    show: function (show) {
      if (show) {
        this.queryServer("");
      }
    },
  },
  methods: {
    cancel() {
      this.$emit("cancel");
    },
    confirm() {
      if (this.loading) {
        return;
      }

      if (typeof this.album === "object" && this.album?.UID) {
        this.$emit("confirm", this.album?.UID);
      } else if (typeof this.album === "string" && this.album.length > 0) {
        this.loading = true;

        let newAlbum = new Album({ Title: this.album, UID: "", Favorite: false });

        newAlbum
          .save()
          .then((a) => {
            this.loading = false;
            this.album = a;
            this.$emit("confirm", a.UID);
          })
          .catch(() => {
            this.loading = false;
          });
      }
    },
    queryServer(q) {
      if (this.loading) {
        return;
      }

      this.loading = true;

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
          this.$nextTick(() => this.$refs.input.focus());
        })
        .catch(() => {
          this.$nextTick(() => this.$refs.input.focus());
        })
        .finally(() => {
          this.loading = false;
        });
    },
  },
};
</script>
