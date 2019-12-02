<template>
    <div class="p-page p-page-library">
        <v-tabs
                v-model="active"
                flat
                grow
                color="secondary"
                slider-color="secondary-dark"
                height="64"
        >
            <v-tab id="tab-upload" v-if="!readonly" ripple @click="changePath('/library/upload')">
                Upload
            </v-tab>
            <v-tab-item v-if="!readonly">
                <p-tab-upload></p-tab-upload>
            </v-tab-item>

            <v-tab id="tab-import" v-if="!readonly" ripple @click="changePath('/library/import')">
                Import
            </v-tab>
            <v-tab-item v-if="!readonly">
                <p-tab-import></p-tab-import>
            </v-tab-item>

            <v-tab id="tab-index" ripple @click="changePath('/library/index')">
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
        props: {
            tab: Number
        },
        components: {
            'p-tab-upload': uploadTab,
            'p-tab-import': importTab,
            'p-tab-index': indexTab,
        },
        data() {
            return {
                readonly: this.$config.getValue("readonly"),
                active: this.tab,
            }
        },
        methods: {
            changePath: function(path) {
                if (this.$route.path !== path) {
                    this.$router.replace(path)
                }
            }
        }
    };
</script>
