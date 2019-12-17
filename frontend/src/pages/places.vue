<template>
    <v-container fluid fill-height class="pa-0 p-page p-page-places">
        <l-map :zoom="zoom" :center="center" :bounds="bounds" :options="options"
               @update:zoom="onZoom"
               @update:center="onCenter">

            <l-control position="bottomright">
                <!-- v-container class="pb-0 pt-0 pl-3 pr-3 mb-0 mt-0" v-if="loading">
                    <v-progress-linear :indeterminate="true" color="light-blue lighten-1"></v-progress-linear>
                </v-container -->
                <v-toolbar dense floating color="accent lighten-4 mt-0" v-on:dblclick.stop v-on:click.stop>
                    <v-btn icon v-on:click="currentPosition()">
                        <v-icon>my_location</v-icon>
                    </v-btn>
                    <v-spacer></v-spacer>
                    <v-text-field class="pt-3 pr-3"
                                  single-line
                                  :label="labels.search"
                                  prepend-inner-icon="search"
                                  clearable
                                  color="secondary-dark"
                                  @click:clear="clearQuery"
                                  v-model="query.q"
                                  @keyup.enter.native="formChange"
                    ></v-text-field>
                </v-toolbar>
            </l-control>
            <l-tile-layer :url="url" :attribution="attribution"></l-tile-layer>
            <l-marker v-for="(photo, index) in photos" v-bind:data="photo"
                      v-bind:key="index" :lat-lng="photo.location" :icon="photo.icon"
                      :options="photo.options" @click="openPhoto(index)"></l-marker>
            <l-marker v-if="position" :lat-lng="position" :z-index-offset="100"></l-marker>
        </l-map>
    </v-container>
</template>

