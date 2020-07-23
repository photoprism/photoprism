<template>
  <div>
    <v-container fluid class="pa-0" v-if="selection.length > 0">
      <v-speed-dial
              fixed bottom right
              direction="top"
              v-model="expanded"
              transition="slide-y-reverse-transition"
              class="p-clipboard p-album-clipboard"
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
                :title="labels.download"
                color="download"
                @click.stop="download()"
                class="action-download"
                :disabled="selection.length !== 1 || !$config.feature('download')"
        >
          <v-icon>get_app</v-icon>
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
  </div>
</template>
<script>
    import Api from "common/api";
    import Notify from "common/notify";
    import Album from "model/album";

    export default {
        name: 'p-album-clipboard',
        props: {
            selection: Array,
            refresh: Function,
            clearSelection: Function,
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
                labels: {
                    share: this.$gettext("Share"),
                    download: this.$gettext("Download"),
                    clone: this.$gettext("Add to album"),
                    delete: this.$gettext("Delete"),
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

                this.onDownload(`/api/v1/albums/${this.selection[0]}/dl?t=${this.$config.downloadToken()}`);

                this.expanded = false;
            },
            onDownload(path) {
                Notify.success(this.$gettext("Downloading…"));
                const link = document.createElement('a')
                link.href = path;
                link.download = "album.zip";
                link.click();
            },
        }
    };
</script>
