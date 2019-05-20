<template>
    <v-container grid-list-xs fluid class="pa-0 p-photos p-photo-tiles">
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
                    :key="index"
                    v-bind:class="{ selected: selection.includes(photo.ID) }"
                    class="p-photo"
                    xs12 sm6 md3 lg2 d-flex
            >
                <v-hover>
                    <v-card tile slot-scope="{ hover }"
                            :class="selection.includes(photo.ID) ? 'elevation-15 ma-1' : hover ? 'elevation-6 ma-2' : 'elevation-2 ma-2'">
                        <v-img :src="photo.getThumbnailUrl('tile_500')"
                               aspect-ratio="1"
                               class="grey lighten-2"
                               style="cursor: pointer"
                               @click="open(index)"
                        >
                            <v-layout
                                    slot="placeholder"
                                    fill-height
                                    align-center
                                    justify-center
                                    ma-0
                            >
                                <v-progress-circular indeterminate
                                                     color="grey lighten-5"></v-progress-circular>
                            </v-layout>

                            <v-btn v-if="hover || selection.includes(photo.ID)" :flat="!hover" icon large absolute
                                   :ripple="false" style="right: 4px; bottom: 4px;"
                                   class="p-photo-select"
                                   @click.stop.prevent="select(photo)">
                                <v-icon v-if="selection.length && selection.includes(photo.ID)" color="white">check_box</v-icon>
                                <v-icon v-else-if="!selection.includes(photo.ID)" color="white">check_box_outline_blank</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" icon large absolute
                                   :ripple="false" style="bottom: 4px; left: 4px"
                                   class="p-photo-like"
                                   @click.stop.prevent="like(photo)">
                                <v-icon v-if="photo.PhotoFavorite" color="white">favorite</v-icon>
                                <v-icon v-else color="white">favorite_border</v-icon>
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
        name: 'PPhotoTiles',
        props: {
            photos: Array,
            selection: Array,
            select: Function,
            open: Function,
            like: Function,
        },
        methods: {
        }
    };
</script>
