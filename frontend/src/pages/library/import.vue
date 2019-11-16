<template>
    <div class="p-tab p-tab-import">
        <v-form ref="form" class="p-photo-import" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="fileName">Indexed {{ fileName}}...</span>
                    <span v-else-if="busy">Importing files from directory...</span>
                    <span v-else-if="completed">Done.</span>
                    <span v-else>Press button to import photos from directory...</span>
                </p>

                <v-progress-linear color="blue-grey" :value="completed" :indeterminate="busy"></v-progress-linear>

                <v-btn
                        :disabled="busy"
                        color="blue-grey"
                        class="white--text ml-0 mt-2"
                        depressed
                        @click.stop="startImport()"
                >
                    Import
                    <v-icon right dark>create_new_folder</v-icon>
                </v-btn>
            </v-container>
        </v-form>
    </div>
</template>

<script>
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
            }
        },
        methods: {
            submit() {
                // DO NOTHING
            },
            startImport() {
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;
                this.fileName = '';

                const ctx = this;

                this.$api.post('import').then(function () {
                    ctx.busy = false;
                    ctx.completed = 100;
                    ctx.fileName = '';
                }).catch(function () {
                    Notify.error("Import failed");
                    ctx.busy = false;
                    ctx.completed = 0;
                    ctx.fileName = '';
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
            this.subscriptionId = Event.subscribe('import', this.handleEvent);
        },
        destroyed() {
            Event.unsubscribe(this.subscriptionId);
        },
    };
</script>
