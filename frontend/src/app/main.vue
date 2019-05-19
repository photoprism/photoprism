<template>
    <div id="app">
        <p-loading-bar ref="loadingBar" height="4"></p-loading-bar>

        <p-alert></p-alert>

        <v-app>
            <p-navigation></p-navigation>

            <v-content>
                <router-view></router-view>
            </v-content>
        </v-app>

        <p-photo-viewer></p-photo-viewer>
    </div>
</template>

<script>
    export default {
        name: 'app',
        computed: {},
        methods: {
            login() {
                // this.$refs.loginDialog.open();
            },

            logout() {
                this.$session.logout();
            },
        },
    };
</script>
