<template>
    <v-dialog lazy v-model="show" persistent max-width="500" class="p-account-create-dialog" @keydown.esc="cancel">
        <v-card raised elevation="24">
            <v-card-title primary-title>
                <div>
                    <h3 class="headline mb-0"><translate>Add Account</translate></h3>
                </div>
            </v-card-title>
            <v-container fluid class="pt-0 pb-2 pr-2 pl-2">
                <v-layout row wrap>
                    <v-flex xs12 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.url"
                                placeholder="https://www.example.com/"
                                color="secondary-dark"
                                v-model="model.AccURL"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.user"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccUser"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.pass"
                                placeholder="optional"
                                color="secondary-dark"
                                v-model="model.AccPass"
                                :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                                :type="showPassword ? 'text' : 'password'"
                                @click:append="showPassword = !showPassword"
                        ></v-text-field>
                    </v-flex>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="secondary-light"
                               class="p-account-create-dialog-cancel">
                            <span>{{ label.cancel }}</span>
                        </v-btn>
                        <v-btn depressed dark color="secondary-dark" @click.stop="confirm"
                               class="p-account-create-dialog-confirm">
                            <span>{{ label.confirm }}</span>
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import Account from "model/account";

    export default {
        name: 'p-account-create-dialog',
        props: {
            show: Boolean,
        },
        data() {
            return {
                showPassword: false,
                loading: false,
                search: null,
                model: new Account(),
                label: {
                    url: this.$gettext("Service URL"),
                    user: this.$gettext("Username"),
                    pass: this.$gettext("Password"),
                    cancel: this.$gettext("Cancel"),
                    confirm: this.$gettext("Connect"),
                }
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
                this.loading = true;

                this.model.save().then((a) => {
                    this.loading = false;
                    this.$emit('confirm', a.AlbumUUID);
                });
            },
        },
        watch: {
            show: function (show) {
            }
        },
    }
</script>
