<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial id="t-clipboard" v-model="expanded" fixed bottom direction="top" transition="slide-y-reverse-transition" :right="!rtl" :left="rtl" :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-subject-clipboard`">
        <template #activator>
          <v-btn fab dark color="accent darken-2" class="action-menu">
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn fab dark small :title="$gettext('Download')" color="download" class="action-download" :disabled="!canDownload || selection.length !== 1" @click.stop="download()">
          <v-icon>get_app</v-icon>
        </v-btn>

        <v-btn v-if="canAddAlbums" fab dark small :title="$gettext('Add to album')" color="album" :disabled="selection.length === 0" class="action-album" @click.stop="dialog.album = true">
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
  name: "PSubjectClipboard",
  props: {
    selection: {
      type: Array,
      default: () => [],
    },
    refresh: {
      type: Function,
      default: () => {},
    },
    clearSelection: {
      type: Function,
      default: () => {},
    },
  },
  data() {
    return {
      canManage: this.$config.allow("people", "manage"),
      canDownload: this.$config.allow("people", "download") && this.$config.feature("download"),
      canAddAlbums: this.$config.allow("albums", "create") && this.$config.feature("albums"),
      features: this.$config.settings().features,
      expanded: false,
      dialog: {
        delete: false,
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

      Api.post(`albums/${ppid}/photos`, { subjects: this.selection }).then(() => this.onAdded());
    },
    onAdded() {
      this.clearClipboard();
    },
    download() {
      if (this.selection.length !== 1) {
        Notify.error(this.$gettext("You can only download one album"));
        return;
      }

      Notify.success(this.$gettext("Downloading…"));

      Api.post("zip", { subjects: this.selection }).then((r) => {
        this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken}`);
      });

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photos.zip");
    },
  },
};
</script>
