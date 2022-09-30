<template>
  <div>
    <v-container v-if="selection.length > 0" fluid class="pa-0">
      <v-speed-dial
          id="t-clipboard" v-model="expanded" fixed
          bottom
          right
          direction="top"
          transition="slide-y-reverse-transition"
          class="p-clipboard p-album-clipboard"
      >
        <template #activator>
          <v-btn
              fab dark
              color="accent darken-2"
              class="action-menu"
          >
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else class="count-clipboard">{{ selection.length }}</span>
          </v-btn>
        </template>

        <v-btn
            fab dark small
            :title="$gettext('Download')"
            color="download"
            class="action-download"
            :disabled="selection.length !== 1 || !$config.feature('download')"
            @click.stop="download()"
        >
          <v-icon>get_app</v-icon>
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
  </div>
</template>
<script>
import Notify from "common/notify";
import Album from "model/album";
import download from "common/download";

export default {
  name: 'PAlbumClipboard',
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
    context: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      expanded: false,
      model: new Album(),
      dialog: {
        delete: false,
        album: false,
        edit: false,
        share: false,
        upload: false,
      },
    };
  },
  methods: {
    clearClipboard() {
      this.clearSelection();
      this.expanded = false;
    },
    download() {
      if (this.selection.length !== 1) {
        Notify.error(this.$gettext("You can only download one album"));
        return;
      }

      Notify.success(this.$gettext("Downloading…"));

      this.onDownload(`${this.$config.apiUri}/albums/${this.selection[0]}/dl?t=${this.$config.downloadToken()}`);

      this.expanded = false;
    },
    onDownload(path) {
      download(path, "photoprism-album.zip");
    },
  }
};
</script>
