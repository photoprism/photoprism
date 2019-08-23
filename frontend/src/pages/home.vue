<template>
    <div id="photoprism">
        <v-app>
            <p-navigation></p-navigation>
            <v-content>
                <router-view></router-view>
            </v-content>
        </v-app>
    </div>
</template>

<script>
    export default {
        name: 'home',
    };
</script>
