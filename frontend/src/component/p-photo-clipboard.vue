<template>
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
                color="grey darken-2"
                dark
                fab
        >
            <v-icon v-if="selection.length === 0">menu</v-icon>
            <span v-else>{{ selection.length }}</span>
        </v-btn>

        <v-btn
                fab
                dark
                small
                color="deep-purple lighten-2"
                @click.stop="batchLike()"
                :disabled="!selection.length"
        >
            <v-icon>favorite</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="cyan accent-4"
                @click.stop="batchTag()"
        >
            <v-icon>label</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="teal accent-4"
                @click.stop="batchDownload()"
        >
            <v-icon>save</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="yellow accent-4"
                @click.stop="batchAlbum()"
        >
            <v-icon>create_new_folder</v-icon>
        </v-btn>

        <v-btn
                fab
                dark
                small
                color="delete"
                @click.stop="batchDelete()"
                :disabled="!selection.length"
        >
            <v-icon>delete</v-icon>
        </v-btn>
        <v-btn
                fab
                dark
                small
                color="grey"
                @click.stop="clearClipboard()"
                :disabled="!selection.length"
        >
            <v-icon>clear</v-icon>
        </v-btn>
    </v-speed-dial>
</template>
<script>
    import Event from "pubsub-js";
    import axios from "axios";

    export default {
        name: 'p-photo-clipboard',
        props: {
            selection: Array,
            refresh: Function,
        },
        data() {
            return {
                expanded: false,
            };
        },
        methods: {
            clearClipboard() {
                this.$clipboard.clear();
                this.expanded = false;
            },
            batchLike() {
                Event.publish("ajax.start");

                const ctx = this;

                axios.post("/api/v1/batch/photos/like", {"ids": this.selection}).then(function () {
                    Event.publish("ajax.end");
                    Event.publish("alert.success", "Photos liked");
                    ctx.clearClipboard();
                    ctx.refresh();
                }).catch(() => {
                    Event.publish("ajax.end");
                });
            },
            batchDelete() {
                Event.publish("ajax.start");

                const ctx = this;

                axios.post("/api/v1/batch/photos/delete", {"ids": this.selection}).then(function () {
                    Event.publish("ajax.end");
                    Event.publish("alert.success", "Photos deleted");
                    ctx.clearClipboard();
                    ctx.refresh();
                }).catch(() => {
                    Event.publish("ajax.end");
                });
            },
            batchTag() {
                this.$alert.warning("Not implemented yet");
                this.expanded = false;
            },
            batchAlbum() {
                this.$alert.warning("Not implemented yet");
                this.expanded = false;
            },
            batchDownload() {
                this.$alert.warning("Not implemented yet");
                this.expanded = false;
            },
        }
    };
</script>
