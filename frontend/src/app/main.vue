<template>
    <div id="app">
        <app-loading-bar ref="loadingBar" height="4"></app-loading-bar>

        <app-alert></app-alert>

        <v-app>
            <app-navigation></app-navigation>

            <v-content>
                <router-view></router-view>
            </v-content>
        </v-app>

        <app-gallery></app-gallery>
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
