<template>
    <div class="p-tab p-tab-index">
        <v-form ref="form" class="p-photo-index" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="busy">Re-indexing existing files and photos...</span>
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
    import axios from "axios";
    import Event from "pubsub-js";

    export default {
        name: 'p-tab-index',
        data() {
            return {
                started: false,
                busy: false,
                completed: 0,
            }
        },
        methods: {
            submit() {
                console.log("SUBMIT");
            },
            startIndexing() {
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;

                this.$alert.info("Indexing photos...");

                const ctx = this;

                axios.post('/api/v1/index').then(function () {
                    Event.publish("alert.success", "Indexing complete");
                    ctx.busy = false;
                    ctx.completed = 100;
                }).catch(function () {
                    Event.publish("alert.error", "Indexing failed");
                    ctx.busy = false;
                    ctx.completed = 0;
                });
            },
        }
    };
</script>
