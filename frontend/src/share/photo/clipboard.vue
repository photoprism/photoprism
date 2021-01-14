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
  }
};
</script>
