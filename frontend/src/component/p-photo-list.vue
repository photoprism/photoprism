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
                       v-longclick="longClick"
                       @contextmenu="contextMenu($event, props.item, props.index)"
                       @click="onClick($event, props.item, props.index)"
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

                    <v-btn v-if="selection.length > 0" :flat="true" :ripple="false"
                           icon large absolute
                           class="p-photo-select">
                        <v-icon v-if="selection.length && $clipboard.has(props.item)" color="white"
                                class="t-select t-on">check_circle
                        </v-icon>
                    </v-btn>
                </v-img>
            </td>
            <td class="p-photo-desc p-pointer" @click.exact="openPhoto(props.index)" style="user-select: none;">
                {{ props.item.PhotoTitle }}
            </td>
            <td class="p-photo-desc hidden-xs-only" :title="props.item.TakenAt | luxon:format('dd/MM/yyyy HH:mm:ss')">
                <button @click.stop.prevent="editPhoto(props.index)" style="user-select: none;">
                    {{ props.item.TakenAt | luxon:locale }}
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
            <td>
                <v-btn class="p-photo-like" icon small flat :ripple="false"
                       @click.stop.prevent="props.item.toggleLike()">
                    <v-icon v-if="props.item.PhotoFavorite" color="pink lighten-3">favorite</v-icon>
                    <v-icon v-else color="accent lighten-4">favorite_border</v-icon>
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
                    {text: this.$gettext('Favorite'), value: 'PhotoFavorite', align: 'left'},
                ],
                showLocation: this.$config.settings().features.places,
                wasLong: false,
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
            longClick() {
                this.wasLong = true;
            },
            onClick(ev, model, index) {
                ev.preventDefault();
                ev.stopPropagation();

                if (this.wasLong || ev.shiftKey) {
                    this.selectRange(index);
                } else {
                    this.$clipboard.toggle(model);
                }

                this.wasLong = false;
            },
            contextMenu(ev, model, index) {
                ev.preventDefault();
                ev.stopPropagation();

                if (this.wasLong) {
                    this.selectRange(index);
                } else {
                    this.$clipboard.toggle(model);
                }

                this.wasLong = false;
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
