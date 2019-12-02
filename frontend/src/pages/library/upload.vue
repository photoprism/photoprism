<template>
    <div class="p-tab p-tab-upload">
        <v-form ref="form" class="p-photo-upload" lazy-validation @submit.prevent="submit" dense>
            <input type="file" ref="upload" multiple @change.stop="upload()" class="d-none">

            <v-container fluid>
                <p class="subheading">
                    <span v-if="total === 0">Select photos to start upload...</span>
                    <span v-else-if="total > 0 && completed < 100">Uploading {{current}} of {{total}}...</span>
                    <span v-else-if="indexing">Upload complete. Indexing...</span>
                    <span v-else-if="completed === 100">Done.</span>
                </p>

                <v-progress-linear color="secondary-dark" v-model="completed" :indeterminate="indexing"></v-progress-linear>

                <v-btn
                        :disabled="busy"
                        color="secondary-dark"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="uploadDialog()"
                >
                    Upload
                    <v-icon right dark>cloud_upload</v-icon>
                </v-btn>
            </v-container>
        </v-form>
    </div>
</template>

<script>
    import Api from "common/api";
    import Notify from "common/notify";

    export default {
        name: 'p-tab-upload',
        data() {
            return {
                selected: [],
                uploads: [],
                busy: false,
                indexing: false,
                current: 0,
                total: 0,
                completed: 0,
                started: 0,
            }
        },
        methods: {
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
                this.total = this.selected.length;
                this.current = 0;
                this.completed = 0;
                this.uploads = [];

                if (!this.total) {
                    return
                }

                Notify.info("Uploading photos...");
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
                        ).then(function () {
                            ctx.completed = Math.round((ctx.current / ctx.total) * 100);
                        }).catch(function () {
                            Notify.error("Upload failed");
                        });
                    }
                }

                performUpload(this).then(() => {
                    this.indexing = true;
                    const ctx = this;

                    Api.post('import/upload/' + this.started).then(function () {
                        Notify.unblockUI();
                        Notify.success("Upload complete");
                        ctx.busy = false;
                        ctx.indexing = false;
                    }).catch(function () {
                        Notify.unblockUI();
                        Notify.error("Failure while importing uploaded files");
                        ctx.busy = false;
                        ctx.indexing = false;
                    });
                });
            },
        }
    };
</script>
