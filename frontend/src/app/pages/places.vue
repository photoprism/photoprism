<template>
    <v-container fluid fill-height class="pa-0 map">
        <l-map :zoom="zoom" :center="center" :bounds="bounds" :options="options">
            <l-control position="bottomright">
                <v-toolbar dense floating color="grey lighten-4" v-on:dblclick.stop v-on:click.stop>
                    <v-btn icon v-on:click="currentPosition()">
                        <v-icon>my_location</v-icon>
                    </v-btn>
                    <v-spacer></v-spacer>
                    <v-text-field class="pt-3 pr-3"
                                  single-line
                                  label="Search"
                                  prepend-inner-icon="search"
                                  clearable
                                  color="blue-grey"
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
            <l-marker v-if="position" :lat-lng="position" z-index-offset="1"></l-marker>
        </l-map>
    </v-container>
</template>

<script>
    import * as L from "leaflet";
    import Photo from "model/photo";

    export default {
        name: 'places',
        data() {
            const query = this.$route.query;
            const order = query['order'] ? query['order'] : 'newest';
            const q = query['q'] ? query['q'] : '';

            return {
                zoom: 15,
                position: null,
                center: L.latLng(52.5259279, 13.414496),
                url: 'https://{s}.tile.osm.org/{z}/{x}/{y}.png',
                attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors',
                options: {
                    icon: {
                        iconSize: [40, 40]
                    }
                },
                photos: [],
                results: [],
                query: {
                    order: order,
                    q: q,
                },
                offset: 0,
                pageSize: 100,
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
                this.$gallery.show(this.results, index)
            },
            currentPosition() {
                if ("geolocation" in navigator) {
                    const self = this;
                    this.$alert.success('Finding your position...');
                    navigator.geolocation.getCurrentPosition(function(position) {
                        self.center = L.latLng(position.coords.latitude, position.coords.longitude);
                        self.position = L.latLng(position.coords.latitude, position.coords.longitude);
                    });
                } else {
                    this.$alert.warning('Geolocation is not available');
                }
            },
            formChange() {
                this.refreshList();
            },
            clearQuery() {
                this.query.q = '';
                this.refreshList();
            },
            resetBoundingBox() {
                this.minLat = null;
                this.maxLat = null;
                this.minLong = null;
                this.maxLong = null;
            },
            fitBoundingBox(lat, long) {
                if(this.maxLat === null || lat > this.maxLat) {
                    this.maxLat = lat;
                }

                if(this.minLat === null || lat < this.minLat) {
                    this.minLat = lat;
                }

                if(this.maxLong === null || long > this.maxLong) {
                    this.maxLong = long;
                }

                if(this.minLong === null || long < this.minLong) {
                    this.minLong = long;
                }
            },
            updateMap() {
                const photos = [];

                this.resetBoundingBox();

                for (let i = 0, len = this.results.length; i < len; i++) {
                    let result = this.results[i];

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
                            iconUrl: result.getThumbnailUrl('square', 50),
                            iconRetinaUrl: result.getThumbnailUrl('square', 100),
                            iconSize: [50, 50],
                            className: 'leaflet-marker-photo',
                        }),
                        location: L.latLng(result.PhotoLat, result.PhotoLong),
                    });
                }

                this.center = photos[photos.length - 1].location;

                this.bounds = [[this.maxLat, this.minLong], [this.minLat, this.maxLong]];

                this.$alert.info(photos.length + ' photos found');

                this.photos = photos;
            },

            refreshList() {
                // Don't query the same data more than once
                if (JSON.stringify(this.lastQuery) === JSON.stringify(this.query)) return;

                Object.assign(this.lastQuery, this.query);

                this.offset = 0;

                this.$router.replace({query: this.query});

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.query);

                Photo.search(params).then(response => {
                    if (!response.models.length) {
                        this.$alert.warning('No photos found');
                        return;
                    }

                    this.results = response.models;

                    this.updateMap();
                });
            },
        },
        created() {
            this.refreshList();
        },
    };

    /*  L.icon({
        html: '<div style="background-image: url(/api/v1/thumbnails/square/40/cc1a022c30fff3d5603f1c3f722ec1960e3fa95e);"></div>​',
        className: 'leaflet-marker-photo' }), */
</script>

