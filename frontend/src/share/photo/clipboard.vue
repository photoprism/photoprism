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
            v-if="context !== 'archive'" fab dark
            small
            :title="$gettext('Download')"
            color="download"
            :disabled="!$config.feature('download')"
            class="action-download"
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
import Api from "common/api";
import Notify from "common/notify";
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
      expanded: false,
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
  }
};
</script>
