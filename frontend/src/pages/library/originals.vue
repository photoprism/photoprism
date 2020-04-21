<template>
    <div class="p-tab p-tab-index">
        <v-form ref="form" class="p-photo-index" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="fileName">{{ action }} {{ fileName }}...</span>
                    <span v-else-if="busy">Indexing photos and sidecar files...</span>
                    <span v-else-if="completed">Done.</span>
                    <span v-else>Press button to start indexing...</span>
                </p>

                <p class="options">
                    <v-progress-linear color="secondary-dark" :value="completed"
                                       :indeterminate="busy"></v-progress-linear>
                </p>

                <v-layout wrap align-top class="pb-3">
                    <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
                        <v-checkbox
                                :disabled="busy"
                                class="ma-0 pa-0"
                                v-model="options.raw"
                                color="secondary-dark"
                                :label="labels.raw"
                                hint="RAWs need to be converted to JPEG so that they can be displayed in a browser. You can also do this manually."
                                prepend-icon="photo_camera"
                                persistent-hint
                        >
                        </v-checkbox>
                    </v-flex>

                    <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
                        <v-checkbox
                                :disabled="busy"
                                class="ma-0 pa-0"
                                v-model="options.thumbs"
                                color="secondary-dark"
                                :label="labels.thumbs"
                                hint="Enable to pre-render thumbnails if not done already. On-demand rendering saves storage but requires a powerful CPU."
                                prepend-icon="photo_size_select_large"
                                persistent-hint
                        >
                        </v-checkbox>
                    </v-flex>

                    <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
                        <v-checkbox
                                :disabled="busy"
                                class="ma-0 pa-0"
                                v-model="options.rescan"
                                color="secondary-dark"
                                :label="labels.rescan"
                                hint="Re-index all originals, including already indexed and unchanged files."
                                prepend-icon="cached"
                                persistent-hint
                        >
                        </v-checkbox>
                    </v-flex>
                </v-layout>

                <v-btn
                        :disabled="!busy"
                        color="secondary-dark"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="cancelIndexing()"
                >
                    <translate>Cancel</translate>
                </v-btn>

                <v-btn
                        :disabled="busy"
                        color="secondary-dark"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="startIndexing()"
                >
                    <translate>Index</translate>
                    <v-icon right dark>update</v-icon>
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
        name: 'p-tab-index',
        data() {
            let settings = this.$config.settings();

            return {
                readonly: this.$config.getValue("readonly"),
                started: false,
                busy: false,
                completed: 0,
                subscriptionId: "",
                action: "",
                fileName: "",
                source: null,
                options: {
                    rescan: settings.library.rescan,
                    thumbs: settings.library.thumbs,
                    raw: settings.library.raw,
                },
                labels: {
                    rescan: this.$gettext("Complete rescan"),
                    thumbs: this.$gettext("Create thumbnails"),
                    raw: this.$gettext("Convert RAW files"),
                }
            }
        },
        methods: {
            submit() {
                // DO NOTHING
            },
            cancelIndexing() {
                Api.delete('index');
            },
            startIndexing() {
                this.source = Axios.CancelToken.source();
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;
                this.fileName = '';

                const ctx = this;
                Notify.blockUI();

                Api.post('index', this.options, {cancelToken: this.source.token}).then(function () {
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

                    Notify.error(this.$gettext("Indexing failed"));

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
                    case "indexing":
                        this.action = "Indexing";
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
                        break;
                    case "converting":
                        this.action = "Converting";
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
                        break;
                    case "thumbnails":
                        this.action = "Creating thumbnails for";
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
                        break;
                    case 'completed':
                        this.action = "";
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
            this.subscriptionId = Event.subscribe('index', this.handleEvent);
        },
        destroyed() {
            Event.unsubscribe(this.subscriptionId);
        },
    };
</script>
