<template>
    <div class="p-tab p-tab-index">
        <v-form ref="form" class="p-photo-index" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="fileName"><translate>Indexing</translate> {{ fileName }}...</span>
                    <span v-else-if="busy"><translate>Indexing photos and sidecar files...</translate></span>
                    <span v-else-if="completed"><translate>Done.</translate></span>
                    <span v-else><translate>Press button to start indexing...</translate></span>
                </p>

                <p class="options">
                    <v-progress-linear color="secondary-dark" :value="completed" :indeterminate="busy"></v-progress-linear>
                </p>

                <v-checkbox
                        v-model="options.skip"
                        color="secondary-dark"
                        :disabled="busy"
                        :label="labels.skip"
                ></v-checkbox>

                <v-btn
                        :disabled="busy"
                        color="secondary-dark"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="startIndexing()"
                >
                    <translate>Start</translate>
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
                started: false,
                busy: false,
                completed: 0,
                subscriptionId: '',
                fileName: '',
                source: null,
                options: {
                    skip: true
                },
                labels: {
                    skip: this.$gettext("Skip existing photos and sidecar files"),
                }
            }
        },
        methods: {
            submit() {
                // DO NOTHING
            },
            startIndexing() {
                this.source = Axios.CancelToken.source();
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;
                this.fileName = '';

                const ctx = this;
                Notify.blockUI();

                Api.post('index', this.options, { cancelToken: this.source.token }).then(function () {
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
                if(this.source) {
                    this.source.cancel('run in background');
                    this.source = null;
                    Notify.unblockUI();
                }

                const type = ev.split('.')[1];

                switch (type) {
                    case 'file':
                        this.busy = true;
                        this.completed = 0;
                        this.fileName = data.fileName;
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
            this.subscriptionId = Event.subscribe('index', this.handleEvent);
        },
        destroyed() {
            Event.unsubscribe(this.subscriptionId);
        },
    };
</script>
