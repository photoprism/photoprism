<template>
  <div>
    <div v-if="selection.length > 0" class="clipboard-container">
      <v-speed-dial
        id="t-clipboard"
        v-model="expanded"
        :class="`p-clipboard p-people-clipboard`"
        :end="!rtl"
        :start="rtl"
        :attach="true"
        location="top"
        transition="slide-y-reverse-transition"
        offset="12"
      >
        <template #activator="{ props }">
          <v-btn v-bind="props" icon size="52" color="highlight" variant="elevated" density="comfortable" class="action-menu opacity-95 ma-5">
            <span class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn-toggle v-if="selection.length > 1" v-model="searchMode" mandatory density="compact" rounded="pill" variant="elevated" class="action-search-mode">
          <v-btn value="or" density="compact" class="action-search-or">OR</v-btn>
          <v-btn value="and" density="compact" class="action-search-and">AND</v-btn>
        </v-btn-toggle>
        <v-btn
          key="action-search"
          :title="$gettext('Search')"
          icon="mdi-magnify"
          color="primary"
          density="comfortable"
          class="action-search"
          :disabled="!canSearch || selection.length === 0"
          @click.stop="search()"
        ></v-btn>
        <v-btn
          key="action-download"
          :title="$gettext('Download')"
          icon="mdi-download"
          color="download"
          density="comfortable"
          class="action-download"
          :disabled="!canDownload || selection.length !== 1"
          @click.stop="download()"
        ></v-btn>
        <v-btn
          v-if="canAddAlbums"
          key="action-album"
          :title="$gettext('Add to album')"
          icon="mdi-bookmark"
          color="album"
          density="comfortable"
          :disabled="selection.length === 0"
          class="action-album"
          @click.stop="dialog.album = true"
        ></v-btn>
        <v-btn key="action-close" icon="mdi-close" color="grey-darken-2" density="comfortable" class="action-clear" @click.stop="clearClipboard()"></v-btn>
      </v-speed-dial>
    </div>
    <p-photo-album-dialog :visible="dialog.album" @close="dialog.album = false" @confirm="addToAlbum"></p-photo-album-dialog>
  </div>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import download from "common/download";
import PPhotoAlbumDialog from "component/photo/album/dialog.vue";

export default {
  name: "PPeopleClipboard",
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
    return {
      canManage: this.$config.allow("people", "manage"),
      canSearch: this.$config.allow("photos", "search") && this.$config.feature("search"),
      canDownload: this.$config.allow("people", "download") && this.$config.feature("download"),
      canAddAlbums: this.$config.allow("albums", "create") && this.$config.feature("albums"),
      features: this.$config.getSettings().features,
      expanded: false,
      searchMode: "or",
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
    // search opens selected people in photo search with the selected operator.
    search() {
      if (!this.canSearch || this.selection.length === 0) {
        return;
      }

      const subjects = this.selection.filter((uid) => uid).join(this.searchMode === "and" ? "&" : "|");

      if (!subjects) {
        return;
      }

      this.$router.push({ name: "all", query: { q: `subject:${subjects}` } });
      this.clearClipboard();
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
      const body = { subjects: this.selection };

      Promise.all(uniqueAlbumUids.map((uid) => $api.post(`albums/${uid}/photos`, body)))
        .then(() => this.onAdded())
        .catch(() => {
          $notify.error(this.$gettext("Some albums could not be updated"));
        });
    },
    onAdded() {
      this.clearClipboard();
    },
    download() {
      if (this.selection.length !== 1) {
        $notify.error(this.$gettext("You can only download one album"));
        return;
      }

      $notify.success(this.$gettext("Downloading…"));

      $api.post("zip", { subjects: this.selection }).then((r) => {
        this.onDownload(`${this.$config.apiUri}/zip/${r.data.filename}?t=${this.$config.downloadToken}`);
      });

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photos.zip");
    },
  },
};
</script>
