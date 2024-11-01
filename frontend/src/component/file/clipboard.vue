<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial id="t-clipboard" v-model="expanded" position="fixed" location="bottom" direction="top" transition="slide-y-reverse-transition" :end="!rtl" :start="rtl" :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-file-clipboard`">
        <template #activator>
          <v-btn theme="dark" color="accent-darken-2 rounded-circle" class="action-menu">
            <!-- TODO: change this icon -->
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn v-if="$config.feature('download')" theme="dark" size="small" :title="$gettext('Download')" color="download" class="action-download rounded-circle" :disabled="selection.length === 0" @click.stop="download()">
          <v-icon>mdi-download</v-icon>
        </v-btn>

        <v-btn v-if="$config.feature('albums')" theme="dark" size="small" :title="$gettext('Add to album')" color="album" :disabled="selection.length === 0" class="action-album rounded-circle" @click.stop="dialog.album = true">
          <v-icon>mdi-bookmark</v-icon>
        </v-btn>

        <v-btn theme="dark" size="small" color="accent" class="action-clear rounded-circle" @click.stop="clearClipboard()">
          <v-icon>mdi-close</v-icon>
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
