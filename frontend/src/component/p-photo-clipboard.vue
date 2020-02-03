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
                    class="p-clipboard p-photo-clipboard"
                    id="t-clipboard"
            >
                <v-btn
                        slot="activator"
                        color="accent darken-2"
                        dark
                        fab
                        class="p-photo-clipboard-menu"
                >
                    <v-icon v-if="selection.length === 0">menu</v-icon>
                    <span v-else class="t-clipboard-count">{{ selection.length }}</span>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        :title="labels.share"
                        color="share"
                        @click.stop="dialog.share = true"
                        :disabled="selection.length === 0"
                        v-if="context !== 'archive'"
                        class="p-photo-clipboard-private"
                >
                    <v-icon>share</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.edit"
                        color="edit"
                        :disabled="selection.length === 0"
                        @click.stop="dialog.edit = true"
                        v-if="context !== 'archive'"
                        class="p-photo-clipboard-edit"
                >
                    <v-icon>edit</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.download"
                        color="download"
                        @click.stop="download()"
                        v-if="context !== 'archive'"
                        class="p-photo-clipboard-download"
                >
                    <v-icon>cloud_download</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.addToAlbum"
                        color="album"
                        :disabled="selection.length === 0"
                        @click.stop="dialog.album = true"
                        v-if="context !== 'archive'"
                        class="p-photo-clipboard-album"
                >
                    <v-icon>folder</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="remove"
                        :title="labels.archive"
                        @click.stop="dialog.archive = true"
                        :disabled="selection.length === 0"
                        v-if="!album && context !== 'archive'"
                        class="p-photo-clipboard-archive"
                >
                    <v-icon>archive</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="restore"
                        :title="labels.restore"
                        @click.stop="batchRestorePhotos"
                        :disabled="selection.length === 0"
                        v-if="!album && context === 'archive'"
                        class="p-photo-clipboard-restore"
                >
                    <v-icon>unarchive</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        :title="labels.removeFromAlbum"
                        color="remove"
                        @click.stop="removeFromAlbum"
                        :disabled="selection.length === 0"
                        v-if="album"
                        class="p-photo-clipboard-delete"
                >
                    <v-icon>remove</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="accent"
                        @click.stop="clearClipboard()"
                        class="p-photo-clipboard-clear"
                >
                    <v-icon>clear</v-icon>
                </v-btn>
            </v-speed-dial>
        </v-container>
        <p-photo-album-dialog :show="dialog.album" @cancel="dialog.album = false"
                              @confirm="addToAlbum"></p-photo-album-dialog>
        <p-photo-archive-dialog :show="dialog.archive" @cancel="dialog.archive = false"
                               @confirm="batchArchivePhotos"></p-photo-archive-dialog>
        <p-photo-edit-dialog :show="dialog.edit" :selection="selection" :album="album" @cancel="dialog.edit = false"
                             @confirm="dialog.edit = fals"></p-photo-edit-dialog>
        <p-photo-share-dialog :show="dialog.share" :selection="selection" :album="album" @cancel="dialog.share = false"
                             @confirm="dialog.share = false"></p-photo-share-dialog>
    </div>
</template>
<script>
    import Api from "common/api";
    import Notify from "common/notify";

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
                expanded: false,
                dialog: {
                    archive: false,
                    album: false,
                    edit: false,
                    share: false,
                },
                labels: {
                    share: this.$gettext("Share"),
                    edit: this.$gettext("Edit"),
                    story: this.$gettext("Story"),
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
            batchArchivePhotos() {
                this.dialog.archive = false;

                Api.post("batch/photos/archive", {"photos": this.selection}).then(() => this.onArchived());
            },
            onArchived() {
                Notify.success(this.$gettext("Photos archived"));
                this.clearClipboard();
            },
            batchRestorePhotos() {
                Api.post("batch/photos/restore", {"photos": this.selection}).then(() => this.onRestored());
            },
            onRestored() {
                Notify.success(this.$gettext("Photos restored"));
                this.clearClipboard();
            },
            addToAlbum(albumUUID) {
                this.dialog.album = false;

                Api.post(`albums/${albumUUID}/photos`, {"photos": this.selection}).then(() => this.onAdded());
            },
            onAdded() {
                this.clearClipboard();
            },
            removeFromAlbum() {
                if(!this.album) {
                    this.$notify.error(this.$gettext("remove failed: unknown album"));
                    return
                }

                const albumUUID = this.album.AlbumUUID;

                this.dialog.album = false;

                Api.delete(`albums/${albumUUID}/photos`, {"data": {"photos": this.selection}}).then(() => this.onRemoved());
            },
            onRemoved() {
                this.clearClipboard();
            },
            download() {
                if(this.selection.length === 1) {
                    this.onDownload(`/api/v1/photos/${this.selection[0]}/download`);
                } else {
                    Api.post("zip", {"photos": this.selection}).then(r => {
                        this.onDownload("/api/v1/zip/" + r.data.filename);
                    });
                }

                this.expanded = false;
            },
            onDownload(path) {
                Notify.success(this.$gettext("Downloading..."));
                window.open(path, "_blank");
            },
        }
    };
</script>
