<template>
  <v-form ref="form" lazy-validation
          dense autocomplete="off" class="p-photo-toolbar p-album-toolbar" accept-charset="UTF-8"
          @submit.prevent="updateQuery">
    <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown" class="page-toolbar" color="secondary">
      <v-toolbar-title :title="album.Title">
        {{ album.Title }}
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon class="hidden-xs-only action-reload" :title="$gettext('Reload')" @click.stop="refresh">
        <v-icon>refresh</v-icon>
      </v-btn>

      <v-btn icon class="action-edit" :title="$gettext('Edit')" @click.stop="dialog.edit = true">
        <v-icon>edit</v-icon>
      </v-btn>

      <v-btn v-if="$config.feature('share')" icon class="action-share" :title="$gettext('Share')"
             @click.stop="dialog.share = true">
        <v-icon>share</v-icon>
      </v-btn>

      <v-btn v-if="$config.feature('download')" icon class="hidden-xs-only action-download" :title="$gettext('Download')"
             @click.stop="download">
        <v-icon>get_app</v-icon>
      </v-btn>

      <v-btn v-if="settings.view === 'cards'" icon :title="$gettext('Toggle View')" @click.stop="setView('list')">
        <v-icon>view_list</v-icon>
      </v-btn>
      <v-btn v-else-if="settings.view === 'list'" icon :title="$gettext('Toggle View')" @click.stop="setView('mosaic')">
        <v-icon>view_comfy</v-icon>
      </v-btn>
      <v-btn v-else icon :title="$gettext('Toggle View')" @click.stop="setView('cards')">
        <v-icon>view_column</v-icon>
      </v-btn>

      <v-btn v-if="!$config.values.readonly && $config.feature('upload')" icon class="hidden-sm-and-down action-upload"
             :title="$gettext('Upload')" @click.stop="showUpload()">
        <v-icon>cloud_upload</v-icon>
      </v-btn>
    </v-toolbar>

    <template v-if="album.Description">
      <v-card flat class="px-2 py-1 hidden-sm-and-down"
              color="secondary-light"
      >
        <v-card-text>
          {{ album.Description }}
        </v-card-text>
      </v-card>
      <v-card flat class="pa-0 hidden-md-and-up"
              color="secondary-light"
      >
        <v-card-text>
          {{ album.Description }}
        </v-card-text>
      </v-card>
    </template>

    <p-share-dialog :show="dialog.share" :model="album" @upload="webdavUpload"
                    @close="dialog.share = false"></p-share-dialog>
    <p-share-upload-dialog :show="dialog.upload" :items="{albums: album.getId()}" :model="album" @cancel="dialog.upload = false"
                           @confirm="dialog.upload = false"></p-share-upload-dialog>
    <p-album-edit-dialog :show="dialog.edit" :album="album" @close="dialog.edit = false"></p-album-edit-dialog>
  </v-form>
</template>
<script>
import Event from "pubsub-js";
import Notify from "common/notify";
import download from "common/download";

export default {
  name: 'PAlbumToolbar',
  props: {
    album: {
      type: Object,
      default: () => {},
    },
    filter: {
      type: Object,
      default: () => {},
    },
    settings: {
      type: Object,
      default: () => {},
    },
    refresh: Function,
    filterChange: Function,
  },
  data() {
    const cameras = [{
      ID: 0,
      Name: this.$gettext('All Cameras')
    }].concat(this.$config.get('cameras'));
    const countries = [{
      ID: '',
      Name: this.$gettext('All Countries')
    }].concat(this.$config.get('countries'));

    return {
      experimental: this.$config.get("experimental"),
      isFullScreen: !!document.fullscreenElement,
      categories: this.$config.albumCategories(),
      searchExpanded: false,
      options: {
        'views': [
          {value: 'mosaic', text: this.$gettext('Mosaic')},
          {value: 'cards', text: this.$gettext('Cards')},
          {value: 'list', text: this.$gettext('List')},
        ],
        'countries': countries,
        'cameras': cameras,
        'sorting': [
          {value: 'added', text: this.$gettext('Recently added')},
          {value: 'edited', text: this.$gettext('Recently edited')},
          {value: 'newest', text: this.$gettext('Newest first')},
          {value: 'oldest', text: this.$gettext('Oldest first')},
          {value: 'name', text: this.$gettext('Sort by file name')},
          {value: 'similar', text: this.$gettext('Group by similarity')},
          {value: 'relevance', text: this.$gettext('Most relevant')},
        ],
      },
      dialog: {
        share: false,
        upload: false,
        edit: false,
      },
      titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
      growDesc: false,
    };
  },
  methods: {
    webdavUpload() {
      this.dialog.share = false;
      this.dialog.upload = true;
    },
    showUpload() {
      Event.publish("dialog.upload");
    },
    expand() {
      this.searchExpanded = !this.searchExpanded;
      this.growDesc = !this.growDesc;
    },
    updateAlbum() {
      if (this.album.wasChanged()) {
        this.album.update();
      }
    },
    dropdownChange() {
      this.updateQuery();

      if (window.innerWidth < 600) {
        this.searchExpanded = false;
      }

      if (this.filter.order !== this.album.Order) {
        this.album.Order = this.filter.order;
        this.updateAlbum();
      }
    },
    setView(name) {
      this.settings.view = name;
      this.updateQuery();
    },
    updateQuery() {
      this.filterChange();
    },
    download() {
      this.onDownload(`${this.$config.apiUri}/albums/${this.album.UID}/dl?t=${this.$config.downloadToken()}`);
    },
    onDownload(path) {
      Notify.success(this.$gettext("Downloading…"));

      download(path, "album.zip");
    },
  }
};
</script>
