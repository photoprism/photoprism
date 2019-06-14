<template>
    <div class="p-page p-page-import">
        <v-form ref="form" class="p-photo-import" lazy-validation @submit.prevent="submit" dense>
            <input type="file" ref="upload" multiple @change.stop="upload()" class="d-none">

            <v-toolbar flat color="blue-grey lighten-4">
                <v-toolbar-title>Import</v-toolbar-title>

                <v-spacer></v-spacer>
            </v-toolbar>

            <v-container fluid>
                <p class="subheading">
                    <span v-if="total === 0">Select photos to start import...</span>
                    <span v-else-if="total > 0 && completed < 100">Adding {{current}} of {{total}}...</span>
                    <span v-else-if="completed === 100">Done.</span>
                </p>

                <v-progress-linear color="blue-grey" v-model="completed"></v-progress-linear>

                <v-container grid-list-xs fluid class="pa-0 p-photos p-photo-mosaic">
                    <v-layout row wrap>
                        <v-flex
                                v-for="(file, index) in uploads"
                                :key="index"
                                class="p-photo"
                                xs4 sm3 md2 lg1 d-flex
                        >
                            <v-card tile class="elevation-2 ma-2">
                                <v-img :src="file.data"
                                       aspect-ratio="1"
                                       :title="file.name"
                                       class="grey lighten-2"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate
                                                             color="grey lighten-5"></v-progress-circular>
                                    </v-layout>
                                </v-img>
                            </v-card>
                        </v-flex>
                    </v-layout>
                </v-container>

                <v-btn
                        :disabled="busy"
                        color="blue-grey lighten-2"
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
        name: 'p-page-import',
        data() {
            return {
                selected: [],
                uploads: [],
                busy: false,
                current: 0,
                total: 0,
                completed: 0,
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
                this.selected = this.$refs.upload.files;
                this.busy = true;
                this.total = this.selected.length;
                this.current = 0;
                this.completed = 0;
                this.uploads = [];

                if(!this.total) {
                    return
                }

                this.$alert.info("Uploading photos...");

                Event.publish("ajax.start");

                async function performUpload(ctx) {
                    for (let i = 0; i < ctx.selected.length; i++) {
                        ctx.current = i + 1;
                        ctx.completed = Math.round((ctx.current / ctx.total) * 100);
                        let file = ctx.selected[i];
                        let formData = new FormData();

                        formData.append('files', file);

                        if (file.type.match('image.*')) {
                            const reader = new FileReader;

                            reader.onload = e => {
                                ctx.uploads.push({name: file.name, data: e.target.result});
                            };

                            reader.readAsDataURL(file)
                        }

                        await axios.post('/api/v1/upload',
                            formData,
                            {
                                headers: {
                                    'Content-Type': 'multipart/form-data'
                                }
                            }
                        ).then(function () {
                        }).catch(function () {
                            Event.publish("alert.error", "Upload failed");
                        });
                    }
                }

                performUpload(this).then(() => {
                    Event.publish("ajax.end");
                    Event.publish("alert.success", "Photos uploaded and imported");
                    this.busy = false;
                });
            },
        }
    };
</script>
