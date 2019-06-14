<template>
    <div class="p-page p-page-library">
        <v-tabs
                v-model="active"
                flat
                grow
                color="blue-grey lighten-4"
                slider-color="blue-grey darken-1"
                height="64"
        >
            <v-tab id="tab-upload" ripple>
                Upload
            </v-tab>
            <v-tab-item>
                <p-tab-upload></p-tab-upload>
            </v-tab-item>

            <v-tab id="tab-import" ripple>
                Import
            </v-tab>
            <v-tab-item>
                <p-tab-import></p-tab-import>
            </v-tab-item>

            <v-tab id="tab-index" ripple>
                Index
            </v-tab>
            <v-tab-item>
                <p-tab-index></p-tab-index>
            </v-tab-item>
        </v-tabs>
    </div>
</template>

<script>
    import uploadTab from "pages/library/upload.vue";
    import importTab from "pages/library/import.vue";
    import indexTab from "pages/library/index.vue";

    export default {
        name: 'p-page-library',
        components: {
            'p-tab-upload': uploadTab,
            'p-tab-import': importTab,
            'p-tab-index': indexTab,
        },
        data() {
            return {
                active: 0,
            }
        },
        methods: {}
    };
</script>
