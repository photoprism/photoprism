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
            <v-tab id="tab-originals" ripple @click="changePath('/library')">
                <translate>Originals</translate>
            </v-tab>
            <v-tab-item>
                <p-tab-originals></p-tab-originals>
            </v-tab-item>

            <v-tab id="tab-import" :disabled="readonly" ripple @click="changePath('/library/import')">
                <translate>Import</translate>
            </v-tab>
            <v-tab-item :disabled="readonly">
                <p-tab-import></p-tab-import>
            </v-tab-item>

            <!-- v-tab id="tab-upload" :disabled="readonly" ripple @click="changePath('/library/upload')">
                <translate>Upload</translate>
            </v-tab>
            <v-tab-item :disabled="readonly">
                <p-tab-upload></p-tab-upload>
            </v-tab-item -->

            <v-tab id="tab-logs" ripple @click="changePath('/library/logs')">
                <translate>Logs</translate>
            </v-tab>
            <v-tab-item>
                <p-tab-logs></p-tab-logs>
            </v-tab-item>
        </v-tabs>
    </div>
</template>

<script>
    // import uploadTab from "pages/library/upload.vue";
    import importTab from "pages/library/import.vue";
    import originalsTab from "pages/library/originals.vue";
    import tabLogs from "pages/library/logs.vue";

    export default {
        name: 'p-page-library',
        props: {
            tab: Number
        },
        components: {
            'p-tab-originals': originalsTab,
            'p-tab-import': importTab,
            // 'p-tab-upload': uploadTab,
            'p-tab-logs': tabLogs,
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
