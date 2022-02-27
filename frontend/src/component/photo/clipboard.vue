<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial
          id="t-clipboard" v-model="expanded" fixed
          bottom
          direction="top"
          transition="slide-y-reverse-transition"
          :right="!rtl"
          :left="rtl"
          :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-photo-clipboard`"
      >
        <template #activator>
          <v-btn
              fab
              dark
              color="accent darken-2"
              class="action-menu"
          >
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn
            v-if="context !== 'archive' && context !== 'review' && features.share" fab dark
            small
            :title="$gettext('Share')"
            color="share"
            :disabled="selection.length === 0"
            class="action-share"
            @click.stop="dialog.share = true"
        >
          <v-icon>cloud</v-icon>
        </v-btn>

        <v-btn
            v-if="context === 'review'" fab dark
            small
            :title="$gettext('Approve')"
            color="share"
            :disabled="selection.length === 0"
            class="action-approve"
            @click.stop="batchApprove"
        >
          <v-icon>check</v-icon>
        </v-btn>
        <v-btn
            v-if="context !== 'archive' && features.edit" fab dark
            small
            :title="$gettext('Edit')"
            color="edit"
            :disabled="selection.length === 0"
            class="action-edit"
            @click.stop="edit"
        >
          <v-icon>edit</v-icon>
        </v-btn>
        <v-btn
            v-if="context !== 'archive' && features.private" fab dark
            small
            :title="$gettext('Change private flag')"
            color="private"
            :disabled="selection.length === 0"
            class="action-private"
            @click.stop="batchPrivate"
        >
          <v-icon>lock</v-icon>
        </v-btn>
        <v-btn
            v-if="context !== 'archive' && features.download" fab dark
            small
            :title="$gettext('Download')"
            color="download"
            class="action-download"
            @click.stop="download()"
        >
          <v-icon>get_app</v-icon>
        </v-btn>
        <v-btn
            v-if="context !== 'archive' && features.albums" fab dark
            small
            :title="$gettext('Add to album')"
            color="album"
            :disabled="selection.length === 0"
            class="action-album"
            @click.stop="dialog.album = true"
        >
          <v-icon>bookmark</v-icon>
        </v-btn>
        <v-btn
            v-if="!isAlbum && context !== 'archive' && features.archive" fab dark
            small
            color="remove"
            :title="$gettext('Archive')"
            :disabled="selection.length === 0"
            class="action-archive"
            @click.stop="archivePhotos"
        >
          <v-icon>archive</v-icon>
        </v-btn>
        <v-btn
            v-if="!album && context === 'archive'" fab dark
            small
            color="restore"
            :title="$gettext('Restore')"
            :disabled="selection.length === 0"
            class="action-restore"
            @click.stop="batchRestore"
        >
          <v-icon>unarchive</v-icon>
        </v-btn>
        <v-btn
            v-if="isAlbum && features.albums" fab dark
            small
            :title="$gettext('Remove from album')"
            color="remove"
            :disabled="selection.length === 0"
            class="action-remove"
            @click.stop="removeFromAlbum"
        >
          <v-icon>eject</v-icon>
        </v-btn>
        <v-btn
            v-if="!album && context === 'archive' && features.delete" fab dark
            small
            :title="$gettext('Delete')"
            color="remove"
            :disabled="selection.length === 0"
            class="action-delete"
            @click.stop="deletePhotos"
        >
          <v-icon>delete</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            color="accent"
            class="action-clear"
            @click.stop="clearClipboard()"
        >
          <v-icon>clear</v-icon>
        </v-btn>
      </v-speed-dial>
    </v-container>
    <p-photo-archive-dialog :show="dialog.archive" @cancel="dialog.archive = false"
                            @confirm="batchArchive"></p-photo-archive-dialog>
    <p-photo-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false"
                            @confirm="batchDelete"></p-photo-delete-dialog>
    <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false"
                          @confirm="addToAlbum"></p-photo-album-dialog>
    <p-share-upload-dialog :show="dialog.share" :items="{photos: selection}" :model="album" @cancel="dialog.share = false"
                           @confirm="onShared"></p-share-upload-dialog>
  </div>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import Event from "pubsub-js";
import download from "common/download";
import Photo from "model/photo";

export default {
  name: 'PPhotoClipboard',
  props: {
    selection: {
      type: Array,
      default: () => [],
    },
    refresh: Function,
    album: {
      type: Object,
      default: () => {},
    },
    context: String,
  },
  data() {
    return {
      config: this.$config.values,
      features: this.$config.settings().features,
      expanded: false,
      isAlbum: this.album && this.album.Type === 'album',
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
      Api.post("batch/photos/approve", {"photos": this.selection}).then(() => this.onApproved());
    },
    onApproved() {
      Notify.success(this.$gettext("Selection approved"));
      this.clearClipboard();
    },
    archivePhotos() {
      if (!this.features.delete) {
        this.dialog.archive = true;
      } else {
        this.batchArchive();
      }
    },
    batchArchive() {
      this.dialog.archive = false;

      Api.post("batch/photos/archive", {"photos": this.selection}).then(() => this.onArchived());
    },
    onArchived() {
      Notify.success(this.$gettext("Selection archived"));
      this.clearClipboard();
    },
    deletePhotos() {
      this.dialog.delete = true;
    },
    batchDelete() {
      this.dialog.delete = false;

      Api.post("batch/photos/delete", {"photos": this.selection}).then(() => this.onDeleted());
    },
    onDeleted() {
      Notify.success(this.$gettext("Permanently deleted"));
      this.clearClipboard();
    },
    batchPrivate() {
      Api.post("batch/photos/private", {"photos": this.selection}).then(() => this.onPrivateSaved());
    },
    onPrivateSaved() {
      this.clearClipboard();
    },
    batchRestore() {
      Api.post("batch/photos/restore", {"photos": this.selection}).then(() => this.onRestored());
    },
    onRestored() {
      Notify.success(this.$gettext("Selection restored"));
      this.clearClipboard();
    },
    addToAlbum(ppid) {
      this.dialog.album = false;

      Api.post(`albums/${ppid}/photos`, {"photos": this.selection}).then(() => this.onAdded());
    },
    onAdded() {
      this.clearClipboard();
    },
    removeFromAlbum() {
      if (!this.album) {
        this.$notify.error(this.$gettext("remove failed: unknown album"));
        return;
      }

      const uid = this.album.UID;

      this.dialog.album = false;

      Api.delete(`albums/${uid}/photos`, {"data": {"photos": this.selection}}).then(() => this.onRemoved());
    },
    onRemoved() {
      this.clearClipboard();
    },
    download() {
      switch (this.selection.length) {
        case 0: return;
        case 1: new Photo().find(this.selection[0]).then(p => p.downloadAll()); break;
        default: Api.post("zip", {"photos": this.selection}).then(r => {
          this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken()}`);
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
      Event.PubSub.publish("dialog.edit", {selection: this.selection, album: this.album, index: 0});
    },
    onShared() {
      this.dialog.share = false;
      this.clearClipboard();
    },
  }
};
</script>
