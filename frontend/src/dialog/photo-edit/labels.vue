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
                <td class="text-xs-center">{{ props.item.LabelPriority }}</td>
                <td>{{ props.item.CreatedAt | luxon:format('dd/MM/yyyy') }}</td>
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
                    {text: this.$gettext('Label'), value: 'LabelName', sortable: false, align: 'left'},
                    {text: this.$gettext('Priority'), value: 'LabelPriority', sortable: false, align: 'center'},
                    {text: this.$gettext('Added'), value: 'CreatedAt', sortable: false, align: 'left'},
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
