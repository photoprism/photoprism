<template>
  <div>
    <v-container fluid class="pa-0" v-if="selection.length > 0">
      <v-speed-dial
              fixed bottom right
              direction="top"
              v-model="expanded"
              transition="slide-y-reverse-transition"
              class="p-clipboard p-photo-clipboard"
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
                :disabled="!$config.feature('download')"
                v-if="context !== 'archive'"
                class="action-download"
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
    import Event from "pubsub-js";

    export default {
        name: 'p-photo-clipboard',
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
                labels: {
                    share: this.$gettext("Share"),
                    private: this.$gettext("Change private flag"),
                    edit: this.$gettext("Edit"),
                    addToAlbum: this.$gettext("Add to album"),
                    removeFromAlbum: this.$gettext("Remove"),
                    archive: this.$gettext("Archive"),
                    restore: this.$gettext("Restore"),
                    download: this.$gettext("Download"),
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
                const link = document.createElement('a')
                link.href = path;
                link.download = "photos.zip";
                link.click();
            },
        }
    };
</script>
