<template>
    <div class="p-tab p-tab-upload">
        <v-form ref="form" class="p-photo-upload" lazy-validation @submit.prevent="submit" dense>
            <input type="file" ref="upload" multiple @change.stop="upload()" class="d-none">

            <v-container fluid>
                <p class="subheading">
                    <span v-if="total === 0">Select photos to start upload...</span>
                    <span v-else-if="total > 0 && completed < 100">Uploaded {{current}} of {{total}}...</span>
                    <span v-else-if="indexing">Upload complete. Indexing...</span>
                    <span v-else-if="completed === 100">Done.</span>
                </p>

                <v-progress-linear color="blue-grey" v-model="completed" :indeterminate="indexing"></v-progress-linear>

                <v-btn
                        :disabled="busy"
                        color="blue-grey"
                        class="white--text ml-0"
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
    import axios from "axios";
    import Event from "pubsub-js";

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
                console.log("SUBMIT");
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

                this.$alert.info("Uploading photos...");

                Event.publish("ajax.start");

                async function performUpload(ctx) {
                    for (let i = 0; i < ctx.selected.length; i++) {
                        let file = ctx.selected[i];
                        let formData = new FormData();

                        formData.append('files', file);

                        await axios.post('/api/v1/upload/' + ctx.started,
                            formData,
                            {
                                headers: {
                                    'Content-Type': 'multipart/form-data'
                                }
                            }
                        ).then(function () {
                            ctx.current = i + 1;
                            ctx.completed = Math.round((ctx.current / ctx.total) * 100);
                        }).catch(function () {
                            Event.publish("alert.error", "Upload failed");
                        });
                    }
                }

                performUpload(this).then(() => {
                    this.indexing = true;
                    const ctx = this;

                    axios.post('/api/v1/import/upload/' + this.started).then(function () {
                        Event.publish("alert.success", "Import complete");
                        ctx.busy = false;
                        ctx.indexing = false;
                    }).catch(function () {
                        Event.publish("alert.error", "Import failed");
                        ctx.busy = false;
                        ctx.indexing = false;
                    });

                    Event.publish("ajax.end");
                });
            },
        }
    };
</script>
