<template>
  <v-dialog :value="show" lazy persistent max-width="356" class="p-photo-album-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-text class="pt-3 px-3">
        <v-layout row wrap>
          <v-flex xs3 text-xs-left>
            <v-icon size="60" color="secondary-dark lighten-1">photo_album</v-icon>
          </v-flex>
          <v-flex xs9 text-xs-left align-self-center>
            <v-autocomplete
              ref="input"
              v-model="album"
              browser-autocomplete="off"
              :hint="$gettext('Album Name')"
              :items="items"
              :search-input.sync="search"
              :loading="loading"
              hide-no-data
              hide-details
              box
              flat
              item-text="Title"
              item-value="UID"
              :label="$gettext('Album Name')"
              color="secondary-dark"
              class="input-album"
              @keyup.enter.native="confirm"
            >
            </v-autocomplete>
          </v-flex>
        </v-layout>
      </v-card-text>
      <v-card-actions class="pt-0 pb-3 px-3">
        <v-layout row wrap class="pa-0">
          <v-flex xs12 text-xs-right>
            <v-btn depressed color="secondary-light" class="action-cancel mx-1" @click.stop="cancel">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn depressed color="primary-button" class="action-confirm white--text compact mx-0" @click.stop="confirm">
              <span v-if="!album">{{ labels.createAlbum }}</span>
              <span v-else>{{ labels.addToAlbum }}</span>
            </v-btn>
          </v-flex>
        </v-layout>
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
      search: "",
      newAlbum: null,
      album: false,
      albums: [],
      items: [],
      labels: {
        addToAlbum: this.$gettext("Add to album"),
        createAlbum: this.$gettext("Create album"),
      },
    };
  },
  watch: {
    search(q) {
      const exists = this.albums.findIndex((album) => album.Title === q);

      if (exists !== -1 || !q) {
        this.items = this.albums;
        this.newAlbum = null;
      } else {
        this.newAlbum = new Album({ Title: q, UID: "", Favorite: false });
        this.items = this.albums.concat([this.newAlbum]);
      }
    },
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

      if (this.album) {
        this.$emit("confirm", this.album);
      } else if (this.newAlbum) {
        this.loading = true;

        this.newAlbum
          .save()
          .then((a) => {
            this.loading = false;
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
