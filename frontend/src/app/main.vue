<template>
    <div id="app">
        <app-loading-bar ref="loadingBar" height="4"></app-loading-bar>

        <app-navigation></app-navigation>

        <!-- route outlet -->
        <!-- component matched by the route will render here -->
        <main>
            <router-view></router-view>
        </main>
    </div>
</template>

<script>
    export default {
        name: 'app',
        computed: {},
        methods: {
            login() {
                this.$refs.loginDialog.open();
            },

            logout() {
                this.$session.logout();
            },
        },
    };
</script>

<style lang="css">

</style>
