<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial
          id="t-clipboard" v-model="expanded" fixed
          bottom
          right
          direction="top"
          transition="slide-y-reverse-transition"
          class="p-clipboard p-photo-clipboard"
      >
        <v-btn
            slot="activator" fab
            dark
            color="accent darken-2"
            class="action-menu"
        >
          <v-icon v-if="selection.length === 0">menu</v-icon>
          <span v-else class="count-clipboard">{{ selection.length }}</span>
        </v-btn>

        <v-btn
            v-if="context !== 'archive' && context !== 'review' && config.settings.features.share" fab dark
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
            v-if="context !== 'archive' && config.settings.features.edit" fab dark
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
            v-if="context !== 'archive' && config.settings.features.private" fab dark
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
            v-if="context !== 'archive' && config.settings.features.download" fab dark
            small
            :title="$gettext('Download')"
            color="download"
            class="action-download"
            @click.stop="download()"
        >
          <v-icon>get_app</v-icon>
        </v-btn>
        <v-btn
            v-if="context !== 'archive'" fab dark
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
            v-if="!isAlbum && context !== 'archive' && config.settings.features.archive" fab dark
            small
            color="remove"
            :title="$gettext('Archive')"
            :disabled="selection.length === 0"
            class="action-archive"
            @click.stop="dialog.archive = true"
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
            @click.stop="batchRestorePhotos"
        >
          <v-icon>unarchive</v-icon>
        </v-btn>
        <v-btn
            v-if="isAlbum" fab dark
            small
            :title="$gettext('Remove')"
            color="remove"
            :disabled="selection.length === 0"
            class="action-delete"
            @click.stop="removeFromAlbum"
        >
          <v-icon>eject</v-icon>
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
    <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false"
                          @confirm="addToAlbum"></p-photo-album-dialog>
    <p-photo-archive-dialog :show="dialog.archive" @cancel="dialog.archive = false"
                            @confirm="batchArchivePhotos"></p-photo-archive-dialog>
    <p-share-upload-dialog :show="dialog.share" :selection="selection" :album="album" @cancel="dialog.share = false"
                           @confirm="onShared"></p-share-upload-dialog>
  </div>
</template>
<script>
import Api from "common/api";
import Notify from "common/notify";
import Event from "pubsub-js";

export default {
  name: 'PPhotoClipboard',
  props: {
    selection: Array,
    refresh: Function,
    album: Object,
    context: String,
  },
  data() {
    return {
      config: this.$config.values,
      expanded: false,
      isAlbum: this.album && this.album.Type === 'album',
      dialog: {
        archive: false,
        album: false,
        share: false,
      },
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
    batchArchivePhotos() {
      this.dialog.archive = false;

      Api.post("batch/photos/archive", {"photos": this.selection}).then(() => this.onArchived());
    },
    onArchived() {
      Notify.success(this.$gettext("Selection archived"));
      this.clearClipboard();
    },
    batchPrivate() {
      Api.post("batch/photos/private", {"photos": this.selection}).then(() => this.onPrivateSaved());
    },
    onPrivateSaved() {
      this.clearClipboard();
    },
    batchRestorePhotos() {
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
      if (this.selection.length === 1) {
        this.onDownload(`/api/v1/photos/${this.selection[0]}/dl?t=${this.$config.downloadToken()}`);
      } else {
        Api.post("zip", {"photos": this.selection}).then(r => {
          this.onDownload(`/api/v1/zip/${r.data.filename}?t=${this.$config.downloadToken()}`);
        });
      }

      this.expanded = false;
    },
    onDownload(path) {
      Notify.success(this.$gettext("Downloading…"));
      const link = document.createElement('a');
      link.href = path;
      link.download = "photos.zip";
      link.click();
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
