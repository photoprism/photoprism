<template>
    <v-container grid-list-xs fluid class="pa-0 p-photos p-photo-details">
        <v-card v-if="photos.length === 0" class="p-photos-empty">
            <v-card-title primary-title>
                <div>
                    <h3 class="headline mb-3">No photos matched your search</h3>
                    <div>Try using other terms and search options such as category, country and camera.</div>
                </div>
            </v-card-title>
        </v-card>
        <v-layout row wrap>
            <v-flex
                    v-for="(photo, index) in photos"
                    :key="photo.ID"
                    class="p-photo"
                    xs12 sm6 md4 lg3 d-flex
            >
                <v-hover>
                    <v-card tile slot-scope="{ hover }"
                            :dark="photo.selected"
                            :class="photo.selected ? 'elevation-14 ma-1' : 'elevation-2 ma-2'">
                        <v-img
                                :src="photo.getThumbnailUrl('tile_500')"
                                aspect-ratio="1"
                                v-bind:class="{ selected: photo.selected }"
                                style="cursor: pointer"
                                class="grey lighten-2"
                                @click="open(index)"

                        >
                            <v-layout
                                    slot="placeholder"
                                    fill-height
                                    align-center
                                    justify-center
                                    ma-0
                            >
                                <v-progress-circular indeterminate color="grey lighten-5"></v-progress-circular>
                            </v-layout>

                            <v-btn v-if="hover || photo.selected" :flat="!hover" icon large absolute
                                   :ripple="false" style="right: 4px; bottom: 4px;"
                                   @click.stop.prevent="select(photo)">
                                <v-icon v-if="photo.selected" color="white">check_box</v-icon>
                                <v-icon v-else color="white">check_box_outline_blank</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" icon large absolute
                                   :ripple="false" style="bottom: 4px; left: 4px"
                                   @click.stop.prevent="like(photo)">
                                <v-icon v-if="photo.PhotoFavorite" color="white">favorite
                                </v-icon>
                                <v-icon v-else color="white">favorite_border</v-icon>
                            </v-btn>
                        </v-img>


                        <v-card-title primary-title class="pa-3">
                            <div>
                                <h3 class="subheading mb-2" :title="photo.PhotoTitle">{{ photo.PhotoTitle |
                                    truncate(80) }}</h3>
                                <div class="caption">
                                    <v-icon size="14">date_range</v-icon>
                                    {{ photo.TakenAt | moment('DD/MM/YYYY hh:mm:ss') }}
                                    <br/>
                                    <v-icon size="14">photo_camera</v-icon>
                                    {{ photo.getCamera() }}
                                    <br/>
                                    <v-icon size="14">location_on</v-icon>
                                    <span class="link" :title="photo.getFullLocation()"
                                          @click.stop="openLocation(photo)">{{ photo.getLocation() }}</span>
                                </div>
                            </div>
                        </v-card-title>
                    </v-card>
                </v-hover>
            </v-flex>
        </v-layout>
    </v-container>
</template>
<script>
    export default {
        name: 'PPhotoDetails',
        props: {
            photos: Array,
            open: Function,
            select: Function,
            like: Function,
        },
        methods: {
            openLocation(photo) {
                if (photo.PhotoLat && photo.PhotoLong) {
                    this.$router.push({name: 'Places', query: {lat: photo.PhotoLat, long: photo.PhotoLong}});
                } else if (photo.LocName) {
                    this.$router.push({name: 'Places', query: {q: photo.LocName}});
                } else if (photo.LocCity) {
                    this.$router.push({name: 'Places', query: {q: photo.LocCity}});
                } else if (photo.LocCountry) {
                    this.$router.push({name: 'Places', query: {q: photo.LocCountry}});
                } else {
                    this.$router.push({name: 'Places', query: {q: photo.CountryName}});
                }
            },
        }
    };
</script>
