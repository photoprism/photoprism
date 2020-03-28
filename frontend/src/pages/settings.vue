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
                <translate>General</translate>
            </v-tab>
            <v-tab-item lazy>
                <p-settings-general></p-settings-general>
            </v-tab-item>

            <v-tab id="tab-settings-accounts" ripple @click="changePath('/settings/accounts')">
                <translate>Accounts</translate>
            </v-tab>
            <v-tab-item lazy>
                <p-settings-accounts></p-settings-accounts>
            </v-tab-item>
        </v-tabs>
    </div>
</template>

<script>
    import tabGeneral from "pages/settings/general.vue";
    import tabAccounts from "pages/settings/accounts.vue";

    export default {
        name: 'p-page-settings',
        props: {
            tab: Number
        },
        components: {
            'p-settings-general': tabGeneral,
            'p-settings-accounts': tabAccounts,
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
