<template>
    <div class="p-tab p-tab-import">
        <v-form ref="form" class="p-photo-import" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="fileName">Importing {{fileName}}...</span>
                    <span v-else-if="busy">Importing files from import folder...</span>
                    <span v-else-if="completed">Done.</span>
                    <span v-else>Press button to import photos...</span>
                </p>

                <p class="options">
                    <v-progress-linear color="secondary-dark" :value="completed"
                                       :indeterminate="busy"></v-progress-linear>
                </p>

                <v-checkbox
                        class="mb-0 mt-4 pa-0"
                        v-model="options.move"
                        color="secondary-dark"
                        :disabled="busy"
                >
                    <template v-slot:label>
                        <div>
                            <v-tooltip right max-width="90%" open-delay="750">
                                <template v-slot:activator="{ on }">
                                    <span v-on="on">{{labels.move}}</span>
                                </template>
                                <span>Instead of creating a copy, move files from import to originals to save storage.
                                    Unsupported file types will never be deleted, they remain in their current folder.</span>
                            </v-tooltip>
                        </div>
                    </template>
                </v-checkbox>

                <p class="subheading">
                    When you import files to your library, they will be sorted by date
                    and given a unique name to avoid duplicates.
                    Importing will automatically convert raw images and render thumbnails as needed.
                    Original file names will be stored and indexed.
                    You can as well manage your originals manually or using other tools.
                </p>

                <v-btn
                        :disabled="!busy"
                        color="secondary-dark"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="cancelImport()"
                >
                    <translate>Cancel</translate>
                </v-btn>

                <v-btn v-if="!$config.values.readonly && $config.feature('upload')"
                       :disabled="busy"
                       color="secondary-dark"
                       class="white--text ml-0 mt-2 hidden-xs-only"
                       depressed
                       @click.stop="showUpload()"
                >
                    <translate>Upload</translate>
                    <v-icon right dark>cloud_upload</v-icon>
                </v-btn>

                <v-btn
                        :disabled="busy"
                        color="secondary-dark"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="startImport()"
                >
                    <translate>Import</translate>
                    <v-icon right dark>create_new_folder</v-icon>
                </v-btn>
            </v-container>
        </v-form>
    </div>
</template>

<script>
    import Api from "common/api";
    import Axios from "axios";
    import Notify from "common/notify";
    import Event from "pubsub-js";

    export default {
        name: 'p-tab-import',
        data() {
            return {
                started: false,
                busy: false,
                completed: 0,
                subscriptionId: '',
                fileName: '',
                source: null,
                options: {
                    move: false,
                },
                labels: {
                    move: this.$gettext("Remove imported files from import folder"),
                }
            }
        },
        methods: {
            showUpload() {
                Event.publish("upload.show");
            },
            submit() {
                // DO NOTHING
            },
            cancelImport() {
                Api.delete('import');
            },
            startImport() {
                this.source = Axios.CancelToken.source();
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;
                this.fileName = '';

                const ctx = this;
                Notify.blockUI();

                Api.post('import', this.options, {cancelToken: this.source.token}).then(function () {
                    Notify.unblockUI();
                    ctx.busy = false;
                    ctx.completed = 100;
                    ctx.fileName = '';
                }).catch(function (e) {
                    Notify.unblockUI();

                    if (Axios.isCancel(e)) {
                        // run in background
                        return
                    }

                    Notify.error(this.$gettext("Import failed"));

                    ctx.busy = false;
                    ctx.completed = 0;
                    ctx.fileName = '';
                });
            },
            handleEvent(ev, data) {
                if (this.source) {
                    this.source.cancel('run in background');
                    this.source = null;
                    Notify.unblockUI();
                }

                const type = ev.split('.')[1];

                switch (type) {
                    case 'file':
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.baseName;
                        break;
                    case 'completed':
                        this.busy = false;
                        this.completed = 100;
                        this.fileName = '';
                        break;
                    default:
                        console.log(data)
                }
            },
        },
        created() {
            this.subscriptionId = Event.subscribe('import', this.handleEvent);
        },
        destroyed() {
            Event.unsubscribe(this.subscriptionId);
        },
    };
</script>
