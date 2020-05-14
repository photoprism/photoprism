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
                    <v-btn v-if="props.item.FileType === 'jpg'" flat :ripple="false" icon small
                           @click.stop.prevent="setPrimary(props.item)">
                        <v-icon v-if="props.item.FilePrimary" color="secondary-dark">radio_button_checked</v-icon>
                        <v-icon v-else color="secondary-dark">radio_button_unchecked</v-icon>
                    </v-btn>
                </td>
                <td>
                    <a :href="'/api/v1/download/' + props.item.FileHash" class="secondary-dark--text" target="_blank"
                       v-if="$config.feature('download')">
                        {{ props.item.FileName }}
                    </a>
                    <span v-else>
                        {{ props.item.FileName }}
                    </span>
                </td>
                <td class="hidden-sm-and-down">{{ fileDimensions(props.item) }}</td>
                <td class="hidden-xs-only">{{ fileSize(props.item) }}</td>
                <td>{{ fileType(props.item) }}</td>
                <td>{{ fileStatus(props.item) }}</td>
            </template>
        </v-data-table>
    </div>
</template>

<script>
    import Thumb from "model/thumb";

    export default {
        name: 'p-tab-photo-edit-files',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.get("readonly"),
                selected: [],
                listColumns: [
                    {text: this.$gettext('Primary'), value: 'FilePrimary', sortable: false, align: 'center', class: 'p-col-primary'},
                    {text: this.$gettext('Name'), value: 'FileName', sortable: false, align: 'left'},
                    {text: this.$gettext('Dimensions'), value: '', sortable: false, class: 'hidden-sm-and-down'},
                    {text: this.$gettext('Size'), value: 'FileSize', sortable: false, class: 'hidden-xs-only'},
                    {text: this.$gettext('Type'), value: '', sortable: false, align: 'left'},
                    {text: this.$gettext('Status'), value: '', sortable: false, align: 'left'},
                ],
            };
        },
        computed: {},
        methods: {
            openPhoto() {
                this.$viewer.show(Thumb.fromFiles([this.model]), 0)
            },
            setPrimary(file) {
                this.model.setPrimary(file.FileUUID);
            },
            fileDimensions(file) {
                if (!file.FileWidth || !file.FileHeight) {
                    return "";
                }

                return file.FileWidth + " × " + file.FileHeight;
            },
            fileSize(file) {
                if (!file.FileSize) {
                    return "";
                }

                const size = Number.parseFloat(file.FileSize) / 1048576;

                return size.toFixed(1) + " MB";
            },
            fileType(file) {
                if (file.FileVideo) {
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
