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
                                hide-details flat
                                :label="labels.account"
                                item-text="AccName"
                                item-value="ID"
                                @change="onChange"
                                return-object
                                v-model="account"
                                :items="accounts">
                        </v-select>

                        <v-autocomplete
                                color="secondary-dark"
                                class="mt-3 mr-2"
                                hide-details hide-no-data flat
                                v-model="path"
                                browser-autocomplete="off"
                                hint="Location"
                                :search-input.sync="search"
                                :items="pathItems"
                                :loading="loading"
                                :disabled="loading"
                                item-text="text"
                                item-value="value"
                                :label="labels.path"
                        >
                        </v-autocomplete>
                    </v-flex>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="grey lighten-3" class="p-photo-dialog-cancel">
                            <translate>Cancel</translate>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="p-photo-dialog-confirm">
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
        },
        data() {
            return {
                loading: true,
                search: null,
                account: {},
                accounts: [],
                path: "/",
                paths: [
                    {"text": "/", "value": "/"}
                ],
                pathItems: [],
                newPath: "",
                labels: {
                    account: this.$gettext("Account"),
                    path: this.$gettext("Location"),
                    upload: this.$gettext("Upload"),
                }
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
                if (this.loading) {
                    this.$notify.wait();
                    return;
                }

                this.$emit('confirm', this.account);
            },
            onChange() {
                this.paths = [{"text": "/", "value": "/"}];

                this.loading = true;
                this.account.Ls().then(l => {
                    console.log("result", l);
                    for (let i = 0; i < l.length; i++) {
                        this.paths.push({"text": l[i], "value": l[i]});
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
                    this.account = response.models[0];
                    this.accounts = response.models;
                    this.onChange();
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
                    this.pathItems = this.paths.concat([{"text": q, "value": q}]);
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
