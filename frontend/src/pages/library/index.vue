<template>
    <div class="p-tab p-tab-index">
        <v-form ref="form" class="p-photo-index" lazy-validation @submit.prevent="submit" dense>
            <input type="file" ref="upload" multiple @change.stop="upload()" class="d-none">

            <v-container fluid>
                <h3 class="subheading">Only available using the command-line interface at the moment</h3>
                <p class="body-1 mt-2">
                    Issues labeled <a href="https://github.com/photoprism/photoprism/labels/help%20wanted">help
                    wanted</a> /
                    <a href="https://github.com/photoprism/photoprism/labels/easy">easy</a> can be good (first)
                    contributions.
                    Our <a href="https://github.com/photoprism/photoprism/wiki">Developer Guide</a> contains all
                    information
                    necessary to get you started.
                </p>
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
            }
        },
        methods: {
        }
    };
</script>
