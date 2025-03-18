<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard p-album-clipboard`"
        :end="!rtl"
        :start="rtl"
        :attach="true"
        location="top"
        transition="slide-y-reverse-transition"
        offset="12"
      >
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            icon
            size="52"
            color="highlight"
            variant="elevated"
            density="comfortable"
            class="action-menu opacity-95 ma-5"
          >
            <span class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn
          v-if="canShare"
          key="action-share"
          :title="$gettext('Share')"
          icon="mdi-share-variant"
          color="share"
          density="comfortable"
          :disabled="selection.length !== 1"
          class="action-share"
          @click.stop="shareDialog()"
        ></v-btn>
        <v-btn
          v-if="canManage"
          key="action-edit"
          :title="$gettext('Edit')"
          icon="mdi-pencil"
          color="edit"
          density="comfortable"
          :disabled="selection.length !== 1"
          class="action-edit"
          @click.stop="editDialog()"
        ></v-btn>
        <v-btn
          v-if="canDownload"
          key="action-download"
          :title="$gettext('Download')"
          icon="mdi-download"
          color="download"
          density="comfortable"
          class="action-download"
          :disabled="selection.length !== 1"
          @click.stop="download()"
        ></v-btn>
        <v-btn
          v-if="canManage"
          key="action-album"
          :title="$gettext('Add to album')"
          icon="mdi-bookmark"
          color="album"
          density="comfortable"
          :disabled="selection.length === 0"
          class="action-clone"
          @click.stop="dialog.album = true"
        ></v-btn>
        <v-btn
          v-if="canDelete && deletable.includes(context)"
          key="action-delete"
          :title="$gettext('Delete')"
          icon="mdi-delete"
          color="remove"
          density="comfortable"
          :disabled="selection.length === 0"
          class="action-delete"
          @click.stop="dialog.delete = true"
        ></v-btn>
        <v-btn
          key="action-close"
          icon="mdi-close"
          color="grey-darken-2"
          density="comfortable"
          class="action-clear"
          @click.stop="clearClipboard()"
        ></v-btn>
      </v-speed-dial>
    </div>
    <p-photo-album-dialog
      :visible="dialog.album"
      @close="dialog.album = false"
      @confirm="cloneAlbums"
    ></p-photo-album-dialog>
    <p-album-delete-dialog
      :visible="dialog.delete"
      @close="dialog.delete = false"
      @confirm="batchDelete"
    ></p-album-delete-dialog>
  </div>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
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
    const settings = this.$config.getSettings();
    const features = settings.features;

    return {
      canDelete: this.$config.allow("albums", "delete"),
      canDownload:
        this.$config.allow("albums", "download") && features.download && !settings?.albums?.download?.disabled,
      canShare: this.$config.allow("albums", "share") && features.share,
      canManage: this.$config.allow("albums", "manage"),
      deletable: ["album", "moment", "state"],
      expanded: false,
      dialog: {
        delete: false,
        album: false,
        edit: false,
      },
      rtl: this.$isRtl,
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

      $api.post(`albums/${ppid}/clone`, { albums: this.selection }).then(() => this.onCloned());
    },
    onCloned() {
      this.clearClipboard();
    },
    batchDelete() {
      this.dialog.delete = false;

      $api.post("batch/albums/delete", { albums: this.selection }).then(this.onDeleted.bind(this));
    },
    onDeleted() {
      $notify.success(this.$gettext("Albums deleted"));
      this.clearClipboard();
    },
    download() {
      if (this.selection.length !== 1) {
        $notify.error(this.$gettext("You can only download one album"));
        return;
      }

      $notify.success(this.$gettext("Downloading…"));

      this.onDownload(`${this.$config.apiUri}/albums/${this.selection[0]}/dl?t=${this.$config.downloadToken}`);

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photoprism-album.zip");
    },
  },
};
</script>
