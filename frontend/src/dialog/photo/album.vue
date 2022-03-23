<template>
  <v-dialog :value="show" lazy persistent max-width="350" class="p-photo-album-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <v-flex xs3 text-xs-center>
            <v-icon size="56" color="secondary-dark lighten-1">photo_album</v-icon>
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
                hide-details
                hide-no-data
                item-text="Title"
                item-value="UID"
                :label="$gettext('Album Name')"
                color="secondary-dark"
                flat solo
                class="input-album"
                @keyup.enter.native="confirm"
            >
            </v-autocomplete>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-3">
            <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn color="primary-button" depressed dark class="action-confirm"
                   @click.stop="confirm">
              <span v-if="!album">{{ labels.createAlbum }}</span>
              <span v-else>{{ labels.addToAlbum }}</span>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
import Album from "model/album";

export default {
  name: 'PPhotoAlbumDialog',
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
      }
    };
  },
  watch: {
    search(q) {
      const exists = this.albums.findIndex((album) => album.Title === q);

      if (exists !== -1 || !q) {
        this.items = this.albums;
        this.newAlbum = null;
      } else {
        this.newAlbum = new Album({Title: q, UID: "", Favorite: false});
        this.items = this.albums.concat([this.newAlbum]);
      }
    },
    show: function (show) {
      if (show) {
        this.queryServer("");
      }
    }
  },
  methods: {
    cancel() {
      this.$emit('cancel');
    },
    confirm() {
      if (this.album) {
        this.$emit('confirm', this.album);
      } else if (this.newAlbum) {
        this.loading = true;

        this.newAlbum.save().then((a) => {
          this.loading = false;
          this.$emit('confirm', a.UID);
        });
      }
    },
    queryServer(q) {
      if (this.loading) {
        return;
      }

      this.loading = true;

      // todo: either introduce infinite flag or
      // make count parameter optional for REST API
      const MAX_COUNT = 10000;

      const params = {
        q: q,
        count: MAX_COUNT,
        offset: 0,
        type: "album"
      };

      Album.search(params).then(response => {
        this.loading = false;
        this.albums = response.models;
        this.items = [...this.albums];
        this.$nextTick(() => this.$refs.input.focus());
      }).catch(() => {
        this.loading = false;
        this.$nextTick(() => this.$refs.input.focus());
      });
    },
  },
};
</script>
