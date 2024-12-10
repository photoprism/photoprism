<template>
  <div>
    <div v-if="selection.length > 0" class="position-fixed bottom-0 right-0">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-subject-clipboard`"
        :end="!rtl"
        :start="rtl"
        :attach="true"
        location="top"
        transition="slide-y-reverse-transition"
        offset="8"
        z-index="10"
      >
        <template #activator="{ props }">
          <v-btn v-bind="props" icon size="52" color="secondary" density="comfortable" class="action-menu ma-5">
            <span class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn key="download" :title="$gettext('Download')" icon="mdi-download" color="download" density="comfortable"  class="action-download" :disabled="!canDownload || selection.length !== 1" @click.stop="download()"></v-btn>
        <v-btn v-if="canAddAlbums" key="bookmark" :title="$gettext('Add to album')" icon="mdi-bookmark" color="album" density="comfortable"  :disabled="selection.length === 0" class="action-album" @click.stop="dialog.album = true"></v-btn>
        <v-btn key="close" icon="mdi-close" color="grey-darken-2" density="comfortable"  class="action-clear" @click.stop="clearClipboard()"></v-btn>
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
