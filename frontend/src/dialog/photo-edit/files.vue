<template>
    <div class="p-tab p-tab-photo-edit-files">
        <v-data-table
                :headers="listColumns"
                :items="model.Files"
                hide-actions
                disable-initial-sort
                class="elevation-0 p-files p-files-list p-results"
                item-key="ID"
                v-model="selected"
                :no-data-text="this.$gettext('No files found')"
        >
            <template slot="items" slot-scope="props" class="p-file">
                <td>
                    <a :href="'/api/v1/download/' + props.item.FileHash" class="secondary-dark--text" target="_blank">
                        {{ props.item.FileName }}
                    </a>
                </td>
                <td>{{ props.item.FileType }}</td>
                <td>{{ props.item.FileWidth ? props.item.FileWidth : "" }}</td>
                <td>{{ props.item.FileHeight ? props.item.FileHeight : "" }}</td>
                <td class="text-xs-center">
                    <v-icon v-if="props.item.FilePrimary" color="secondary-dark">check_box</v-icon>
                    <v-icon v-else color="secondary-dark">check_box_outline_blank</v-icon>
                </td>
                <td class="text-xs-center">
                    <v-icon v-if="props.item.FileSidecar" color="secondary-dark">check_box</v-icon>
                    <v-icon v-else color="secondary-dark">check_box_outline_blank</v-icon>
                </td>
                <td class="text-xs-center">
                    <v-icon v-if="props.item.FileMissing" color="secondary-dark">check_box</v-icon>
                    <v-icon v-else color="secondary-dark">check_box_outline_blank</v-icon>
                </td>
                <td class="text-xs-center">
                    <v-icon v-if="props.item.FileDuplicate" color="secondary-dark">check_box</v-icon>
                    <v-icon v-else color="secondary-dark">check_box_outline_blank</v-icon>
                </td>
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
                    {text: this.$gettext('Name'), value: 'FileName', sortable: false, align: 'left'},
                    {text: this.$gettext('Type'), value: 'FileType', sortable: false},
                    {text: this.$gettext('Width'), value: 'FileWidth', sortable: false},
                    {text: this.$gettext('Height'), value: 'FileHeight', sortable: false},
                    {text: this.$gettext('Primary'), value: 'FilePrimary', sortable: false, align: 'center'},
                    {text: this.$gettext('Sidecar'), value: 'FileSidecar', sortable: false, align: 'center'},
                    {text: this.$gettext('Missing'), value: 'FileMissing', sortable: false, align: 'center'},
                    {text: this.$gettext('Duplicate'), value: 'FileDuplicate', sortable: false, align: 'center'},
                ],
            };
        },
        computed: {},
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
