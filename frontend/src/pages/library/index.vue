<template>
    <div class="p-tab p-tab-index">
        <v-form ref="form" class="p-photo-index" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="fileName">Indexed {{ fileName}}...</span>
                    <span v-else-if="busy">Re-indexing existing files and photos...</span>
                    <span v-else-if="completed">Done.</span>
                    <span v-else>Press button to re-index existing files and photos...</span>
                </p>

                <v-progress-linear color="blue-grey" :value="completed" :indeterminate="busy"></v-progress-linear>

                <v-btn
                        :disabled="busy"
                        color="blue-grey"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="startIndexing()"
                >
                    Index
                    <v-icon right dark>update</v-icon>
                </v-btn>
            </v-container>
        </v-form>
    </div>
</template>

<script>
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
            }
        },
        methods: {
            submit() {
                // DO NOTHING
            },
            startIndexing() {
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;
                this.fileName = '';

                const ctx = this;

                this.$api.post('index').then(function () {
                    ctx.busy = false;
                    ctx.completed = 100;
                    this.fileName = '';
                }).catch(function () {
                    this.$alert.error("Indexing failed");
                    ctx.busy = false;
                    ctx.completed = 0;
                    this.fileName = '';
                });
            },
            handleEvent(ev, data) {
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
