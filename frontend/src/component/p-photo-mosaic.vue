<template>
    <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
        <v-card v-if="photos.length === 0" class="p-photos-empty secondary-light lighten-1 ma-1" flat>
            <v-card-title primary-title>
                <div>
                    <h3 class="title mb-3"><translate>No photos matched your search</translate></h3>
                    <div><translate>Try using other terms and search options such as category, country and camera.</translate></div>
                </div>
            </v-card-title>
        </v-card>
        <v-layout row wrap class="p-results">
            <v-flex
                    v-for="(photo, index) in photos"
                    :key="index"
                    v-bind:class="{ selected: $clipboard.has(photo) }"
                    class="p-photo"
                    xs4 sm3 md2 xl1 d-flex
            >
                <v-hover>
                    <v-card tile slot-scope="{ hover }"
                            :class="$clipboard.has(photo) ? 'elevation-10 ma-0' : 'elevation-0 ma-1'"
                            :title="photo.PhotoTitle">
                        <v-img :src="photo.getThumbnailUrl('tile_224')"
                               aspect-ratio="1"
                               class="accent lighten-2"
                               style="cursor: pointer"
                               @click.exact="openPhoto(index)"
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

                            <v-btn v-if="hover || selection.length > 0" :flat="!hover" :ripple="false"
                                   icon small absolute
                                   class="p-photo-select"
                                   @click.shift.prevent="$clipboard.addRange(index, photos)"
                                   @click.exact.stop.prevent="$clipboard.toggle(photo)">
                                <v-icon v-if="selection.length && $clipboard.has(photo)" color="white" class="t-select t-on">check_circle</v-icon>
                                <v-icon v-else color="accent lighten-3" class="t-select t-off">radio_button_off</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" :ripple="false"
                                   icon small absolute
                                   class="p-photo-like"
                                   @click.stop.prevent="photo.toggleLike()">
                                <v-icon v-if="photo.PhotoFavorite" color="white" class="t-like t-on">favorite</v-icon>
                                <v-icon v-else color="accent lighten-3" class="t-like t-off">favorite_border</v-icon>
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
        name: 'p-photo-mosaic',
        props: {
            photos: Array,
            selection: Array,
            openPhoto: Function,
            album: Object,
        },
        methods: {
        },
    };
</script>
