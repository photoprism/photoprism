<template>
    <div id="photoprism">
        <p-loading-bar height="4"></p-loading-bar>

        <p-alert></p-alert>

        <!-- the router view at the highest level will be -->
        <!-- one of either <login> or <main> components -->
        <router-view></router-view>

        <p-photo-viewer></p-photo-viewer>
    </div>
</template>

<script>
    import "./css/app.css";

    export default {
        name: 'photoprism',
    };
</script>
