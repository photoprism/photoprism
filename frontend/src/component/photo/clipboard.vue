<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial id="t-clipboard" v-model="expanded" fixed bottom direction="top" transition="slide-y-reverse-transition" :right="!rtl" :left="rtl" :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-photo-clipboard`">
        <template #activator>
          <v-btn fab dark color="accent darken-2" class="action-menu">
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn v-if="canShare && context !== 'archive' && context !== 'review'" fab dark small :title="$gettext('Share')" color="share" :disabled="selection.length === 0 || busy" class="action-share" @click.stop="dialog.share = true">
          <v-icon>cloud</v-icon>
        </v-btn>

        <v-btn v-if="canManage && context === 'review'" fab dark small :title="$gettext('Approve')" color="share" :disabled="selection.length === 0 || busy" class="action-approve" @click.stop="batchApprove">
          <v-icon>check</v-icon>
        </v-btn>
        <v-btn v-if="canEdit" fab dark small :title="$gettext('Edit')" color="edit" :disabled="selection.length === 0 || busy" class="action-edit" @click.stop="edit">
          <v-icon>edit</v-icon>
        </v-btn>
        <v-btn v-if="canTogglePrivate" fab dark small :title="$gettext('Change private flag')" color="private" :disabled="selection.length === 0 || busy" class="action-private" @click.stop="batchPrivate">
          <v-icon>lock</v-icon>
        </v-btn>
        <v-btn v-if="canDownload && context !== 'archive'" fab dark small :title="$gettext('Download')" :disabled="busy" color="download" class="action-download" @click.stop="download()">
          <v-icon>get_app</v-icon>
        </v-btn>
        <v-btn v-if="canEditAlbum && context !== 'archive'" fab dark small :title="$gettext('Add to album')" color="album" :disabled="selection.length === 0 || busy" class="action-album" @click.stop="dialog.album = true">
          <v-icon>bookmark</v-icon>
        </v-btn>
        <v-btn v-if="canArchive && !isAlbum && context !== 'archive'" fab dark small color="remove" :title="$gettext('Archive')" :disabled="selection.length === 0 || busy" class="action-archive" @click.stop="archivePhotos">
          <v-icon>archive</v-icon>
        </v-btn>
        <v-btn v-if="canArchive && !album && context === 'archive'" fab dark small color="restore" :title="$gettext('Restore')" :disabled="selection.length === 0 || busy" class="action-restore" @click.stop="batchRestore">
          <v-icon>unarchive</v-icon>
        </v-btn>
        <v-btn v-if="canEditAlbum && isAlbum" fab dark small :title="$gettext('Remove from album')" color="remove" :disabled="selection.length === 0 || busy" class="action-remove" @click.stop="removeFromAlbum">
          <v-icon>eject</v-icon>
        </v-btn>
        <v-btn v-if="canDelete && !album && context === 'archive'" fab dark small :title="$gettext('Delete')" color="remove" :disabled="selection.length === 0 || busy" class="action-delete" @click.stop="deletePhotos">
          <v-icon>delete</v-icon>
        </v-btn>
        <v-btn fab dark small color="accent" class="action-clear" @click.stop="clearClipboard()">
          <v-icon>clear</v-icon>
        </v-btn>
      </v-speed-dial>
    </v-container>
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
      canTogglePrivate: this.$config.allow("photos", "manage") && this.context !== "archive" && features.private,
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
