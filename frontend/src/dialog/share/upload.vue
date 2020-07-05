<template>
  <v-dialog lazy v-model="show" persistent max-width="400" class="p-share-upload-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs8>
            <h3 class="headline mb-0">
              <translate>WebDAV Upload</translate>
            </h3>
          </v-flex>
          <v-flex xs4 text-xs-right>
            <v-btn icon flat dark color="secondary-dark" class="ma-0" @click.stop="setup">
              <v-icon>cloud</v-icon>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text class="pt-0">
        <v-layout row wrap>
          <v-flex xs12 text-xs-left class="pt-2">
            <v-select
                    color="secondary-dark"
                    hide-details hide-no-data flat
                    :label="$gettext('Account')"
                    item-text="AccName"
                    item-value="ID"
                    @change="onChange"
                    return-object
                    :disabled="loading || noAccounts"
                    v-model="account"
                    :items="accounts">
            </v-select>
          </v-flex>
          <v-flex xs12 text-xs-left class="pt-2">
            <v-autocomplete
                    color="secondary-dark"
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
                    :label="$gettext('Folder')"
            >
            </v-autocomplete>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-4">
            <v-btn @click.stop="cancel" depressed color="secondary-light" class="action-cancel ml-0 mt-0 mb-0 mr-2">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn color="secondary-dark" depressed dark @click.stop="setup"
                   class="action-setup ma-0" v-if="noAccounts">
              <translate>Setup</translate>
            </v-btn>
            <v-btn color="secondary-dark" depressed dark @click.stop="confirm"
                   class="action-upload ma-0" v-else>
              <translate>Upload</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
<script>
    import Account from "model/account";

    export default {
        name: 'p-share-upload-dialog',
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
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            setup() {
                this.$router.push({name: "settings_sync"});
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
                            this.$notify.success("One file uploaded");
                        } else {
                            this.$notify.success(this.$gettextInterpolate(this.$gettext("%{n} files uploaded"), {n: files.length}));
                        }

                        this.$emit('confirm', this.account);
                    }
                ).catch(() => this.loading = false);
            },
            onChange() {
                this.paths = [{"abs": "/"}];

                this.loading = true;
                this.account.Folders().then(p => {
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
