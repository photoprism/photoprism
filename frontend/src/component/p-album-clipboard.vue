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
                        :disabled="selection.length !== 1"
                >
                    <v-icon>cloud_download</v-icon>
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
        },
        data() {
            return {
                expanded: false,
                dialog: {
                    delete: false,
                    edit: false,
                },
                labels: {
                    download: this.$gettext("Download"),
                    delete: this.$gettext("Delete"),
                },

            };
        },
        methods: {
            clearClipboard() {
                this.selection.splice(0, this.selection.length);
                this.expanded = false;
            },
            batchDelete() {
                this.dialog.delete = false;

                Api.post("batch/albums/delete", {"albums": this.selection}).then(this.onDeleted.bind(this));
            },
            onDeleted() {
                Notify.success(this.$gettext("Albums deleted"));
                this.clearClipboard();
                this.refresh();
            },
            download() {
                if(this.selection.length !== 1) {
                    Notify.error(this.$gettext("You can only download one album"));
                    return;
                }

                this.onDownload(`/api/v1/albums/${this.selection[0]}/download`);

                this.expanded = false;
            },
            onDownload(path) {
                Notify.success(this.$gettext("Downloading..."));
                window.open(path, "_blank");
            },
        }
    };
</script>
