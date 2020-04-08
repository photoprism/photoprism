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
            <td>
                <v-img class="accent lighten-2" style="cursor: pointer" aspect-ratio="1"
                       :src="props.item.getThumbnailUrl('tile_50')"
                       @click.exact="openPhoto(props.index)"
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
                </v-img>
            </td>
            <td class="p-photo-desc" @click.exact="toggleSelection(props)">
                <v-hover>
                    <button @click.stop.prevent="editPhoto(props.index)" slot-scope="{ hover }">
                        {{ props.item.PhotoTitle }}
                        <v-icon v-if="hover" size="16" title="edit">edit</v-icon>
                    </button>
                </v-hover>
            </td>
            <td class="p-photo-desc hidden-xs-only" @click.exact="toggleSelection(props)">
                <button v-if="props.item.LocationID" @click.stop.prevent="openLocation(props.index)">
                    {{ props.item.getLocation() }}
                </button>
                <span v-else>
                    {{ props.item.getLocation() }}
                </span>
            </td>
            <td class="p-photo-desc hidden-sm-and-down" @click.exact="toggleSelection(props)">
                <button @click.stop.prevent="editPhoto(props.index)">
                    {{ props.item.CameraMake }} {{ props.item.CameraModel }}
                </button>
            </td>
            <td class="p-photo-desc hidden-xs-only" :title="props.item.TakenAt | luxon:format('dd/MM/yyyy HH:mm:ss')">
                <button @click.stop.prevent="editPhoto(props.index)">
                    {{ props.item.TakenAt | luxon:locale }}
                </button>
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
                    {text: this.$gettext('Location'), class: 'hidden-xs-only', value: 'LocLabel'},
                    {text: this.$gettext('Camera'), class: 'hidden-sm-and-down', value: 'CameraModel'},
                    {text: this.$gettext('Taken'), class: 'hidden-xs-only', value: 'TakenAt'},
                    {text: this.$gettext('Favorite'), value: 'PhotoFavorite', align: 'left'},
                ],
            };
        },
        watch: {
            photos: function (photos) {
                this.selected.splice(0);

                for (let i = 0; i <= photos.length; i++) {
                    if(this.$clipboard.has(photos[i])) {
                        this.selected.push(photos[i]);
                    }
                }
            },
            selection: function () {
                this.refreshSelection();
            },
        },
        methods: {
            toggleSelection(props) {
                this.$clipboard.toggle(props.item);
                props.selected = this.$clipboard.has(props.item);
            },
            refreshSelection() {
                this.selected.splice(0);

                for (let i = 0; i <= this.photos.length; i++) {
                    if(this.$clipboard.has(this.photos[i])) {
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
