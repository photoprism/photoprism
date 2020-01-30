<template>
    <v-dialog fullscreen hide-overlay scrollable lazy
              v-model="show" persistent class="p-upload-dialog" @keydown.esc="cancel">
        <v-card color="application">
            <v-toolbar dark color="navigation">
                <v-btn icon dark @click.stop="cancel" :disabled="busy">
                    <v-icon>close</v-icon>
                </v-btn>
                <v-toolbar-title><translate>Upload</translate></v-toolbar-title>
            </v-toolbar>
            <v-container grid-list-xs text-xs-left fluid>
                <v-form ref="form" class="p-photo-upload" lazy-validation @submit.prevent="submit" dense>
                    <input type="file" ref="upload" multiple @change.stop="upload()" class="d-none">

                    <v-container fluid>
                        <p class="subheading">
                            <span v-if="total === 0">Select photos to start upload...</span>
                            <span v-else-if="failed">Upload failed</span>
                            <span v-else-if="total > 0 && completed < 100">
                        Uploading {{current}} of {{total}}...
                    </span>
                            <span v-else-if="indexing">Upload complete. Indexing...</span>
                            <span v-else-if="completed === 100">Done.</span>
                        </p>

                        <v-progress-linear color="secondary-dark" v-model="completed"
                                           :indeterminate="indexing"></v-progress-linear>


                        <p class="subheading" v-if="safe">
                            Please don't upload photos containing offensive content. Uploads
                            that may contain such images will be rejected automatically.
                        </p>

                        <v-btn
                                :disabled="busy"
                                color="secondary-dark"
                                class="white--text ml-0 mt-2"
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

    export default {
        name: 'p-tab-upload',
        props: {
            show: Boolean,
        },
        data() {
            return {
                selected: [],
                uploads: [],
                busy: false,
                indexing: false,
                failed: false,
                current: 0,
                total: 0,
                completed: 0,
                started: 0,
                safe: !this.$config.getValue("uploadNSFW")
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
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
                Notify.blockUI();

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
                            Notify.unblockUI();
                            throw Error("upload failed");
                        });
                    }
                }

                performUpload(this).then(() => {
                    this.indexing = true;
                    const ctx = this;

                    Api.post('import/upload/' + this.started).then(() => {
                        Notify.unblockUI();
                        Notify.success(ctx.$gettext("Upload complete"));
                        ctx.busy = false;
                        ctx.indexing = false;
                        ctx.$emit('confirm');
                    }).catch(() => {
                        Notify.unblockUI();
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
            }
        },
    };
</script>
