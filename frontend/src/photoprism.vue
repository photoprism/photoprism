<template>
    <div id="photoprism">
        <p-loading-bar height="4"></p-loading-bar>

        <p-notify></p-notify>

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
    import "./css/app.css";

    export default {
        name: 'photoprism',
        computed: {},
        methods: {},
    };
</script>