<script>
    import * as L from "leaflet";
    import Photo from "model/photo";

    export default {
        name: 'p-page-places',
        data() {
            const pos = this.startPos();
            const query = this.$route.query;
            const q = query['q'] ? query['q'] : "";
            const zoom = query['zoom'] ? parseInt(query['zoom']) : 15;
            const dist = this.getDistance(zoom);

            return {
                loading: false,
                zoom: zoom,
                position: null,
                center: L.latLng(parseFloat(pos.lat), parseFloat(pos.long)),
                url: 'https://{s}.tile.osm.org/{z}/{x}/{y}.png',
                attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors',
                options: {
                    icon: {
                        iconSize: [50, 50]
                    },
                    minZoom: 3,
                },
                photos: [],
                results: [],
                query: {
                    q: q,
                    lat: pos.lat,
                    long: pos.long,
                    dist: dist.toString(),
                    zoom: zoom.toString(),
                },
                offset: 0,
                pageSize: 101,
                lastQuery: {},
                bounds: null,
                minLat: null,
                maxLat: null,
                minLong: null,
                maxLong: null,
                labels: {
                    search: this.$gettext("Search"),
                },
                config: this.$config.values,
            }
        },
        methods: {
            getDistance(zoom) {
                switch (zoom) {
                    case 18:
                        return 1;
                    case 17:
                        return 3;
                    case 16:
                        return 5;
                    case 15:
                        return 6;
                    case 14:
                        return 10;
                    case 13:
                        return 15;
                    case 12:
                        return 30;
                    case 11:
                        return 60;
                    case 10:
                        return 100;
                    case 9:
                        return 300;
                    case 8:
                        return 400;
                    case 7:
                        return 800;
                    case 6:
                        return 1600;
                    case 5:
                        return 2000;
                }

                return 2500;
            },
            onZoom(zoom) {
                if(this.query.zoom === zoom.toString()) return;

                this.query.zoom = zoom.toString();
                this.query.dist = this.getDistance(zoom).toString();

                this.search();
            },
            onCenter(pos) {
                const changed = Math.abs(this.query.lat - pos.lat) > 0.001 ||
                    Math.abs(this.query.long - pos.lng) > 0.001;

                this.query.lat = pos.lat.toString();
                this.query.long = pos.lng.toString();

                if(!changed) return;

                this.search();
            },
            startPos() {
                const pos = this.$config.getValue("pos");
                const query = this.$route.query;

                let result = {
                    lat: pos.lat.toString(),
                    long: pos.long.toString(),
                };

                const queryLat = query['lat'];
                const queryLong = query['long'];

                let storedLat = window.localStorage.getItem("lat");
                let storedLong = window.localStorage.getItem("long");

                if (queryLat && queryLong) {
                    result.lat = queryLat;
                    result.long = queryLong;
                } else if (storedLat && storedLong) {
                    result.lat = storedLat;
                    result.long = storedLong;
                }

                return result;
            },
            openPhoto(index) {
                this.$viewer.show(this.results, index)
            },
            onPosition(position) {
                this.position = L.latLng(position.coords.latitude, position.coords.longitude);
                this.center = L.latLng(position.coords.latitude, position.coords.longitude);
                this.query.q = "";
            },
            onPositionError(error) {
                this.$notify.warning(error.message);
            },
            currentPosition() {
                if ("geolocation" in navigator) {
                    this.$notify.success(this.$gettext('Finding your position...'));
                    navigator.geolocation.getCurrentPosition(this.onPosition.bind(this), this.onPositionError.bind(this));
                } else {
                    this.$notify.warning(this.$gettext('Geolocation is not available'));
                }
            },
            formChange() {
                this.query.lat = "";
                this.query.long = "";
                this.search();
            },
            clearQuery() {
                this.position = null;
                this.query.q = "";
                this.query.lat = "";
                this.query.long = "";
                this.search();
            },
            resetBoundingBox() {
                this.minLat = null;
                this.maxLat = null;
                this.minLong = null;
                this.maxLong = null;
            },
            fitBoundingBox(lat, long) {
                if (this.maxLat === null || lat > this.maxLat) {
                    this.maxLat = lat;
                }

                if (this.minLat === null || lat < this.minLat) {
                    this.minLat = lat;
                }

                if (this.maxLong === null || long > this.maxLong) {
                    this.maxLong = long;
                }

                if (this.minLong === null || long < this.minLong) {
                    this.minLong = long;
                }
            },
            updateMap(results) {
                for (let i = 0, len = results.length; i < len; i++) {
                    let result = results[i];

                    if (!result.hasLocation()) continue;

                    let index = this.results.findIndex((p) => p.PhotoUUID === result.PhotoUUID);

                    if (index !== -1) continue;

                    this.results.push(result);
                    this.photos.push({
                        id: result.getId(),
                        options: {
                            title: result.getTitle(),
                            clickable: true,
                        },
                        icon: L.icon({
                            iconUrl: result.getThumbnailUrl('tile_50'),
                            iconRetinaUrl: result.getThumbnailUrl('tile_100'),
                            iconSize: [50, 50],
                            className: 'leaflet-marker-photo',
                        }),
                        location: L.latLng(result.PhotoLat, result.PhotoLong),
                    });
                }

                if (this.photos.length === 0) {
                    this.$notify.warning(this.$gettext('Nothing to see here'));
                    return;
                }

                this.$nextTick(() => {
                    if(!this.query.q) return;
                    this.center = this.photos[this.photos.length - 1].location;
                    this.position = this.photos[this.photos.length - 1].location;
                });
            },
            updateQuery() {
                const query = Object(this.query);

                if (this.query.lat && this.query.long) {
                    window.localStorage.setItem("lat", this.query.lat.toString());
                    window.localStorage.setItem("long", this.query.long.toString());
                } else {
                    this.position = null;
                }

                if (JSON.stringify(this.$route.query) !== JSON.stringify(query)) {
                    this.$router.replace({query: query});
                }
            },
            search() {
                if (this.loading) return;

                // Don't query the same data more than once
                if (JSON.stringify(this.lastQuery) === JSON.stringify(this.query)) return;

                this.offset = 0;
                this.loading = true;

                Object.assign(this.lastQuery, this.query);

                this.updateQuery();

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                    location: 1,
                };

                Object.assign(params, this.query);

                Photo.search(params).then(response => {
                    this.loading = false;

                    if (!response.models.length) {
                        return;
                    }

                    this.updateMap(response.models);
                }).catch(() => this.loading = false);
            },
        },
        created() {
            this.search();
        },
    };
</script>

