<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-label-clipboard`"
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

        <!-- v-btn key="download" :title="$gettext('Download')" icon="mdi-download" color="download" density="comfortable class="action-download" :disabled="selection.length !== 1" @click.stop="download()"></v-btn -->
        <v-btn key="bookmark" :title="$gettext('Add to album')" icon="mdi-bookmark" color="album" density="comfortable" :disabled="!canAddAlbums || selection.length === 0" class="action-album" @click.stop="dialog.album = true"></v-btn>
        <v-btn key="delete" :title="$gettext('Delete')" icon="mdi-delete" color="remove" density="comfortable" :disabled="!canManage || selection.length === 0" class="action-delete" @click.stop="dialog.delete = true"></v-btn>
        <v-btn key="close" icon="mdi-close" color="grey-darken-2" density="comfortable"  class="action-clear" @click.stop="clearClipboard()"></v-btn>
      </v-speed-dial>
    </div>
    <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false" @confirm="addToAlbum"></p-photo-album-dialog>
    <p-label-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false" @confirm="batchDelete"></p-label-delete-dialog>
  </div>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import download from "common/download";

export default {
  name: "PLabelClipboard",
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
      canManage: this.$config.allow("labels", "manage"),
      canDownload: this.$config.allow("labels", "download"),
      canAddAlbums: this.$config.allow("albums", "create") && this.$config.feature("albums"),
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
      if (!this.canAddAlbums) {
        return;
      }

      this.dialog.album = false;

      Api.post(`albums/${ppid}/photos`, { labels: this.selection }).then(() => this.onAdded());
    },
    onAdded() {
      this.clearClipboard();
    },
    batchDelete() {
      if (!this.canManage) {
        return;
      }

      this.dialog.delete = false;

      Api.post("batch/labels/delete", { labels: this.selection }).then(this.onDeleted.bind(this));
    },
    onDeleted() {
      Notify.success(this.$gettext("Labels deleted"));
      this.clearClipboard();
    },
    download() {
      if (!this.canDownload) {
        return;
      }

      if (this.selection.length !== 1) {
        Notify.error(this.$gettext("You can only download one label"));
        return;
      }

      this.onDownload(`${this.$config.apiUri}/labels/${this.selection[0]}/dl?t=${this.$config.downloadToken}`);

      this.expanded = false;
    },
    onDownload(path) {
      Notify.success(this.$gettext("Downloading…"));

      download(path, "label.zip");
    },
  },
};
</script>
