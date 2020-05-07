<template>
    <v-dialog lazy v-model="show" persistent max-width="400" class="p-photo-share-dialog" @keydown.esc="cancel">
        <v-card raised elevation="24">
            <v-container fluid class="pb-2 pr-2 pl-2">
                <v-layout row wrap>
                    <v-flex xs3 text-xs-center>
                        <v-icon size="54" color="grey lighten-1">share</v-icon>
                    </v-flex>
                    <v-flex xs9 text-xs-left align-self-center>
                        <v-select
                                color="secondary-dark"
                                class="mr-2"
                                hide-details hide-no-data flat
                                :label="labels.account"
                                item-text="AccName"
                                item-value="ID"
                                @change="onChange"
                                return-object
                                :disabled="loading || noAccounts"
                                v-model="account"
                                :items="accounts">
                        </v-select>

                        <v-autocomplete
                                color="secondary-dark"
                                class="mt-3 mr-2"
                                hide-details hide-no-data flat
                                v-model="path"
                                browser-autocomplete="off"
                                hint="Folder"
                                :search-input.sync="search"
                                :items="pathItems"
                                :loading="loading"
                                :disabled="loading || noAccounts"
                                item-text="abs"
                                item-value="abs"
                                :label="labels.path"
                        >
                        </v-autocomplete>
                    </v-flex>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="grey lighten-3" class="action-cancel">
                            <translate>Cancel</translate>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="setup"
                               class="action-setup" v-if="noAccounts">
                            <span>{{ labels.setup }}</span>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="action-upload" v-else>
                            <span>{{ labels.upload }}</span>
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import Account from "../model/account";

    export default {
        name: 'p-photo-share-dialog',
        props: {
            show: Boolean,
            selection: Array,
        },
        data() {
            return {
                noAccounts: false,
                loading: true,
                search: null,
                account: {},
                accounts: [],
                path: "/",
                paths: [
                    {"abs": "/"}
                ],
                pathItems: [],
                newPath: "",
                labels: {
                    account: this.$gettext("Account"),
                    path: this.$gettext("Folder"),
                    upload: this.$gettext("Upload"),
                    setup: this.$gettext("Setup"),
                }
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            setup() {
                this.$router.push({name: "settings_accounts"});
            },
            confirm() {
                if (this.loading) {
                    this.$notify.wait();
                    return;
                }

                this.loading = true;
                this.account.Share(this.selection, this.path).then(
                    (files) => {
                        this.loading = false;

                        if (files.length === 1) {
                            this.$notify.success("One photo shared");
                        } else {
                            this.$notify.success(files.length + " photos shared");
                        }

                        this.$emit('confirm', this.account);
                    }
                ).catch(() => this.loading = false);
            },
            onChange() {
                this.paths = [{"abs": "/"}];

                this.loading = true;
                this.account.Dirs().then(p => {
                    for (let i = 0; i < p.length; i++) {
                        this.paths.push(p[i]);
                    }

                    this.pathItems = [...this.paths];
                    this.path = this.account.SharePath;
                }).finally(() => this.loading = false);
            },
            load() {
                this.loading = true;

                const params = {
                    share: true,
                    count: 1000,
                    offset: 0,
                };

                Account.search(params).then(response => {
                    if (!response.models.length) {
                        this.noAccounts = true;
                        this.loading = false;
                    } else {
                        this.account = response.models[0];
                        this.accounts = response.models;
                        this.onChange();
                    }
                }).catch(() => this.loading = false)
            },
        },
        watch: {
            search(q) {
                if (this.loading) return;

                const exists = this.paths.findIndex((p) => p.value === q);

                if (exists !== -1 || !q) {
                    this.pathItems = this.paths;
                    this.newPath = "";
                } else {
                    this.newPath = q;
                    this.pathItems = this.paths.concat([{"abs": q}]);
                }
            },
            show: function (show) {
                if (show) {
                    this.load();
                }
            }
        },
    }
</script>
