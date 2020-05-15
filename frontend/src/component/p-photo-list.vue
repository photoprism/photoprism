<template>
    <v-data-table
            :headers="listColumns"
            :items="photos"
            hide-actions
            class="elevation-0 p-photos p-photo-list p-results"
            disable-initial-sort
            item-key="ID"
            v-model="selected"
            :no-data-text="this.$gettext('No photos matched your search')"
    >
        <template slot="items" slot-scope="props">
            <td style="user-select: none;">
                <v-img class="accent lighten-2" style="cursor: pointer" aspect-ratio="1"
                       :src="props.item.getThumbnailUrl('tile_50')"
                       @mousedown="onMouseDown($event, props.index)"
                       @contextmenu="onContextMenu($event, props.index)"
                       @click.stop.prevent="onClick($event, props.index)"
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

                    <v-btn v-if="selection.length && $clipboard.has(props.item)" :ripple="false"
                           flat icon large absolute class="p-photo-select">
                        <v-icon color="white" class="t-select t-on">check_circle</v-icon>
                    </v-btn>
                    <v-btn v-else-if="!selection.length && props.item.PhotoVideo && props.item.isPlayable()" :ripple="false"
                           flat icon large absolute class="p-photo-play opacity-75"
                           @click.stop.prevent="openPhoto(props.index, true)">
                        <v-icon color="white" class="action-play">play_arrow</v-icon>
                    </v-btn>
                </v-img>
            </td>
            <td class="p-photo-desc p-pointer" @click.exact="editPhoto(props.index)" style="user-select: none;">
                {{ props.item.PhotoTitle }}
            </td>
            <td class="p-photo-desc hidden-xs-only" :title="props.item.getDateString()">
                <button @click.stop.prevent="editPhoto(props.index)" style="user-select: none;">
                    {{ props.item.shortDateString() }}
                </button>
            </td>
            <td class="p-photo-desc hidden-sm-and-down" style="user-select: none;">
                <button @click.stop.prevent="editPhoto(props.index)">
                    {{ props.item.CameraMake }} {{ props.item.CameraModel }}
                </button>
            </td>
            <td class="p-photo-desc hidden-xs-only">
                <button v-if="props.item.LocationID && showLocation" @click.stop.prevent="openLocation(props.index)"
                        style="user-select: none;">
                    {{ props.item.getLocation() }}
                </button>
                <span v-else>
                    {{ props.item.getLocation() }}
                </span>
            </td>
            <td class="text-xs-center">
                <v-btn v-if="hidePrivate" class="p-photo-private" icon small flat :ripple="false"
                       @click.stop.prevent="props.item.togglePrivate()">
                    <v-icon v-if="props.item.PhotoPrivate" color="secondary-dark">lock</v-icon>
                    <v-icon v-else color="accent lighten-3">lock_open</v-icon>
                </v-btn>
                <v-btn class="p-photo-like" icon small flat :ripple="false"
                       @click.stop.prevent="props.item.toggleLike()">
                    <v-icon v-if="props.item.PhotoFavorite" color="pink lighten-3">favorite</v-icon>
                    <v-icon v-else color="accent lighten-3">favorite_border</v-icon>
                </v-btn>
            </td>
        </template>
    </v-data-table>
</template>
<script>
    export default {
        name: 'p-photo-list',
        props: {
            photos: Array,
            selection: Array,
            openPhoto: Function,
            editPhoto: Function,
            openLocation: Function,
            album: Object,
        },
        data() {
            return {
                'selected': [],
                'listColumns': [
                    {text: '', value: '', align: 'center', sortable: false, class: 'p-col-select'},
                    {text: this.$gettext('Title'), value: 'PhotoTitle'},
                    {text: this.$gettext('Taken'), class: 'hidden-xs-only', value: 'TakenAt'},
                    {text: this.$gettext('Camera'), class: 'hidden-sm-and-down', value: 'CameraModel'},
                    {text: this.$gettext('Location'), class: 'hidden-xs-only', value: 'LocLabel'},
                    {text: '', value: '', sortable: false, align: 'center'},
                ],
                showLocation: this.$config.settings().features.places,
                hidePrivate: this.$config.settings().features.private,
                mouseDown: {
                    index: -1,
                    timeStamp: -1,
                },
            };
        },
        watch: {
            photos: function (photos) {
                this.selected.splice(0);

                for (let i = 0; i <= photos.length; i++) {
                    if (this.$clipboard.has(photos[i])) {
                        this.selected.push(photos[i]);
                    }
                }
            },
            selection: function () {
                this.refreshSelection();
            },
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
                } else if(this.photos[index]) {
                    let photo = this.photos[index];

                    if(photo.PhotoVideo && photo.isPlayable()) {
                        this.openPhoto(index, true);
                    } else {
                        this.openPhoto(index, false);
                    }
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
            },
            refreshSelection() {
                this.selected.splice(0);

                for (let i = 0; i <= this.photos.length; i++) {
                    if (this.$clipboard.has(this.photos[i])) {
                        this.selected.push(this.photos[i]);
                    }
                }
            },
        },
        mounted: function () {
            this.$nextTick(function () {
                this.refreshSelection();
            })
        }
    };
</script>
