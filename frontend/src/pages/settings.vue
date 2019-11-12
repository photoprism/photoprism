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
            <v-tab id="tab-upload">
                General
            </v-tab>
            <v-tab-item>
                <p-tab-general></p-tab-general>
            </v-tab-item>
        </v-tabs>
    </div>
</template>

<script>
    import tabGeneral from "pages/settings/general.vue";

    export default {
        name: 'p-page-settings',
        components: {
            'p-tab-general': tabGeneral,
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
