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
            >
                <v-btn
                        slot="activator"
                        color="accent darken-2"
                        dark
                        fab
                        class="p-photo-clipboard-menu"
                >
                    <v-icon v-if="selection.length === 0">menu</v-icon>
                    <span v-else>{{ selection.length }}</span>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        :title="labels.private"
                        color="deep-purple lighten-2"
                        @click.stop="batchPrivate()"
                        :disabled="selection.length === 0"
                        class="p-photo-clipboard-private"
                >
                    <v-icon>vpn_key</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.story"
                        color="cyan accent-4"
                        :disabled="selection.length === 0"
                        @click.stop="batchStory()"
                        class="p-photo-clipboard-story"
                >
                    <v-icon>wifi</v-icon>
                </v-btn>
                <!-- v-btn
                        fab
                        dark
                        small
                        color="light-blue accent-4"
                        @click.stop="openDocs()"
                        class="p-photo-clipboard-docs"
                >
                    <v-icon>info</v-icon>
                </v-btn -->
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.download"
                        color="teal accent-4"
                        @click.stop="download()"
                        class="p-photo-clipboard-download"
                >
                    <v-icon>save</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        :title="labels.addToAlbum"
                        color="yellow accent-4"
                        :disabled="selection.length === 0"
                        @click.stop="dialog.album = true"
                        class="p-photo-clipboard-album"
                >
                    <v-icon>create_new_folder</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="delete"
                        :title="labels.delete"
                        @click.stop="dialog.delete = true"
                        :disabled="selection.length === 0"
                        v-if="!album"
                        class="p-photo-clipboard-delete"
                >
                    <v-icon>delete</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        :title="labels.removeFromAlbum"
                        color="delete"
                        @click.stop="removeFromAlbum"
                        :disabled="selection.length === 0"
                        v-if="album"
                        class="p-photo-clipboard-delete"
                >
                    <v-icon>remove_circle</v-icon>
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
        <p-photo-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false"
                               @confirm="batchDeletePhotos"></p-photo-delete-dialog>
        <p-photo-edit-dialog :show="dialog.edit" @cancel="dialog.edit = false"
                             @confirm="batchEditPhotos"></p-photo-edit-dialog>
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
                    private: this.$gettext("Private"),
                    story: this.$gettext("Story"),
                    addToAlbum: this.$gettext("Add to album"),
                    removeFromAlbum: this.$gettext("Remove from album"),
                    delete: this.$gettext("Delete"),
                    download: this.$gettext("Download"),
                },
            };
        },
        methods: {
            clearClipboard() {
                this.$clipboard.clear();
                this.expanded = false;
            },
            batchPrivate() {
                Api.post("batch/photos/private", {"photos": this.selection}).then(this.onPrivateToggled.bind(this));
            },
            onPrivateToggled() {
                Notify.success(this.$gettext("Toggled private flag"));
                this.clearClipboard();
                this.refresh();
            },
            batchStory() {
                Api.post("batch/photos/story", {"photos": this.selection}).then(this.onStoryToggled.bind(this));
            },
            onStoryToggled() {
                Notify.success(this.$gettext("Toggled story flag"));
                this.clearClipboard();
                this.refresh();
            },
            batchDeletePhotos() {
                this.dialog.delete = false;

                Api.post("batch/photos/delete", {"photos": this.selection}).then(this.onDeleted.bind(this));
            },
            onDeleted() {
                Notify.success(this.$gettext("Photos deleted"));
                this.clearClipboard();
                this.refresh();
            },
            batchTag() {
                Notify.warning(this.$gettext("Not implemented yet"));
                this.expanded = false;
            },
            addToAlbum(albumUUID) {
                this.dialog.album = false;

                Api.post(`albums/${albumUUID}/photos`, {"photos": this.selection}).then(this.onAdded.bind(this));
            },
            onAdded() {
                this.clearClipboard();
                this.refresh();
            },
            removeFromAlbum() {
                if(!this.album) {
                    this.$notify.error(this.$gettext("remove failed: unknown album"));
                    return
                }

                const albumUUID = this.album.AlbumUUID;

                this.dialog.album = false;

                Api.delete(`albums/${albumUUID}/photos`, {"data": {"photos": this.selection}}).then(this.onRemoved.bind(this));
            },
            onRemoved() {
                this.clearClipboard();
                this.refresh();
            },
            batchEditPhotos() {
                this.dialog.edit = false;
                Notify.warning(this.$gettext("Not implemented yet"));
                this.expanded = false;
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
            openDocs() {
                window.open('https://docs.photoprism.org/en/latest/', '_blank');
            },
        }
    };
</script>
