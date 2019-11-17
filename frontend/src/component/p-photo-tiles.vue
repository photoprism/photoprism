<template>
    <v-container grid-list-xs fluid class="pa-0 p-photos p-photo-tiles">
        <v-card v-if="photos.length === 0" class="p-photos-empty" flat>
            <v-card-title primary-title>
                <div>
                    <h3 class="title mb-3">No photos matched your search</h3>
                    <div>Try using other terms and search options such as category, country and camera.</div>
                </div>
            </v-card-title>
        </v-card>
        <v-row wrap>
            <v-col
                    v-for="(photo, index) in photos"
                    :key="index"
                    v-bind:class="{ selected: $clipboard.has(photo) }"
                    class="p-photo"
                    cols="12"
                    sm="6"
                    md="3"
                    lg="2"
                    d-flex
            >
                <v-hover>
                    <v-card tile slot-scope="{ hover }"
                            :class="$clipboard.has(photo) ? 'elevation-10 ma-0' : 'elevation-0 ma-1'"
                            :title="photo.PhotoTitle">
                        <v-img :src="photo.getThumbnailUrl('tile_500')"
                               aspect-ratio="1"
                               class="grey lighten-2"
                               style="cursor: pointer"
                               @click="openPhoto(index)"
                        >
                            <v-row
                                    slot="placeholder"
                                    fill-height
                                    align-center
                                    justify-center
                                    ma-0
                            >
                                <v-progress-circular indeterminate
                                                     color="grey lighten-5"></v-progress-circular>
                            </v-row>

                            <v-btn v-if="hover || $clipboard.has(photo)" :text="!hover" :ripple="false"
                                   icon large absolute
                                   class="p-photo-select"
                                   @click.stop.prevent="$clipboard.toggle(photo)">
                                <v-icon v-if="selection.length && $clipboard.has(photo)" color="white">check_circle</v-icon>
                                <v-icon v-else-if="!$clipboard.has(photo)" color="grey lighten-3">radio_button_off</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || photo.PhotoFavorite" :text="!hover" :ripple="false"
                                   icon large absolute
                                   class="p-photo-like"
                                   @click.stop.prevent="photo.toggleLike()">
                                <v-icon v-if="photo.PhotoFavorite" color="white">favorite</v-icon>
                                <v-icon v-else color="grey lighten-3">favorite_border</v-icon>
                            </v-btn>
                        </v-img>

                    </v-card>
                </v-hover>
            </v-col>
        </v-row>
    </v-container>
</template>
<script>
    export default {
        name: 'p-photo-tiles',
        props: {
            photos: Array,
            selection: Array,
            openPhoto: Function,
        },
        methods: {
        }
    };
</script>
