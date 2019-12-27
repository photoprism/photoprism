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
        <template slot="items" slot-scope="props" class="p-photo">
            <td>
                <v-btn icon small :ripple="false"
                       class="p-photo-select"
                       @click.stop.prevent="$clipboard.toggle(props.item)">
                    <v-icon v-if="selection.length && $clipboard.has(props.item)" color="accent darken-2">check_circle</v-icon>
                    <v-icon v-else-if="!$clipboard.has(props.item)" color="accent lighten-4">radio_button_off</v-icon>
                </v-btn>
            </td>
            <td @click="openPhoto(props.index)" class="p-pointer">{{ props.item.PhotoTitle }}</td>
            <td>{{ props.item.TakenAt | luxon:format('dd/MM/yyyy hh:mm:ss') }}</td>
            <td @click="openLocation(props.index)" class="p-pointer">{{ props.item.getLocation() }}</td>
            <td>{{ props.item.CameraMake }} {{ props.item.CameraModel }}</td>
            <td><v-btn icon small flat :ripple="false"
                       class="p-photo-like"
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
            openLocation: Function,
            album: Object,
        },
        data() {
            return {
                'selected': [],
                'listColumns': [
                    {text: '', value: '', align: 'center', sortable: false, class: 'p-col-select'},
                    {text: this.$gettext('Title'), value: 'PhotoTitle'},
                    {text: this.$gettext('Taken At'), value: 'TakenAt'},
                    {text: this.$gettext('Location'), value: 'LocPlace'},
                    {text: this.$gettext('Camera'), value: 'CameraModel'},
                    {text: this.$gettext('Favorite'), value: 'PhotoFavorite', align: 'left'},
                ],
            };
        },
        methods: {
        }
    };
</script>
