<template>
    <div class="p-page p-page-settings">
        <v-tabs
                v-model="active"
                flat
                grow
                color="secondary"
                slider-color="secondary-dark"
                height="64"
        >
            <v-tab id="tab-settings-general" ripple @click="changePath('/settings')">
                General
            </v-tab>
            <v-tab-item>
                <p-tab-general></p-tab-general>
            </v-tab-item>

            <v-tab id="tab-settings-logs" ripple @click="changePath('/settings/logs')">
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
        props: {
            tab: Number
        },
        components: {
            'p-tab-general': tabGeneral,
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
        },
    };
</script>
