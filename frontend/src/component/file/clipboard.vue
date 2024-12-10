<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-file-clipboard`"
        :end="!rtl"
        :start="rtl"
        :attach="true"
        location="top"
        transition="slide-y-reverse-transition"
        offset="8"
      >
        <template #activator="{ props }">
          <v-btn v-bind="props" icon size="52" color="secondary" density="comfortable" class="action-menu ma-5">
            <span class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn v-if="$config.feature('download')" key="download" :title="$gettext('Download')" icon="mdi-download" color="download" density="comfortable" :disabled="selection.length === 0" class="action-download" @click.stop="download()"></v-btn>
        <v-btn v-if="$config.feature('albums')" key="bookmark" :title="$gettext('Add to album')" icon="mdi-bookmark" color="album" density="comfortable" :disabled="selection.length === 0" class="action-album" @click.stop="dialog.album = true"></v-btn>
        <v-btn key="close" icon="mdi-close" color="grey-darken-2" density="comfortable" class="action-clear" @click.stop="clearClipboard()"></v-btn>
      </v-speed-dial>
    </div>
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
