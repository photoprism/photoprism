<template>
    <div class="p-tab p-tab-photo-edit-labels">
        <v-data-table
                :headers="listColumns"
                :items="model.Labels"
                hide-actions
                class="elevation-0 p-files p-files-list p-results"
                disable-initial-sort
                item-key="ID"
                v-model="selected"
                :no-data-text="this.$gettext('No labels found')"
        >
            <template slot="items" slot-scope="props" class="p-file">
                <td>{{ props.item.LabelName }}</td>
                <td>{{ props.item.LabelPriority }}</td>
                <td align="left">{{ props.item.CreatedAt | luxon:format('dd/MM/yyyy hh:mm:ss') }}</td>
            </template>
        </v-data-table>
    </div>
</template>

<script>
    export default {
        name: 'p-tab-photo-edit-labels',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.getValue("readonly"),
                selected: [],
                listColumns: [
                    {text: this.$gettext('Label'), value: 'LabelName', align: 'left'},
                    {text: this.$gettext('Priority'), value: 'LabelPriority'},
                    {text: this.$gettext('Added'), value: 'CreatedAt', align: 'left'},
                ],
            };
        },
        computed: {
        },
        methods: {
            refresh() {
            },
        },
    };
</script>
