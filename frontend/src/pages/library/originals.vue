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

                <v-checkbox
                        class="mb-0 mt-4 pa-0"
                        v-model="options.skipUnchanged"
                        color="secondary-dark"
                        :disabled="busy"
                        :label="labels.skipUnchanged"
                ></v-checkbox>
                <v-checkbox
                        class="ma-0 pa-0"
                        v-model="options.convertRaw"
                        color="secondary-dark"
                        :disabled="busy || readonly"
                        :label="labels.convertRaw"
                ></v-checkbox>
                <v-checkbox
                        class="ma-0 pa-0"
                        v-model="options.createThumbs"
                        color="secondary-dark"
                        :disabled="busy"
                        :label="labels.createThumbs"
                ></v-checkbox>
                <!-- v-checkbox
                        class="ma-0 pa-0"
                        v-model="options.groomMetadata"
                        color="secondary-dark"
                        :disabled="busy"
                        :label="labels.groomMetadata"
                ></v-checkbox -->

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
                    skipUnchanged: true,
                    createThumbs: false,
                    convertRaw: false,
                    groomMetadata: false,
                },
                labels: {
                    skipUnchanged: this.$gettext("Skip unchanged files"),
                    createThumbs: this.$gettext("Pre-render thumbnails"),
                    convertRaw: this.$gettext("Convert RAW to JPEG"),
                    groomMetadata: this.$gettext("Groom metadata and estimate locations"),
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
