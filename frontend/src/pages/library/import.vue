<template>
    <div class="p-tab p-tab-import">
        <v-form ref="form" class="p-photo-import" lazy-validation @submit.prevent="submit" dense>
            <v-container fluid>
                <p class="subheading">
                    <span v-if="busy">Importing files from directory...</span>
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
    import axios from "axios";
    import Event from "pubsub-js";

    export default {
        name: 'p-tab-import',
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
            startImport() {
                this.started = Date.now();
                this.busy = true;
                this.completed = 0;

                this.$alert.info("Importing photos...");

                const ctx = this;

                axios.post('/api/v1/import').then(function () {
                    Event.publish("alert.success", "Import complete");
                    ctx.busy = false;
                    ctx.completed = 100;
                }).catch(function () {
                    Event.publish("alert.error", "Import failed");
                    ctx.busy = false;
                    ctx.completed = 0;
                });
            },
        }
    };
</script>
