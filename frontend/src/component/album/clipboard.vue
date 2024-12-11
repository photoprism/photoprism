<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
          id="t-clipboard"
          v-model="expanded"
          :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-album-clipboard`"
          :end="!rtl"
          :start="rtl"
          :attach="true"
          location="top"
          transition="slide-y-reverse-transition"
          offset="12"
      >
        <template #activator="{ props }">
          <v-btn v-bind="props" icon size="52" color="primary-button" variant="elevated" density="comfortable" class="action-menu ma-5">
            <span class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn v-if="canShare" key="share" :title="$gettext('Share')" icon="mdi-share-variant" color="share" density="comfortable" :disabled="selection.length !== 1" class="action-share" @click.stop="shareDialog()"></v-btn>
        <v-btn v-if="canManage" key="pencil" :title="$gettext('Edit')" icon="mdi-pencil" color="edit" density="comfortable" :disabled="selection.length !== 1" class="action-edit" @click.stop="editDialog()"></v-btn>
        <v-btn key="download" :title="$gettext('Download')" icon="mdi-download" color="download" density="comfortable" class="action-download" :disabled="!canDownload || selection.length !== 1" @click.stop="download()"></v-btn>
        <v-btn v-if="canManage" key="bookmark" :title="$gettext('Add to album')" icon="mdi-bookmark" color="album" density="comfortable" :disabled="selection.length === 0" class="action-clone" @click.stop="dialog.album = true"></v-btn>
        <v-btn v-if="canDelete && deletable.includes(context)" key="delete" :title="$gettext('Delete')" icon="mdi-delete" color="remove" density="comfortable" :disabled="selection.length === 0" class="action-delete" @click.stop="dialog.delete = true"></v-btn>
        <v-btn key="close" icon="mdi-close" color="grey-darken-2" density="comfortable" class="action-clear" @click.stop="clearClipboard()"></v-btn>
      </v-speed-dial>
    </div>
    <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false" @confirm="cloneAlbums"></p-photo-album-dialog>
    <p-album-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false" @confirm="batchDelete"></p-album-delete-dialog>
  </div>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import Album from "model/album";
import download from "common/download";

export default {
  name: "PAlbumClipboard",
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
    share: {
      type: Function,
      default: () => {},
    },
    edit: {
      type: Function,
      default: () => {},
    },
    context: {
      type: String,
      default: "",
    },
  },
  data() {
    const features = this.$config.settings().features;

    return {
      canDelete: this.$config.allow("albums", "delete"),
      canDownload: this.$config.allow("albums", "download") && features.download,
      canShare: this.$config.allow("albums", "share") && features.share,
      canManage: this.$config.allow("albums", "manage"),
      deletable: ["album", "moment", "state"],
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
    editDialog() {
      if (this.selection.length !== 1) {
        this.$notify.error(this.$gettext("You may only select one item"));
        return;
      }

      this.model = new Album();
      this.model.find(this.selection[0]).then((m) => {
        this.edit(m);
      });
    },
    shareDialog() {
      if (this.selection.length !== 1) {
        this.$notify.error(this.$gettext("You may only select one item"));
        return;
      }

      this.model = new Album();
      this.model.find(this.selection[0]).then((m) => {
        this.share(m);
      });
    },
    clearClipboard() {
      this.clearSelection();
      this.expanded = false;
    },
    cloneAlbums(ppid) {
      this.dialog.album = false;

      Api.post(`albums/${ppid}/clone`, { albums: this.selection }).then(() => this.onCloned());
    },
    onCloned() {
      this.clearClipboard();
    },
    batchDelete() {
      this.dialog.delete = false;

      Api.post("batch/albums/delete", { albums: this.selection }).then(this.onDeleted.bind(this));
    },
    onDeleted() {
      Notify.success(this.$gettext("Albums deleted"));
      this.clearClipboard();
    },
    download() {
      if (this.selection.length !== 1) {
        Notify.error(this.$gettext("You can only download one album"));
        return;
      }

      Notify.success(this.$gettext("Downloading…"));

      this.onDownload(`${this.$config.apiUri}/albums/${this.selection[0]}/dl?t=${this.$config.downloadToken}`);

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photoprism-album.zip");
    },
  },
};
</script>
