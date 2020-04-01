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
                <td>{{ props.item.FileWidth ? props.item.FileWidth : "" }}</td>
                <td>{{ props.item.FileHeight ? props.item.FileHeight : "" }}</td>
                <td>{{ fileType(props.item) }}</td>
                <td>{{ fileStatus(props.item) }}</td>
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
                    {text: this.$gettext('Width'), value: 'FileWidth', sortable: false},
                    {text: this.$gettext('Height'), value: 'FileHeight', sortable: false},
                    {text: this.$gettext('Type'), value: '', sortable: false, align: 'left'},
                    {text: this.$gettext('Status'), value: '', sortable: false, align: 'left'},
                ],
            };
        },
        computed: {},
        methods: {
            openPhoto() {
                this.$viewer.show([this.model], 0)
            },
            fileType(file) {
                if (file.FilePrimary) {
                    return this.$gettext("Primary");
                } else if (file.FileVideo) {
                    return this.$gettext("Video");
                } else if (file.FileSidecar) {
                    return this.$gettext("Sidecar");
                }

                return file.FileType.toUpperCase();
            },
            fileStatus(file) {
                if (file.FileMissing) {
                    return this.$gettext("Missing");
                } else if (file.FileError) {
                    return file.FileError;
                } else if (file.Duplicate) {
                    return this.$gettext("Duplicate");
                }

                return "OK";
            },
            refresh() {
            },
        },
    };
</script>
