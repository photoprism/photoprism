<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard p-file-clipboard`"
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
          v-if="canDownload"
          key="action-download"
          :title="$gettext('Download')"
          icon="mdi-download"
          color="download"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0"
          class="action-download"
          @click.stop="download()"
        ></v-btn>
        <v-btn
          v-if="canManage"
          key="action-album"
          :title="$gettext('Add to album')"
          icon="mdi-bookmark"
          color="album"
          variant="elevated"
          density="comfortable"
          :disabled="selection.length === 0"
          class="action-album"
          @click.stop="dialog.album = true"
        ></v-btn>
        <v-btn
          key="action-close"
          icon="mdi-close"
          color="grey-darken-2"
          variant="elevated"
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
  </div>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import download from "common/download";
import PPhotoAlbumDialog from "component/photo/album/dialog.vue";

export default {
  name: "PFileClipboard",
  components: {
    PPhotoAlbumDialog,
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
    const features = this.$config.getSettings().features;

    return {
      expanded: false,
      canDownload: this.$config.allow("photos", "download") && features.download,
      canShare: this.$config.allow("photos", "share") && features.share,
      canManage: this.$config.allow("photos", "manage") && features.albums,
      dialog: {
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
    addToAlbum(ppidOrList) {
      if (!ppidOrList) {
        return;
      }

      // Validate array input
      if (Array.isArray(ppidOrList) && ppidOrList.length === 0) {
        return;
      }

      this.dialog.album = false;

      const albumUids = Array.isArray(ppidOrList) ? ppidOrList : [ppidOrList];
      // Deduplicate album UIDs
      const uniqueAlbumUids = [...new Set(albumUids.filter((uid) => uid))];
      const body = { files: this.selection };

      Promise.all(uniqueAlbumUids.map((uid) => $api.post(`albums/${uid}/photos`, body)))
        .then(() => this.onAdded())
        .catch((error) => {
          $notify.error(this.$gettext("Some albums could not be updated"));
        });
    },
    onAdded() {
      this.clearClipboard();
    },
    download() {
      $api.post("zip", { files: this.selection }).then((r) => {
        this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken}`);
      });

      this.expanded = false;
    },
    onDownload(path) {
      $notify.success(this.$gettext("Downloading…"));

      download(path, "photos.zip");
    },
  },
};
</script>
