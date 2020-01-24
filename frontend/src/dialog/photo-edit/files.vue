<template>
    <div class="p-tab p-tab-photo-edit-files">
        <v-data-table
                :headers="listColumns"
                :items="model.Files"
                hide-actions
                class="elevation-0 p-files p-files-list p-results"
                disable-initial-sort
                item-key="ID"
                v-model="selected"
                :no-data-text="this.$gettext('No files found')"
        >
            <template slot="items" slot-scope="props" class="p-file">
                <td @click="openPhoto(props.index)" class="p-pointer" align="left">{{ props.item.FileName }}</td>
                <td>{{ props.item.FileType }}</td>
                <td>{{ props.item.FileWidth ? props.item.FileWidth : "" }}</td>
                <td>{{ props.item.FileHeight ? props.item.FileHeight : "" }}</td>
                <td><v-btn icon small flat :ripple="false"
                           class="p-photo-like"
                           @click.stop.prevent="changePrimary(props.index)">
                    <v-icon v-if="props.item.FilePrimary" color="secondary-dark">check_box</v-icon>
                    <v-icon v-else color="secondary-dark">check_box_outline_blank</v-icon>
                </v-btn>
                </td>
                <td>{{ props.item.CreatedAt | luxon:format('dd/MM/yyyy hh:mm:ss') }}</td>
            </template>
        </v-data-table>
    </div>
</template>

<script>
    export default {
        name: 'p-tab-photo-edit-files',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.getValue("readonly"),
                selected: [],
                listColumns: [
                    {text: this.$gettext('Name'), value: 'FileName', align: 'left'},
                    {text: this.$gettext('Type'), value: 'FileType'},
                    {text: this.$gettext('Width'), value: 'FileWidth'},
                    {text: this.$gettext('Height'), value: 'FileHeight'},
                    {text: this.$gettext('Primary'), value: 'FilePrimary', align: 'left'},
                    {text: this.$gettext('Added'), value: 'CreatedAt'},
                ],
            };
        },
        computed: {
        },
        methods: {
            openPhoto() {
                this.$viewer.show([this.model], 0)
            },
            changePrimary() {
            },
            refresh() {
            },
        },
    };
</script>
