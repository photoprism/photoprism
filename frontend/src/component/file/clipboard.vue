<template>
  <div>
    <v-container fluid class="pa-0" v-if="selection.length > 0">
      <v-speed-dial
              fixed bottom right
              direction="top"
              v-model="expanded"
              transition="slide-y-reverse-transition"
              class="p-clipboard p-file-clipboard"
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
                v-if="$config.feature('download')"
                :disabled="selection.length === 0"
        >
          <v-icon>get_app</v-icon>
        </v-btn>

        <v-btn
                fab dark small
                :title="labels.addToAlbum"
                color="album"
                :disabled="selection.length === 0"
                @click.stop="dialog.album = true"
                class="action-album"
        >
          <v-icon>folder</v-icon>
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
  </div>
</template>
<script>
    import Api from "common/api";
    import Notify from "common/notify";

    export default {
        name: 'p-file-clipboard',
        props: {
            selection: Array,
            refresh: Function,
            clearSelection: Function,
        },
        data() {
            return {
                expanded: false,
                dialog: {
                    album: false,
                    edit: false,
                },
                labels: {
                    download: this.$gettext("Download"),
                    addToAlbum: this.$gettext("Add to album"),
                    removeFromAlbum: this.$gettext("Remove"),
                },

            };
        },
        methods: {
            clearClipboard() {
                this.clearSelection();
                this.expanded = false;
            },
            addToAlbum(ppid) {
                this.dialog.album = false;

                Api.post(`albums/${ppid}/photos`, {"files": this.selection}).then(() => this.onAdded());
            },
            onAdded() {
                this.clearClipboard();
            },
            download() {
                Api.post("zip", {"files": this.selection}).then(r => {
                    this.onDownload("/api/v1/zip/" + r.data.filename + "?t=" + this.$config.downloadToken());
                });

                this.expanded = false;
            },
            onDownload(path) {
                Notify.success(this.$gettext("Downloading…"));
                const link = document.createElement('a')
                link.href = path;
                link.download = "photos.zip";
                link.click();
            },
        }
    };
</script>
