<template>
    <div>
        <v-container fluid class="pa-0" v-if="selection.length > 0">
            <v-speed-dial
                    fixed
                    bottom
                    right
                    direction="top"
                    v-model="expanded"
                    transition="slide-y-reverse-transition"
                    class="p-clipboard p-album-clipboard"
                    id="t-clipboard"
            >
                <v-btn
                        slot="activator"
                        color="accent darken-2"
                        dark
                        fab
                        class="p-album-clipboard-menu"
                >
                    <v-icon v-if="selection.length === 0">menu</v-icon>
                    <span v-else  class="t-clipboard-count">{{ selection.length }}</span>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        :title="labels.download"
                        color="download"
                        @click.stop="download()"
                        class="p-album-clipboard-download"
                        v-if="$config.feature('download')"
                        :disabled="selection.length !== 1"
                >
                    <v-icon>get_app</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.clone"
                        color="album"
                        :disabled="selection.length === 0"
                        @click.stop="dialog.album = true"
                        class="p-album-clipboard-clone"
                >
                    <v-icon>folder</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="remove"
                        :title="labels.delete"
                        @click.stop="dialog.delete = true"
                        :disabled="selection.length === 0"
                        class="p-album-clipboard-delete"
                >
                    <v-icon>delete</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="accent"
                        @click.stop="clearClipboard()"
                        class="p-album-clipboard-clear"
                >
                    <v-icon>clear</v-icon>
                </v-btn>
            </v-speed-dial>
        </v-container>
        <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false"
                              @confirm="cloneAlbums"></p-photo-album-dialog>
        <p-album-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false"
                               @confirm="batchDelete"></p-album-delete-dialog>
    </div>
</template>
<script>
    import Api from "common/api";
    import Notify from "common/notify";

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
                dialog: {
                    delete: false,
                    album: false,
                    edit: false,
                },
                labels: {
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
            cloneAlbums(ppid) {
                this.dialog.album = false;

                Api.post(`albums/${ppid}/clone`, {"albums": this.selection}).then(() => this.onCloned());
            },
            onCloned() {
                this.clearClipboard();
            },
            batchDelete() {
                this.dialog.delete = false;

                Api.post("batch/albums/delete", {"albums": this.selection}).then(this.onDeleted.bind(this));
            },
            onDeleted() {
                Notify.success(this.$gettext("Albums deleted"));
                this.clearClipboard();
            },
            download() {
                if(this.selection.length !== 1) {
                    Notify.error(this.$gettext("You can only download one album"));
                    return;
                }

                this.onDownload(`/api/v1/albums/${this.selection[0]}/dl?t=${this.$config.downloadToken()}`);

                this.expanded = false;
            },
            onDownload(path) {
                Notify.success(this.$gettext("Downloading..."));
                const link = document.createElement('a')
                link.href = path;
                link.download = "album.zip";
                link.click();
            },
        }
    };
</script>
