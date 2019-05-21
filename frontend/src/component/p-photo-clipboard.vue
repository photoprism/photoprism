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
                @click.stop="$clipboard.clear()"
                :disabled="!selection.length"
        >
            <v-icon>clear</v-icon>
        </v-btn>
    </v-speed-dial>
</template>
<script>
    export default {
        name: 'p-photo-clipboard',
        props: {
            selection: Array,
        },
        data() {
            return {
                expanded: false,
            };
        },
        methods: {
            batchLike() {
                this.$alert.warning("Not implemented yet");
                this.expanded = false;
            },
            batchDelete() {
                this.$alert.warning("Not implemented yet");
                this.expanded = false;
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
