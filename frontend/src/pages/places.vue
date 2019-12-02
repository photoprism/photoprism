<template>
    <v-container fluid fill-height class="pa-0 p-page p-page-places">
        <l-map :zoom="zoom" :center="center" :bounds="bounds" :options="options">
            <l-control position="bottomright">
                <v-toolbar dense floating color="accent lighten-4" v-on:dblclick.stop v-on:click.stop>
                    <v-btn icon v-on:click="currentPosition()">
                        <v-icon>my_location</v-icon>
                    </v-btn>
                    <v-spacer></v-spacer>
                    <v-text-field class="pt-3 pr-3"
                                  single-line
                                  label="Search"
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
            <l-marker v-for="photo in photos" v-bind:data="photo"
                      v-bind:key="photo.index" :lat-lng="photo.location" :icon="photo.icon"
                      :options="photo.options" @click="openPhoto(photo.index)"></l-marker>
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
            const query = this.$route.query;
            const q = query['q'] ? query['q'] : '';
            const lat = query['lat'] ? query['lat'] : '';
            const long = query['long'] ? query['long'] : '';
            const dist = query['dist'] ? query['dist'] : 20;

            return {
                zoom: 15,
                position: null,
                center: L.latLng(0, 0),
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
                    lat: lat,
                    long: long,
                    dist: dist,
                },
                offset: 0,
                pageSize: 101,
                lastQuery: {},
                bounds: null,
                minLat: null,
                maxLat: null,
                minLong: null,
                maxLong: null,
            }
        },
        methods: {
            openPhoto(index) {
                this.$viewer.show(this.results, index)
            },
            currentPositionSuccess(position) {
                this.query.lat = position.coords.latitude;
                this.query.long = position.coords.longitude;
                this.query.q = "";
                this.search();
            },
            currentPositionError(error) {
                this.$notify.warning(error.message);
            },
            currentPosition() {
                if ("geolocation" in navigator) {
                    this.$notify.success('Finding your position...');
                    navigator.geolocation.getCurrentPosition(this.currentPositionSuccess, this.currentPositionError);
                } else {
                    this.$notify.warning('Geolocation is not available');
                }
            },
            formChange() {
                this.query.lat = "";
                this.query.long = "";
                this.search();
            },
            clearQuery() {
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
                const photos = [];

                this.resetBoundingBox();

                for (let i = 0, len = results.length; i < len; i++) {
                    let result = results[i];

                    if (!result.hasLocation()) continue;

                    this.fitBoundingBox(result.PhotoLat, result.PhotoLong);

                    photos.push({
                        id: result.getId(),
                        index: i,
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

                if (photos.length === 0) {
                    this.$notify.warning('No locations found');
                    return;
                }

                this.results = results;
                this.photos = photos;

                this.$nextTick(() => {
                    this.center = photos[0].location;
                    this.bounds = [[this.maxLat, this.minLong], [this.minLat, this.maxLong]];
                });

                if (photos.length > 100) {
                    this.$notify.info('More than 100 photos found');
                } else {
                    this.$notify.info(photos.length + ' photos found');
                }
            },
            updateQuery() {
                this.$router.replace({query: this.query}).catch(err => {});

                if(this.query.lat && this.query.long) {
                    this.position = L.latLng(this.query.lat, this.query.long);
                    this.center = L.latLng(this.query.lat, this.query.long);
                } else {
                    this.position = null;
                }
            },
            search() {
                // Don't query the same data more than once
                if (JSON.stringify(this.lastQuery) === JSON.stringify(this.query)) return;

                Object.assign(this.lastQuery, this.query);

                this.offset = 0;

                this.updateQuery();

                this.$router.replace({query: this.query}).catch(err => {});

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                    location: 1,
                };

                Object.assign(params, this.query);

                Photo.search(params).then(response => {
                    if (!response.models.length) {
                        this.$notify.warning('No photos found');
                        return;
                    }

                    this.updateMap(response.models);
                });
            },
        },
        created() {
            this.search();
        },
    };
</script>

