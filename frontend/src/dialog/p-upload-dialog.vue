<template>
    <v-dialog fullscreen hide-overlay scrollable lazy
              v-model="show" persistent class="p-upload-dialog" @keydown.esc="cancel">
        <v-card color="application">
            <v-toolbar dark color="navigation">
                <v-btn icon dark @click.stop="cancel">
                    <v-icon>close</v-icon>
                </v-btn>
                <v-toolbar-title>
                    <translate>Upload</translate>
                </v-toolbar-title>
            </v-toolbar>
            <v-container grid-list-xs text-xs-left fluid>
                <v-form ref="form" class="p-photo-upload" lazy-validation @submit.prevent="submit" dense>
                    <input type="file" ref="upload" multiple @change.stop="upload()" class="d-none">

                    <v-container fluid>
                        <p class="subheading">
                            <v-combobox v-if="total === 0" flat solo hide-details chips deletable-chips
                                        multiple color="secondary-dark" class="my-0"
                                        v-model="selectedAlbums"
                                        :items="albums"
                                        item-text="Title"
                                        item-value="UID"
                                        :allow-overflow="false"
                                        label="Select albums or create a new one"
                                        return-object
                            >
                                <template v-slot:no-data>
                                    <v-list-tile>
                                        <v-list-tile-content>
                                            <v-list-tile-title>
                                                Press <kbd>enter</kbd> to create a new album.
                                            </v-list-tile-title>
                                        </v-list-tile-content>
                                    </v-list-tile>
                                </template>
                                <template v-slot:selection="data">
                                    <v-chip
                                            :key="JSON.stringify(data.item)"
                                            :selected="data.selected"
                                            :disabled="data.disabled"
                                            class="v-chip--select-multi"
                                            @input="data.parent.selectItem(data.item)"
                                    >
                                        <v-icon class="pr-1">folder</v-icon>
                                        {{ data.item.Title ? data.item.Title : data.item | truncate(40) }}
                                    </v-chip>
                                </template>
                            </v-combobox>
                            <span v-else-if="failed">Upload failed</span>
                            <span v-else-if="total > 0 && completed < 100">
                        Uploading {{current}} of {{total}}...
                    </span>
                            <span v-else-if="indexing">Upload complete. Indexing...</span>
                            <span v-else-if="completed === 100">Done.</span>
                        </p>


                        <v-progress-linear color="secondary-dark" v-model="completed"
                                           :indeterminate="indexing"></v-progress-linear>


                        <p class="body-1" v-if="safe">
                            Please don't upload photos containing offensive content. Uploads
                            that may contain such images will be rejected automatically.
                        </p>

                        <p class="body-1" v-if="review">
                            Low-quality photos require a review before they appear in search results.
                        </p>

                        <v-btn
                                :disabled="busy"
                                color="secondary-dark"
                                class="white--text ml-0 mt-2 action-upload"
                                depressed
                                @click.stop="uploadDialog()"
                        >
                            <translate>Upload</translate>
                            <v-icon right dark>cloud_upload</v-icon>
                        </v-btn>
                    </v-container>
                </v-form>

            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import Api from "common/api";
    import Notify from "common/notify";
    import Album from "../model/album";

    export default {
        name: 'p-tab-upload',
        props: {
            show: Boolean,
        },
        data() {
            return {
                albums: [],
                selectedAlbums: [],
                selected: [],
                loading: false,
                uploads: [],
                busy: false,
                indexing: false,
                failed: false,
                current: 0,
                total: 0,
                completed: 0,
                started: 0,
                review: this.$config.feature("review"),
                safe: !this.$config.get("uploadNSFW"),
            }
        },
        methods: {
            findAlbums(q) {
                if (this.loading) {
                    return;
                }

                this.loading = true;

                const params = {
                    q: q,
                    count: 1000,
                    offset: 0,
                    type: "album"
                };

                Album.search(params).then(response => {
                    this.loading = false;
                    this.albums = response.models;
                }).catch(() => this.loading = false);
            },
            cancel() {
                if (this.busy) {
                    Notify.info(this.$gettext("Uploading photos..."));
                    return;
                }

                this.$emit('cancel');
            },
            confirm() {
                if (this.busy) {
                    Notify.info(this.$gettext("Uploading photos..."));
                    return;
                }

                this.$emit('confirm');
            },
            submit() {
                // DO NOTHING
            },
            uploadDialog() {
                this.$refs.upload.click();
            },
            upload() {
                this.started = Date.now();
                this.selected = this.$refs.upload.files;
                this.busy = true;
                this.indexing = false;
                this.failed = false;
                this.total = this.selected.length;
                this.current = 0;
                this.completed = 0;
                this.uploads = [];

                if (!this.total) {
                    return
                }

                Notify.info(this.$gettext("Uploading photos..."));

                let addToAlbums = [];

                if (this.selectedAlbums && this.selectedAlbums.length > 0) {
                    this.selectedAlbums.forEach((a) => {
                        if (typeof a === "string") {
                            addToAlbums.push(a)
                        } else if (a instanceof Album && a.UID) {
                            addToAlbums.push(a.UID)
                        }
                    });
                }

                async function performUpload(ctx) {
                    for (let i = 0; i < ctx.selected.length; i++) {
                        let file = ctx.selected[i];
                        let formData = new FormData();

                        ctx.current = i + 1;

                        formData.append('files', file);

                        await Api.post('upload/' + ctx.started,
                            formData,
                            {
                                headers: {
                                    'Content-Type': 'multipart/form-data'
                                }
                            }
                        ).then(() => {
                            ctx.completed = Math.round((ctx.current / ctx.total) * 100);
                        }).catch(() => {
                            ctx.busy = false;
                            ctx.indexing = false;
                            ctx.completed = 100;
                            ctx.failed = true;

                            Notify.error(ctx.$gettext("Upload failed"));
                        });
                    }
                }

                performUpload(this).then(() => {
                    this.indexing = true;
                    const ctx = this;

                    Api.post('import/upload/' + this.started, {
                        move: true,
                        albums: addToAlbums,
                    }).then(() => {
                        Notify.success(ctx.$gettext("Upload complete"));
                        ctx.busy = false;
                        ctx.indexing = false;
                        ctx.$emit('confirm');
                    }).catch(() => {
                        Notify.error(ctx.$gettext("Failure while importing uploaded files"));
                        ctx.busy = false;
                        ctx.indexing = false;
                    });
                });
            },
        },
        watch: {
            show: function () {
                this.selected = [];
                this.uploads = [];
                this.busy = false;
                this.indexing = false;
                this.failed = false;
                this.current = 0;
                this.total = 0;
                this.completed = 0;
                this.started = 0;
                this.review = this.$config.feature("review");
                this.safe = !this.$config.get("uploadNSFW");
                this.findAlbums("");
            }
        },
    };
</script>
