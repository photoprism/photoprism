<template>
    <div>
    <v-container fluid class="pa-0">
    <v-speed-dial
            fixed
            bottom
            right
            direction="top"
            v-model="expanded"
            transition="slide-y-reverse-transition"
            class="p-photo-clipboard"
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
                color="deep-purple lighten-2"
                @click.stop="batchPrivate()"
                v-if="selection.length"
                class="p-photo-clipboard-private"
        >
            <v-icon>vpn_key</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="cyan accent-4"
                v-if="selection.length"
                @click.stop="batchStory()"
                class="p-photo-clipboard-story"
        >
            <v-icon>wifi</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="light-blue accent-4"
                v-if="!selection.length"
                @click.stop="openDocs()"
                class="p-photo-clipboard-docs"
        >
            <v-icon>info</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="teal accent-4"
                @click.stop="batchDownload()"
                class="p-photo-clipboard-download"
        >
            <v-icon>save</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="yellow accent-4"
                @click.stop="batchAlbum()"
                class="p-photo-clipboard-album"
        >
            <v-icon>create_new_folder</v-icon>
        </v-btn>

        <v-btn
                fab
                dark
                small
                color="delete"
                @click.stop="dialog.delete = true"
                v-if="selection.length"
                class="p-photo-clipboard-delete"
        >
            <v-icon>delete</v-icon>
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
    <p-photo-delete-dialog :show="dialog.delete" @cancel="dialog.delete = false" @confirm="batchDelete"></p-photo-delete-dialog>
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
        },
        data() {
            return {
                expanded: false,
                dialog: {
                    delete: false
                },
            };
        },
        methods: {
            clearClipboard() {
                this.$clipboard.clear();
                this.expanded = false;
            },
            batchPrivate() {
                const ctx = this;

                Api.post("batch/photos/private", {"ids": this.selection}).then(function () {
                    Notify.success("Toggled private flag");
                    ctx.clearClipboard();
                    ctx.refresh();
                });
            },
            batchStory() {
                const ctx = this;

                Api.post("batch/photos/story", {"ids": this.selection}).then(function () {
                    Notify.success("Toggled story flag");
                    ctx.clearClipboard();
                    ctx.refresh();
                });
            },
            batchDelete() {
                this.dialog.delete = false;

                const ctx = this;

                Api.post("batch/photos/delete", {"ids": this.selection}).then(function () {
                    Notify.success("Photos deleted");
                    ctx.clearClipboard();
                    ctx.refresh();
                });
            },
            batchTag() {
                Notify.warning("Not implemented yet");
                this.expanded = false;
            },
            batchAlbum() {
                Notify.warning("Not implemented yet");
                this.expanded = false;
            },
            batchDownload() {
                Notify.warning("Not implemented yet");
                this.expanded = false;
            },
            openDocs() {
                window.open('https://docs.photoprism.org/en/latest/', '_blank');
            },
        }
    };
</script>
