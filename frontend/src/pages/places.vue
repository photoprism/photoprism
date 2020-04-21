<template>
    <v-container fluid fill-height class="pa-0 p-page p-page-places">
        <div id="map" style="width: 100%; height: 100%;">
            <div class="p-map-control">
                <div class="mapboxgl-ctrl mapboxgl-ctrl-group">
                    <v-text-field class="pa-0 ma-0"
                                  single-line
                                  solo
                                  flat
                                  :label="labels.search"
                                  prepend-inner-icon="search"
                                  clearable
                                  hide-details
                                  browser-autocomplete="off"
                                  color="secondary-dark"
                                  @click:clear="clearQuery"
                                  v-model="filter.q"
                                  @keyup.enter.native="formChange"
                    ></v-text-field>
                </div>
            </div>
        </div>
    </v-container>
</template>

<script>
    import Photo from "model/photo";
    import mapboxgl from "mapbox-gl";
    import Api from "../common/api";
    import Thumb from "../model/thumb";

    export default {
        name: 'p-page-places',
        watch: {
            '$route'() {
                this.filter.q = this.query();
                this.lastFilter = {};
                this.search();
            }
        },
        data() {
            const s = this.$config.values.settings.maps;

            return {
                initialized: false,
                map: null,
                markers: {},
                markersOnScreen: {},
                loading: false,
                url: 'https://api.maptiler.com/maps/' + s.style + '/{z}/{x}/{y}.png?key=xCDwZsNKW3rlveVG0WUU',
                attribution: '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>',
                options: {
                    container: "map",
                    style: "https://api.maptiler.com/maps/" + s.style + "/style.json?key=xCDwZsNKW3rlveVG0WUU",
                    attributionControl: true,
                    customAttribution: this.attribution,
                    zoom: 0,
                },
                photos: [],
                result: {},
                filter: {
                    q: this.query(),
                },
                lastFilter: {},
                labels: {
                    search: this.$gettext("Search"),
                },
                config: this.$config.values,
                settings: s,
            }
        },
        methods: {
            query: function () {
                return this.$route.params.q ? this.$route.params.q : "";
            },
            openPhoto(id) {
                if (!this.photos || !this.photos.length) {
                    this.photos = this.result.features.map((f) => new Photo(f.properties));
                }

                if (this.photos.length > 0) {
                    const index = this.photos.findIndex((p) => p.PhotoUUID === id);

                    this.$viewer.show(Thumb.fromPhotos(this.photos), index)
                } else {
                    this.$notify.warning("No photos found");
                }
            },
            formChange() {
                this.search();
            },
            clearQuery() {
                this.filter.q = "";
                this.search();
            },
            updateQuery() {
                if (this.query() !== this.filter.q) {
                    if (this.filter.q) {
                        this.$router.replace({name: "place", params: {q: this.filter.q}});
                    } else {
                        this.$router.replace({name: "places"});
                    }
                }
            },
            search() {
                if (this.loading) return;
                // Don't query the same data more than once
                if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) return;
                this.loading = true;

                Object.assign(this.lastFilter, this.filter);

                this.updateQuery();

                const options = {
                    params: this.filter,
                };

                return Api.get("geo", options).then((response) => {
                    if (!response.data.features || response.data.features.length === 0) {
                        this.loading = false;

                        this.$notify.warning("No photos found");

                        return;
                    }

                    this.photos = {};
                    this.result = response.data;

                    this.map.getSource("photos").setData(this.result);

                    if (this.filter.q || !this.initialized) {
                        this.map.fitBounds(this.result.bbox, {maxZoom: 17, padding: 100, duration: this.settings.animate, essential: false, animate: this.settings.animate > 0});
                    }

                    this.initialized = true;
                    this.loading = false;

                    this.updateMarkers();
                }).catch(() => this.loading = false);
            },
            renderMap() {
                this.map = new mapboxgl.Map(this.options);

                this.map.addControl(new mapboxgl.NavigationControl({showCompass: false}, 'top-right'));
                this.map.addControl(new mapboxgl.FullscreenControl({container: document.querySelector('body')}));
                this.map.addControl(new mapboxgl.GeolocateControl({
                    positionOptions: {
                        enableHighAccuracy: true
                    },
                    trackUserLocation: true
                }));

                this.map.on("load", () => this.onMapLoad());
            },
            updateMarkers() {
                if (this.loading) return;
                let newMarkers = {};
                let features = this.map.querySourceFeatures("photos");

                for (let i = 0; i < features.length; i++) {
                    let coords = features[i].geometry.coordinates;
                    let props = features[i].properties;
                    if (props.cluster) continue;
                    let id = features[i].id;

                    let marker = this.markers[id];
                    if (!marker) {
                        let el = document.createElement('div');
                        el.className = 'marker';
                        el.title = props.PhotoTitle;
                        el.style.backgroundImage =
                            'url(/api/v1/thumbnails/' +
                            props.FileHash + '/tile_50)';
                        el.style.width = '50px';
                        el.style.height = '50px';

                        el.addEventListener('click', () => this.openPhoto(props.PhotoUUID));
                        marker = this.markers[id] = new mapboxgl.Marker({
                            element: el
                        }).setLngLat(coords);
                    } else {
                        marker.setLngLat(coords);
                    }

                    newMarkers[id] = marker;

                    if (!this.markersOnScreen[id]) {
                        marker.addTo(this.map);
                    }
                }
                for (let id in this.markersOnScreen) {
                    if (!newMarkers[id]) {
                        this.markersOnScreen[id].remove();
                    }
                }
                this.markersOnScreen = newMarkers;
            },
            onMapLoad() {
                this.map.addSource('photos', {
                    type: 'geojson',
                    data: null,
                    cluster: true,
                    clusterMaxZoom: 14, // Max zoom to cluster points on
                    clusterRadius: 50 // Radius of each cluster when clustering points (defaults to 50)
                });

                this.map.addLayer({
                    id: 'clusters',
                    type: 'circle',
                    source: 'photos',
                    filter: ['has', 'point_count'],
                    paint: {
                        'circle-color': [
                            'step',
                            ['get', 'point_count'],
                            '#2DC4B2',
                            100,
                            '#3BB3C3',
                            750,
                            '#669EC4'
                        ],
                        'circle-radius': [
                            'step',
                            ['get', 'point_count'],
                            20,
                            100,
                            30,
                            750,
                            40
                        ]
                    }
                });

                this.map.addLayer({
                    id: 'cluster-count',
                    type: 'symbol',
                    source: 'photos',
                    filter: ['has', 'point_count'],
                    layout: {
                        'text-field': '{point_count_abbreviated}',
                        'text-font': ['Roboto', 'sans-serif'],
                        'text-size': 13
                    }
                });

                this.map.on('render', this.updateMarkers);

                this.map.on('click', 'clusters', (e) => {
                    const features = this.map.queryRenderedFeatures(e.point, {
                        layers: ['clusters']
                    });
                    const clusterId = features[0].properties.cluster_id;
                    this.map.getSource('photos').getClusterExpansionZoom(
                        clusterId,
                        (err, zoom) => {
                            if (err) return;

                            this.map.easeTo({
                                center: features[0].geometry.coordinates,
                                zoom: zoom
                            });
                        }
                    );
                });

                this.map.on('mouseenter', 'clusters', () => {
                    this.map.getCanvas().style.cursor = 'pointer';
                });
                this.map.on('mouseleave', 'clusters', () => {
                    this.map.getCanvas().style.cursor = '';
                });

                this.search();
            },
        },
        mounted() {
            this.renderMap();
        },
    };
</script>

