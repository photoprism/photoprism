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
                        :title="labels.share"
                        color="share"
                        @click.stop="share()"
                        :disabled="selection.length !== 1"
                        v-if="$config.feature('share')"
                        class="action-share"
                >
                    <v-icon>share</v-icon>
                </v-btn>
                <v-btn
                        fab dark small
                        :title="labels.download"
                        color="download"
                        @click.stop="download()"
                        class="action-download"
                        v-if="$config.feature('download')"
                        :disabled="selection.length !== 1"
                >
                    <v-icon>get_app</v-icon>
                </v-btn>
                <v-btn
                        fab dark small
                        :title="labels.clone"
                        color="album"
                        :disabled="selection.length === 0"
                        @click.stop="dialog.album = true"
                        class="action-clone"
                >
                    <v-icon>folder</v-icon>
                </v-btn>
                <v-btn
                        fab dark small
                        color="remove"
                        :title="labels.delete"
                        @click.stop="dialog.delete = true"
                        :disabled="selection.length === 0"
                        class="action-delete"
                >
                    <v-icon>delete</v-icon>
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
                              @confirm="cloneAlbums"></p-photo-album-dialog>
        <p-album-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false"
                               @confirm="batchDelete"></p-album-delete-dialog>
        <p-share-dialog :show="dialog.share" title="Share Album" :model="model" @upload="upload"
                        @close="dialog.share = false"></p-share-dialog>
        <p-share-upload-dialog :show="dialog.upload" :selection="selection" @cancel="dialog.upload = false"
                               @confirm="dialog.upload = false"></p-share-upload-dialog>
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
            share() {
                if (this.selection.length !== 1) {
                    this.$notify.error("select one album to share");
                    return;
                }

                this.model = new Album();
                this.model.find(this.selection[0]).then(
                    (m) => {
                        this.model = m;
                        this.dialog.share = true;
                    }
                );
            },
            upload() {
                this.dialog.share = false;
                this.dialog.upload = true;
            },
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
                if (this.selection.length !== 1) {
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
