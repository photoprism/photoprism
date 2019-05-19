<template>
    <v-data-table
            :headers="listColumns"
            :items="photos"
            hide-actions
            class="elevation-1"
            select-all
            disable-initial-sort
            item-key="ID"
            v-model="selectedPhotos"
            :no-data-text="'No photos matched your search'"
    >
        <template slot="items" slot-scope="props">
            <td>
                <v-checkbox
                        v-model="props.selected"
                        primary
                        hide-details
                ></v-checkbox>
            </td>
            <td>{{ props.item.PhotoTitle }}</td>
            <td>{{ props.item.TakenAt | moment('DD/MM/YYYY hh:mm:ss') }}</td>
            <td>{{ props.item.LocCity }}</td>
            <td>{{ props.item.LocCountry }}</td>
            <td>{{ props.item.CameraMake }} {{ props.item.CameraModel }}</td>
            <td>{{ props.item.PhotoFavorite ? 'Yes' : 'No' }}</td>
        </template>
    </v-data-table>
</template>
<script>
    export default {
        name: 'app-photo-list',
        props: {
            photos: Array,
            selectedPhotos: Array,
            open: Function,
            select: Function,
            like: Function,
        },
        data() {
            return {
                'listColumns': [
                    {text: 'Title', value: 'PhotoTitle'},
                    {text: 'Taken At', value: 'TakenAt'},
                    {text: 'City', value: 'LocCity'},
                    {text: 'Country', value: 'LocCountry'},
                    {text: 'Camera', value: 'CameraModel'},
                    {text: 'Favorite', value: 'PhotoFavorite'},
                ],
            };
        },
        methods: {
        }
    };
</script>
