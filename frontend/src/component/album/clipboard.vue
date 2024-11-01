<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
<!--      TODO: check v-speed-dial view and position-->
      <v-speed-dial id="t-clipboard" location="bottom right" v-model="expanded" fixed bottom direction="top" transition="slide-y-reverse-transition" :end="!rtl" :start="rtl" :class="`p-clipboard ${!rtl ? '--ltr' : '--rtl'} p-album-clipboard`">
        <template #activator>
          <v-btn theme="dark" color="accent-darken-2 rounded-circle" class="action-menu">
            <!-- TODO: change this icon -->
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn v-if="canShare" theme="dark" size="small" :title="$gettext('Share')" color="share" :disabled="selection.length !== 1" class="action-share rounded-circle" @click.stop="shareDialog()">
          <v-icon>mdi-share-variant</v-icon>
        </v-btn>
        <v-btn v-if="canManage" theme="dark" size="small" :title="$gettext('Edit')" color="edit" :disabled="selection.length !== 1" class="action-edit rounded-circle" @click.stop="editDialog()">
          <v-icon>mdi-pencil</v-icon>
        </v-btn>
        <v-btn theme="dark" size="small" :title="$gettext('Download')" color="download" class="action-download rounded-circle" :disabled="!canDownload || selection.length !== 1" @click.stop="download()">
          <v-icon>mdi-download</v-icon>
        </v-btn>
        <v-btn v-if="canManage" theme="dark" size="small" :title="$gettext('Add to album')" color="album" :disabled="selection.length === 0" class="action-clone rounded-circle" @click.stop="dialog.album = true">
          <v-icon>mdi-bookmark</v-icon>
        </v-btn>
        <v-btn v-if="canDelete && deletable.includes(context)" theme="dark" size="small" color="remove" :title="$gettext('Delete')" :disabled="selection.length === 0" class="action-delete rounded-circle" @click.stop="dialog.delete = true">
          <v-icon>mdi-delete</v-icon>
        </v-btn>
        <v-btn theme="dark" size="small" color="accent" class="action-clear rounded-circle" @click.stop="clearClipboard()">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-speed-dial>
    </v-container>
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
