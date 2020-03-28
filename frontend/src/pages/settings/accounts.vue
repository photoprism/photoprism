<template>
    <div class="p-tab p-settings-accounts">
        <v-data-table
                :headers="listColumns"
                :items="accounts"
                hide-actions
                disable-initial-sort
                class="elevation-0 p-accounts p-accounts-list p-results"
                item-key="ID"
                v-model="selected"
                :no-data-text="this.$gettext('No accounts configured')"
        >
            <template slot="items" slot-scope="props" class="p-account">
                <td>{{ props.item.AccName }}</td>
                <td>{{ formatBool(props.item.AccShare) }}</td>
                <td>{{ formatBool(props.item.AccSync) }}</td>
                <td>{{ formatDate(props.item.AccSyncedAt) }}</td>
            </template>
        </v-data-table>
    </div>
</template>

<script>
    import Settings from "model/settings";
    import options from "resources/options.json";
    import Account from "../../model/account";
    import {DateTime} from "luxon";

    export default {
        name: 'p-settings-accounts',
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.getValue("readonly"),
                settings: new Settings(this.$config.values.settings),
                options: options,
                model: new Account(),
                accounts: [],
                labels: {},
                selected: [],
                listColumns: [
                    {text: this.$gettext('Name'), value: 'AccName', sortable: false, align: 'left'},
                    {text: this.$gettext('Share'), value: 'AccShare', sortable: false},
                    {text: this.$gettext('Sync'), value: 'AccSync', sortable: false},
                    {text: this.$gettext('Synced'), value: 'AccSyncedAt', sortable: false, align: 'left'},
                ],
            };
        },
        methods: {
            formatBool(b) {
                if (b) {
                    return this.$gettext('Yes');
                }

                return this.$gettext('No');
            },
            formatDate(d) {
                if (!d) {
                    return this.$gettext('Never');
                }

                return DateTime.fromISO(d).toFormat('dd/MM/yyyy hh:mm:ss');
            },
            load() {
                Account.search({count: 100}).then(r => this.accounts = r.models);
            },
            save() {
            },
        },
        created() {
            this.load();
        },
    };
</script>
