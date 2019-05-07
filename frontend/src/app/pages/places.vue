<template>
    <v-container fluid fill-height class="pa-0 map">
        <l-map :zoom="zoom" :center="center">
            <l-tile-layer :url="url" :attribution="attribution"></l-tile-layer>
            <l-marker :lat-lng="marker"></l-marker>
        </l-map>
    </v-container>
</template>

<script>
    export default {
        name: 'places',
        data() {
            return {
                zoom: 13,
                center: L.latLng(52.5259279,13.414496),
                url: 'https://{s}.tile.osm.org/{z}/{x}/{y}.png',
                attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors',
                marker: L.latLng(52.5259279,13.414496),
            }
        },
        methods: {},
    };
</script>
