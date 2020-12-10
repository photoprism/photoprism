<template>
  <div>
    <v-container fluid class="pa-0" v-if="selection.length > 0">
      <v-speed-dial
          fixed bottom right
          direction="top"
          v-model="expanded"
          transition="slide-y-reverse-transition"
          class="p-clipboard p-photo-clipboard"
          id="t-clipboard"
      >
        <v-btn
            fab dark
            slot="activator"
            color="accent darken-2"
            class="action-menu"
        >
          <v-icon v-if="selection.length === 0">menu</v-icon>
          <span v-else class="count-clipboard">{{ selection.length }}</span>
        </v-btn>

        <v-btn
            fab dark small
            :title="$gettext('Share')"
            color="share"
            @click.stop="dialog.share = true"
            :disabled="selection.length === 0"
            v-if="context !== 'archive' && context !== 'review' && $config.feature('share')"
            class="action-share"
        >
          <v-icon>cloud</v-icon>
        </v-btn>

        <v-btn
            fab dark small
            :title="$gettext('Approve')"
            color="share"
            @click.stop="batchApprove"
            :disabled="selection.length === 0"
            v-if="context === 'review'"
            class="action-approve"
        >
          <v-icon>check</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            :title="$gettext('Edit')"
            color="edit"
            :disabled="selection.length === 0"
            @click.stop="edit"
            v-if="context !== 'archive' && $config.feature('edit')"
            class="action-edit"
        >
          <v-icon>edit</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            :title="$gettext('Change private flag')"
            color="private"
            :disabled="selection.length === 0"
            @click.stop="batchPrivate"
            v-if="context !== 'archive' && config.settings.features.private"
            class="action-private"
        >
          <v-icon>lock</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            :title="$gettext('Download')"
            color="download"
            @click.stop="download()"
            v-if="context !== 'archive' && $config.feature('download')"
            class="action-download"
        >
          <v-icon>get_app</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            :title="$gettext('Add to album')"
            color="album"
            :disabled="selection.length === 0"
            @click.stop="dialog.album = true"
            v-if="context !== 'archive'"
            class="action-album"
        >
          <v-icon>folder_special</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            color="remove"
            :title="$gettext('Archive')"
            @click.stop="dialog.archive = true"
            :disabled="selection.length === 0"
            v-if="!isAlbum && context !== 'archive' && $config.feature('archive')"
            class="action-archive"
        >
          <v-icon>archive</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            color="restore"
            :title="$gettext('Restore')"
            @click.stop="batchRestorePhotos"
            :disabled="selection.length === 0"
            v-if="!album && context === 'archive'"
            class="action-restore"
        >
          <v-icon>unarchive</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            :title="$gettext('Remove')"
            color="remove"
            @click.stop="removeFromAlbum"
            :disabled="selection.length === 0"
            v-if="isAlbum"
            class="action-delete"
        >
          <v-icon>remove</v-icon>
        </v-btn>
        <v-btn
            fab dark small
            color="accent"
            @click.stop="clearClipboard()"
            class="action-clear"
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
  name: 'p-photo-clipboard',
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
        return
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
      const link = document.createElement('a')
      link.href = path;
      link.download = "photos.zip";
      link.click();
    },
    edit() {
      // Open Edit Dialog
      Event.publish("dialog.edit", {selection: this.selection, album: this.album, index: 0});
    },
    onShared() {
      this.dialog.share = false;
      this.clearClipboard();
    },
  }
};
</script>
