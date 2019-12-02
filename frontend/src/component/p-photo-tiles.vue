<template>
    <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-tiles">
        <v-card v-if="photos.length === 0" class="p-photos-empty" flat>
            <v-card-title primary-title>
                <div>
                    <h3 class="title mb-3">No photos matched your search</h3>
                    <div>Try using other terms and search options such as category, country and camera.</div>
                </div>
            </v-card-title>
        </v-card>
        <v-layout row wrap>
            <v-flex
                    v-for="(photo, index) in photos"
                    :key="index"
                    v-bind:class="{ selected: $clipboard.has(photo) }"
                    class="p-photo"
                    xs12 sm6 md3 lg2 d-flex
            >
                <v-hover>
                    <v-card tile slot-scope="{ hover }"
                            :class="$clipboard.has(photo) ? 'elevation-10 ma-0' : 'elevation-0 ma-1'"
                            :title="photo.PhotoTitle">
                        <v-img :src="photo.getThumbnailUrl('tile_500')"
                               aspect-ratio="1"
                               class="accent lighten-2"
                               style="cursor: pointer"
                               @click="openPhoto(index)"
                        >
                            <v-layout
                                    slot="placeholder"
                                    fill-height
                                    align-center
                                    justify-center
                                    ma-0
                            >
                                <v-progress-circular indeterminate
                                                     color="accent lighten-5"></v-progress-circular>
                            </v-layout>

                            <v-btn v-if="hover || $clipboard.has(photo)" :flat="!hover" :ripple="false"
                                   icon large absolute
                                   class="p-photo-select"
                                   @click.stop.prevent="$clipboard.toggle(photo)">
                                <v-icon v-if="selection.length && $clipboard.has(photo)" color="white">check_circle</v-icon>
                                <v-icon v-else-if="!$clipboard.has(photo)" color="accent lighten-3">radio_button_off</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" :ripple="false"
                                   icon large absolute
                                   class="p-photo-like"
                                   @click.stop.prevent="photo.toggleLike()">
                                <v-icon v-if="photo.PhotoFavorite" color="white">favorite</v-icon>
                                <v-icon v-else color="accent lighten-3">favorite_border</v-icon>
                            </v-btn>
                        </v-img>

                    </v-card>
                </v-hover>
            </v-flex>
        </v-layout>
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
