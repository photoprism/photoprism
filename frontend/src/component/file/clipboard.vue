<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial id="t-clipboard" v-model="expanded" fixed bottom direction="top" transition="slide-y-reverse-transition" :right="!rtl" :left="rtl" :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-file-clipboard`">
        <template #activator>
          <v-btn fab dark color="accent darken-2" class="action-menu">
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn v-if="$config.feature('download')" fab dark small :title="$gettext('Download')" color="download" class="action-download" :disabled="selection.length === 0" @click.stop="download()">
          <v-icon>get_app</v-icon>
        </v-btn>

        <v-btn v-if="$config.feature('albums')" fab dark small :title="$gettext('Add to album')" color="album" :disabled="selection.length === 0" class="action-album" @click.stop="dialog.album = true">
          <v-icon>bookmark</v-icon>
        </v-btn>

        <v-btn fab dark small color="accent" class="action-clear" @click.stop="clearClipboard()">
          <v-icon>clear</v-icon>
        </v-btn>
      </v-speed-dial>
    </v-container>
    <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false" @confirm="addToAlbum"></p-photo-album-dialog>
  </div>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import download from "common/download";

export default {
  name: "PFileClipboard",
  props: {
    selection: {
      type: Array,
      default: () => [],
    },
    refresh: Function,
    clearSelection: Function,
  },
  data() {
    return {
      expanded: false,
      dialog: {
        album: false,
        edit: false,
      },
      rtl: this.$rtl,
    };
  },
  methods: {
    clearClipboard() {
      this.clearSelection();
      this.expanded = false;
    },
    addToAlbum(ppid) {
      this.dialog.album = false;

      Api.post(`albums/${ppid}/photos`, { files: this.selection }).then(() => this.onAdded());
    },
    onAdded() {
      this.clearClipboard();
    },
    download() {
      Api.post("zip", { files: this.selection }).then((r) => {
        this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken}`);
      });

      this.expanded = false;
    },
    onDownload(path) {
      Notify.success(this.$gettext("Downloading…"));

      download(path, "photos.zip");
    },
  },
};
</script>
