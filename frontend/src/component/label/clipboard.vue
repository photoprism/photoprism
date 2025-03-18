<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard p-label-clipboard`"
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

        <!-- v-btn key="download" :title="$gettext('Download')" icon="mdi-download" color="download" density="comfortable class="action-download" :disabled="selection.length !== 1" @click.stop="download()"></v-btn -->
        <v-btn
          key="action-album"
          :title="$gettext('Add to album')"
          icon="mdi-bookmark"
          color="album"
          density="comfortable"
          :disabled="!canAddAlbums || selection.length === 0"
          class="action-album"
          @click.stop="dialog.album = true"
        ></v-btn>
        <v-btn
          key="action-delete"
          :title="$gettext('Delete')"
          icon="mdi-delete"
          color="remove"
          density="comfortable"
          :disabled="!canManage || selection.length === 0"
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
      @confirm="addToAlbum"
    ></p-photo-album-dialog>
    <p-label-delete-dialog
      :visible="dialog.delete"
      @close="dialog.delete = false"
      @confirm="batchDelete"
    ></p-label-delete-dialog>
  </div>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import download from "common/download";
import PPhotoAlbumDialog from "component/photo/album/dialog.vue";
import PLabelDeleteDialog from "component/label/delete/dialog.vue";

export default {
  name: "PLabelClipboard",
  components: {
    PPhotoAlbumDialog,
    PLabelDeleteDialog,
  },
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
      rtl: this.$isRtl,
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

      $api.post(`albums/${ppid}/photos`, { labels: this.selection }).then(() => this.onAdded());
    },
    onAdded() {
      this.clearClipboard();
    },
    batchDelete() {
      if (!this.canManage) {
        return;
      }

      this.dialog.delete = false;

      $api.post("batch/labels/delete", { labels: this.selection }).then(this.onDeleted.bind(this));
    },
    onDeleted() {
      $notify.success(this.$gettext("Labels deleted"));
      this.clearClipboard();
    },
    download() {
      if (!this.canDownload) {
        return;
      }

      if (this.selection.length !== 1) {
        $notify.error(this.$gettext("You can only download one label"));
        return;
      }

      this.onDownload(`${this.$config.apiUri}/labels/${this.selection[0]}/dl?t=${this.$config.downloadToken}`);

      this.expanded = false;
    },
    onDownload(path) {
      $notify.success(this.$gettext("Downloading…"));

      download(path, "label.zip");
    },
  },
};
</script>
