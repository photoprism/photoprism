<template>
  <v-dialog :model-value="show" persistent max-width="356" class="p-photo-album-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-text class="dense mt-2">
        <v-row dense>
          <v-col cols="3" class="text-left">
            <v-icon size="60" color="surface-variant">mdi-image-album</v-icon>
          </v-col>
          <v-col cols="9" class="text-left" align-self="center">
            <v-combobox
              ref="input"
              v-model="album"
              autocomplete="off"
              :hint="$gettext('Album Name')"
              :items="items"
              :loading="loading"
              hide-no-data
              hide-details
              return-object
              item-title="Title"
              item-value="UID"
              :label="$gettext('Album Name')"
              class="input-album"
              @keyup.enter.native="confirm"
            >
            </v-combobox>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <translate>Cancel</translate>
        </v-btn>
        <v-btn variant="flat" color="primary-button" class="action-confirm text-white" @click.stop="confirm">
          <span v-if="typeof album === 'object'">{{ labels.addToAlbum }}</span>
          <span v-else>{{ labels.createAlbum }}</span>
        </v-btn>
      </v-card-actions>
    </v-card>
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
