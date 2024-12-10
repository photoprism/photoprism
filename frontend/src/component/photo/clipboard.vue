<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-photo-clipboard`"
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

        <v-btn v-if="canShare && context !== 'archive' && context !== 'hidden' && context !== 'review'" key="cloud" :title="$gettext('Share')" icon="mdi-cloud" color="share" density="comfortable" :disabled="selection.length === 0 || busy" class="action-share" @click.stop="dialog.share = true"></v-btn>
        <!-- TODO: change this icon and key -->
        <v-btn v-if="canManage && context === 'review'" key="check" :title="$gettext('Approve')" icon="mdi-check" color="share" density="comfortable" :disabled="selection.length === 0 || busy" class="action-approve" @click.stop="batchApprove"></v-btn>
        <v-btn v-if="canEdit" key="pencil" :title="$gettext('Edit')" icon="mdi-pencil" color="edit" density="comfortable" :disabled="selection.length === 0 || busy" class="action-edit" @click.stop="edit"></v-btn>
        <v-btn v-if="canTogglePrivate && context !== 'archive' && context !== 'hidden'" key="lock" :title="$gettext('Change private flag')" icon="mdi-lock" color="private" density="comfortable" :disabled="selection.length === 0 || busy" class="action-private" @click.stop="batchPrivate"></v-btn>
        <v-btn v-if="canDownload && context !== 'archive'" key="download" :title="$gettext('Download')" icon="mdi-download" color="download" density="comfortable" :disabled="busy" class="action-download" @click.stop="download()"></v-btn>
        <v-btn v-if="canEditAlbum && context !== 'archive' && context !== 'hidden'" key="bookmark" :title="$gettext('Add to album')" icon="mdi-bookmark" color="album" density="comfortable" :disabled="selection.length === 0 || busy" class="action-album" @click.stop="dialog.album = true"></v-btn>
        <v-btn v-if="canArchive && !isAlbum && context !== 'archive' && context !== 'hidden'" key="package-down" :title="$gettext('Archive')" icon="mdi-package-down" color="remove" density="comfortable" :disabled="selection.length === 0 || busy" class="action-archive" @click.stop="archivePhotos"></v-btn>
        <!-- TODO: change this icon and key -->
        <v-btn v-if="canArchive && !album && context === 'archive' && context !== 'hidden'" key="unarchive" :title="$gettext('Restore')" icon="mdi-unarchive" color="restore" density="comfortable" :disabled="selection.length === 0 || busy" class="action-restore" @click.stop="batchRestore"></v-btn>
        <v-btn v-if="canEditAlbum && isAlbum" key="eject" :title="$gettext('Remove from album')" icon="mdi-eject" color="remove" density="comfortable" :disabled="selection.length === 0 || busy" class="action-remove" @click.stop="removeFromAlbum"></v-btn>
        <v-btn v-if="canDelete && !album && context === 'archive'" key="delete" :title="$gettext('Delete')" icon="mdi-delete" color="remove" density="comfortable" :disabled="selection.length === 0 || busy" class="action-delete" @click.stop="deletePhotos"></v-btn>
        <v-btn key="close" icon="mdi-close" color="grey-darken-2" density="comfortable" class="action-clear" @click.stop="clearClipboard()"></v-btn>
      </v-speed-dial>
    </div>
    <p-photo-archive-dialog :show="dialog.archive" @cancel="dialog.archive = false" @confirm="batchArchive"></p-photo-archive-dialog>
    <p-photo-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false" @confirm="batchDelete"></p-photo-delete-dialog>
    <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false" @confirm="addToAlbum"></p-photo-album-dialog>
    <p-share-upload-dialog :show="dialog.share" :items="{ photos: selection }" :model="album" @cancel="dialog.share = false" @confirm="onShared"></p-share-upload-dialog>
  </div>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import Event from "pubsub-js";
import download from "common/download";
import Photo from "model/photo";

export default {
  name: "PPhotoClipboard",
  props: {
    context: {
      type: String,
      default: "photos",
    },
    selection: {
      type: Array,
      default: () => [],
    },
    refresh: {
      type: Function,
      default: () => {},
    },
    album: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    const features = this.$config.settings().features;

    return {
      canTogglePrivate: this.$config.allow("photos", "manage") && features.private,
      canArchive: this.$config.allow("photos", "delete") && features.archive,
      canDelete: this.$config.allow("photos", "delete") && features.delete,
      canDownload: this.$config.allow("photos", "download") && features.download,
      canShare: this.$config.allow("photos", "share") && features.share,
      canManage: this.$config.allow("photos", "manage") && features.albums,
      canEdit: this.$config.allow("photos", "update") && features.edit,
      canEditAlbum: this.$config.allow("albums", "update") && features.albums,
      busy: false,
      config: this.$config.values,
      expanded: false,
      isAlbum: this.album && this.album.Type === "album",
      dialog: {
        archive: false,
        delete: false,
        album: false,
        share: false,
      },
      rtl: this.$rtl,
    };
  },
  methods: {
    clearClipboard() {
      this.$clipboard.clear();
      this.expanded = false;
    },
    batchApprove() {
      if (this.busy || !this.canManage) {
        return;
      }

      this.busy = true;

      Api.post("batch/photos/approve", { photos: this.selection })
        .then(() => this.onApproved())
        .finally(() => {
          this.busy = false;
        });
    },
    onApproved() {
      Notify.success(this.$gettext("Selection approved"));
      this.clearClipboard();
    },
    archivePhotos() {
      if (!this.canArchive) {
        return;
      }

      if (!this.canDelete) {
        this.dialog.archive = true;
      } else {
        this.batchArchive();
      }
    },
    batchArchive() {
      if (this.busy || !this.canArchive) {
        return;
      }

      this.busy = true;
      this.dialog.archive = false;

      Api.post("batch/photos/archive", { photos: this.selection })
        .then(() => this.onArchived())
        .finally(() => {
          this.busy = false;
        });
    },
    onArchived() {
      Notify.success(this.$gettext("Selection archived"));
      this.clearClipboard();
    },
    deletePhotos() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = true;
    },
    batchDelete() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = false;

      Api.post("batch/photos/delete", { photos: this.selection }).then(() => this.onDeleted());
    },
    onDeleted() {
      Notify.success(this.$gettext("Permanently deleted"));
      this.clearClipboard();
    },
    batchPrivate() {
      Api.post("batch/photos/private", { photos: this.selection }).then(() => this.onPrivateSaved());
    },
    onPrivateSaved() {
      this.clearClipboard();
    },
    batchRestore() {
      Api.post("batch/photos/restore", { photos: this.selection }).then(() => this.onRestored());
    },
    onRestored() {
      Notify.success(this.$gettext("Selection restored"));
      this.clearClipboard();
    },
    addToAlbum(ppid) {
      if (!ppid || !this.canManage) {
        return;
      }

      if (this.busy) {
        return;
      }

      this.busy = true;
      this.dialog.album = false;

      Api.post(`albums/${ppid}/photos`, { photos: this.selection })
        .then(() => this.onAdded())
        .finally(() => {
          this.busy = false;
        });
    },
    onAdded() {
      this.clearClipboard();
    },
    removeFromAlbum() {
      if (!this.album) {
        this.$notify.error(this.$gettext("remove failed: unknown album"));
        return;
      }

      if (this.busy || !this.canManage) {
        return;
      }

      this.busy = true;

      const uid = this.album.UID;

      this.dialog.album = false;

      Api.delete(`albums/${uid}/photos`, { data: { photos: this.selection } })
        .then(() => this.onRemoved())
        .finally(() => {
          this.busy = false;
        });
    },
    onRemoved() {
      this.clearClipboard();
    },
    download() {
      if (this.busy || !this.canDownload) {
        return;
      }

      this.busy = true;

      switch (this.selection.length) {
        case 0:
          this.busy = false;
          return;
        case 1:
          new Photo()
            .find(this.selection[0])
            .then((p) => p.downloadAll())
            .finally(() => {
              this.busy = false;
            });
          break;
        default:
          Api.post("zip", { photos: this.selection })
            .then((r) => {
              this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken}`);
            })
            .finally(() => {
              this.busy = false;
            });
      }

      Notify.success(this.$gettext("Downloading…"));

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photos.zip");
    },
    edit() {
      // Open Edit Dialog
      Event.PubSub.publish("dialog.edit", { selection: this.selection, album: this.album, index: 0 });
    },
    onShared() {
      this.dialog.share = false;
      this.clearClipboard();
    },
  },
};
</script>
