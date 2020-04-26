<template>
    <div class="p-tab p-settings-accounts">
        <v-data-table
                :headers="listColumns"
                :items="results"
                hide-actions
                disable-initial-sort
                class="elevation-0 p-accounts p-accounts-list p-results"
                item-key="ID"
                v-model="selected"
                :no-data-text="this.$gettext('No accounts configured')"
        >
            <template slot="items" slot-scope="props" class="p-account">
                <td>
                    <button @click.stop.prevent="edit(props.item)" class="secondary-dark--text">
                        {{ props.item.AccName }}
                    </button>
                </td>
                <td class="text-xs-center">
                    <v-btn icon small flat :ripple="false"
                           class="action-toggle-share"
                           @click.stop.prevent="editSharing(props.item)">
                        <v-icon v-if="props.item.AccShare" color="secondary-dark">check</v-icon>
                        <v-icon v-else color="secondary-dark">settings</v-icon>
                    </v-btn>
                </td>
                <td class="text-xs-center"><v-btn icon small flat :ripple="false"
                           class="action-toggle-sync"
                           @click.stop.prevent="editSync(props.item)">
                    <v-icon v-if="props.item.AccSync" color="secondary-dark">sync</v-icon>
                    <v-icon v-else color="secondary-dark">sync_disabled</v-icon>
                </v-btn></td>
                <td class="hidden-sm-and-down">{{ formatDate(props.item.SyncDate) }}</td>
                <td class="hidden-xs-only text-xs-right" nowrap>
                    <v-btn icon small flat :ripple="false"
                           class="p-account-remove"
                           @click.stop.prevent="remove(props.item)">
                        <v-icon color="secondary-dark">delete</v-icon>
                    </v-btn>
                    <v-btn icon small flat :ripple="false"
                           class="p-account-remove"
                           @click.stop.prevent="edit(props.item)">
                        <v-icon color="secondary-dark">edit</v-icon>
                    </v-btn>
                </td>
            </template>
        </v-data-table>
        <v-container fluid>
            <v-form lazy-validation dense
                    ref="form" class="p-form-settings" accept-charset="UTF-8"
                    @submit.prevent="add">
                <v-btn color="secondary-dark"
                       class="white--text ml-0 mt-2"
                       depressed
                       @click.stop="add">
                    <translate>Add</translate>
                    <v-icon right dark>add</v-icon>
                </v-btn>
            </v-form>
        </v-container>
        <p-account-add-dialog :show="dialog.add" @cancel="onCancel('add')"
                                 @confirm="onAdded"></p-account-add-dialog>
        <p-account-remove-dialog :show="dialog.remove" :model="model" @cancel="onCancel('remove')"
                                 @confirm="onRemoved"></p-account-remove-dialog>
        <p-account-edit-dialog :show="dialog.edit" :model="model" :scope="editScope" @remove="remove(model)" @cancel="onCancel('edit')"
                                 @confirm="onEdited"></p-account-edit-dialog>
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
                readonly: this.$config.get("readonly"),
                settings: new Settings(this.$config.values.settings),
                options: options,
                model: {},
                results: [],
                labels: {},
                selected: [],
                dialog: {
                    add: false,
                    remove: false,
                },
                editScope: "main",
                listColumns: [
                    {text: this.$gettext('Name'), value: 'AccName', sortable: false, align: 'left'},
                    {text: this.$gettext('Upload'), value: 'AccShare', sortable: false, align: 'center'},
                    {text: this.$gettext('Sync'), value: 'AccSync', sortable: false, align: 'center'},
                    {text: this.$gettext('Synced'), value: 'SyncDate', sortable: false, class: 'hidden-sm-and-down', align: 'left'},
                    {text: '', value: '', sortable: false, class: 'hidden-xs-only', align: 'right'},
                ],
            };
        },
        methods: {
            formatDate(d) {
                if (!d || !d.Valid) {
                    return this.$gettext('Never');
                }

                const time = d.Time ? d.Time : d;

                return DateTime.fromISO(time).toLocaleString(DateTime.DATE_FULL);
            },
            load() {
                Account.search({count: 100}).then(r => this.results = r.models);
            },
            remove(model) {
                this.model = model.clone();

                this.dialog.edit = false;
                this.dialog.remove = true;
            },
            onRemoved() {
                this.dialog.remove = false;
                this.model = {};
                this.load();
            },
            editSharing(model) {
                this.model = model.clone();

                this.editScope = "sharing";

                this.dialog.edit = true;
            },
            editSync(model) {
                this.model = model.clone();

                this.editScope = "sync";

                this.dialog.edit = true;
            },
            edit(model) {
                this.model = model.clone();

                this.editScope = "account";

                this.dialog.edit = true;
            },
            onEdited() {
                this.dialog.edit = false;
                this.model = {};
                this.load();
            },
            add() {
                this.dialog.add = true;
            },
            onAdded() {
                this.dialog.add = false;
                this.load();
            },
            onCancel(name) {
                this.dialog[name] = false;
                this.model = {};
            },
        },
        created() {
            this.load();
        },
    };
</script>
