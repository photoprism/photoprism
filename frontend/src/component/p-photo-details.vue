<template>
    <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-details">
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
                    class="p-photo"
                    xs12 sm6 md4 lg3 d-flex
            >
                <v-hover>
                    <v-card tile slot-scope="{ hover }"
                            :dark="$clipboard.has(photo)"
                            :class="$clipboard.has(photo) ? 'elevation-10 ma-0 accent darken-1 white--text' : 'elevation-0 ma-1 accent lighten-3'">
                        <v-img
                                :src="photo.getThumbnailUrl('tile_500')"
                                aspect-ratio="1"
                                v-bind:class="{ selected: $clipboard.has(photo) }"
                                style="cursor: pointer"
                                class="accent lighten-2"
                                @click="openPhoto(index)"

                        >
                            <v-layout
                                    slot="placeholder"
                                    fill-height
                                    align-center
                                    justify-center
                                    ma-0
                            >
                                <v-progress-circular indeterminate color="accent lighten-5"></v-progress-circular>
                            </v-layout>

                            <v-btn v-if="hover || $clipboard.has(photo)" :flat="!hover" :ripple="false"
                                   icon large absolute
                                   class="p-photo-select"
                                   @click.stop.prevent="$clipboard.toggle(photo)">
                                <v-icon v-if="selection.length && $clipboard.has(photo)" color="white">check_circle</v-icon>
                                <v-icon v-else color="accent lighten-3">radio_button_off</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" :ripple="false"
                                   icon large absolute
                                   class="p-photo-like"
                                   @click.stop.prevent="photo.toggleLike()">
                                <v-icon v-if="photo.PhotoFavorite" color="white">favorite
                                </v-icon>
                                <v-icon v-else color="accent lighten-3">favorite_border</v-icon>
                            </v-btn>
                        </v-img>

                        <v-card-title primary-title class="pa-3">
                            <div>
                                <h3 class="body-2 mb-2" :title="photo.PhotoTitle">
                                    {{ photo.PhotoTitle | truncate(80) }}
                                    <v-icon v-if="photo.PhotoPrivate" size="16" title="Private">vpn_key</v-icon>
                                    <v-icon v-if="photo.PhotoStory" size="16" title="Shared with your friends in the story feed">wifi</v-icon>
                                </h3>
                                <div class="caption">
                                    <v-icon size="14">date_range</v-icon>
                                    {{ photo.getDateString() }}
                                    <br/>
                                    <v-icon size="14">photo_camera</v-icon>
                                    {{ photo.getCamera() }}
                                    <br/>
                                    <v-icon size="14">location_on</v-icon>
                                    <span class="p-pointer" :title="photo.getFullLocation()"
                                          @click.stop="openLocation(index)">{{ photo.getLocation() }}</span>
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
        name: 'p-photo-details',
        props: {
            photos: Array,
            selection: Array,
            openPhoto: Function,
            openLocation: Function,
        },
        methods: {
        }
    };
</script>
