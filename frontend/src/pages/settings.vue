<template>
    <div class="p-page p-page-settings">
        <v-tabs
                v-model="active"
                flat
                grow
                color="blue-grey lighten-4"
                slider-color="blue-grey darken-1"
                height="64"
        >
            <v-tab id="tab-settings-general">
                General
            </v-tab>
            <v-tab-item>
                <p-tab-general></p-tab-general>
            </v-tab-item>

            <v-tab id="tab-settings-logs">
                Logs
            </v-tab>
            <v-tab-item>
                <p-tab-logs></p-tab-logs>
            </v-tab-item>
        </v-tabs>
    </div>
</template>

<script>
    import tabGeneral from "pages/settings/general.vue";
    import tabLogs from "pages/settings/logs.vue";

    export default {
        name: 'p-page-settings',
        components: {
            'p-tab-general': tabGeneral,
            'p-tab-logs': tabLogs,
        },
        data() {
            return {
                readonly: this.$config.getValue("readonly"),
                active: 0,
            }
        },
        methods: {
        },
    };
</script>
