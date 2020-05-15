<template>
    <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
        <v-card v-if="photos.length === 0" class="p-photos-empty secondary-light lighten-1 ma-1" flat>
            <v-card-title primary-title>
                <div>
                    <h3 class="title mb-3">
                        <translate>No photos matched your search</translate>
                    </h3>
                    <div>
                        <translate>Try using other terms and search options such as category, country and camera.
                        </translate>
                    </div>
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
                            @contextmenu="onContextMenu($event, index)"
                            :class="$clipboard.has(photo) ? 'elevation-10 ma-0' : 'elevation-0 ma-1'"
                            :title="photo.PhotoTitle">
                        <v-img :src="photo.getThumbnailUrl('tile_224')"
                               aspect-ratio="1"
                               class="accent lighten-2"
                               style="cursor: pointer"
                               @mousedown="onMouseDown($event, index)"
                               @click.stop.prevent="onClick($event, index)"
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

                            <v-btn v-if="hidePrivate && photo.PhotoPrivate"  :ripple="false"
                                   icon flat small absolute
                                   class="p-photo-private opacity-75">
                                <v-icon color="white">lock</v-icon>
                            </v-btn>

                            <v-btn v-if="hover || selection.length && $clipboard.has(photo)"  :ripple="false"
                                   icon flat small absolute
                                   :class="selection.length && $clipboard.has(photo) ? 'p-photo-select' : 'p-photo-select opacity-50'"
                                   @click.stop.prevent="onSelect($event, index)">
                                <v-icon v-if="selection.length && $clipboard.has(photo)" color="white"
                                        class="t-select t-on">check_circle
                                </v-icon>
                                <v-icon v-else color="accent lighten-3" class="t-select t-off">radio_button_off</v-icon>
                            </v-btn>

                            <v-btn icon flat small absolute  :ripple="false"
                                   :class="photo.PhotoFavorite ? 'p-photo-like opacity-75' : 'p-photo-like opacity-50'"
                                   @click.stop.prevent="photo.toggleLike()">
                                <v-icon v-if="photo.PhotoFavorite" color="white" class="t-like t-on">favorite</v-icon>
                                <v-icon v-else color="accent lighten-3" class="t-like t-off">favorite_border</v-icon>
                            </v-btn>

                            <v-btn v-if="photo.PhotoVideo && photo.isPlayable()" color="white"
                                   outline fab absolute class="p-photo-play opacity-75" :depressed="false" :ripple="false"
                                   @click.stop.prevent="openPhoto(index, true)">
                                <v-icon color="white" class="action-play">play_arrow</v-icon>
                            </v-btn>
                            <v-btn v-else-if="!photo.PhotoVideo && photo.Files.length > 1"  :ripple="false"
                                   icon flat small absolute class="p-photo-merged opacity-75"
                                   @click.stop.prevent="openPhoto(index, true)">
                                <v-icon color="white" class="action-burst">burst_mode</v-icon>
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
            editPhoto: Function,
            album: Object,
        },
        data() {
            return {
                hidePrivate: this.$config.settings().features.private,
                mouseDown: {
                    index: -1,
                    timeStamp: -1,
                },
            };
        },
        methods: {
            onSelect(ev, index) {
                if (ev.shiftKey) {
                    this.selectRange(index);
                } else {
                    this.$clipboard.toggle(this.photos[index]);
                }
            },
            onMouseDown(ev, index) {
                this.mouseDown.index = index;
                this.mouseDown.timeStamp = ev.timeStamp;
            },
            onClick(ev, index) {
                let longClick = (this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400);

                if (longClick || this.selection.length > 0) {
                    if (longClick || ev.shiftKey) {
                        this.selectRange(index);
                    } else {
                        this.$clipboard.toggle(this.photos[index]);
                    }
                } else {
                    this.openPhoto(index, false);
                }
            },
            onContextMenu(ev, index) {
                if (this.$isMobile) {
                    ev.preventDefault();
                    ev.stopPropagation();
                    this.selectRange(index);
                }
            },
            selectRange(index) {
                this.$clipboard.addRange(index, this.photos);
            }
        },
    };
</script>
