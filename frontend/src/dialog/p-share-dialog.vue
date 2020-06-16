<template>
    <v-dialog lazy v-model="show" persistent max-width="500" class="p-share-dialog" @keydown.esc="close">
        <v-card raised elevation="24">
            <v-card-title primary-title class="pb-0">
                <v-layout row wrap>
                    <v-flex xs9>
                        <h3 class="headline mb-0">{{ title }}</h3>
                    </v-flex>
                    <v-flex xs3 text-xs-right>
                        <v-btn small depressed color="secondary-light" class="ma-0">
                            Add Link
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-card-title>
            <v-card-text>
                <v-expansion-panel class="pa-0 elevation-0">
                    <v-expansion-panel-content class="pa-0 elevation-0 secondary-light">
                        <template v-slot:header>
                            <div class="action-url">
                                {{ host + '/s/a4bey45buvhg8vnxo4y'}}
                            </div>
                        </template>
                        <v-card>
                            <v-card-text class="grey lighten-4">
                                <v-container fluid class="pa-0">
                                    <v-layout row wrap>
                                        <v-flex xs12 sm6 class="pa-2">
                                            <v-text-field
                                                    hide-details
                                                    browser-autocomplete="off"
                                                    label="Token"
                                                    placeholder="a4bey45buvhg8vnxo4y"
                                                    color="secondary-dark"
                                                    v-model="model.AccURL"
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
                                        <v-flex xs12 text-xs-right class="pa-2">
                                            <v-btn small flat color="remove" class="ma-0">
                                                Delete
                                            </v-btn>
                                            <v-btn small depressed dark color="secondary-dark" class="ma-0">
                                                Save
                                            </v-btn>
                                        </v-flex>
                                    </v-layout>
                                </v-container>
                            </v-card-text>
                        </v-card>
                    </v-expansion-panel-content>
                </v-expansion-panel>
            </v-card-text>
            <v-card-actions>
                <v-container fluid text-xs-right class="pt-0 pb-2 pr-2 pl-2">
                    <v-btn @click.stop="upload" depressed color="secondary-light"
                           class="action-webdav">
                        <v-icon left>cloud</v-icon>
                        <span>Upload</span>
                    </v-btn>
                    <v-btn depressed dark color="secondary-dark" @click.stop="confirm"
                           class="action-close">
                        <span>{{ label.confirm }}</span>
                    </v-btn>
                </v-container>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>
<script>
    import Album from "model/album";

    export default {
        name: 'p-share-dialog',
        props: {
            show: Boolean,
            title: String,
        },
        data() {
            return {
                host: window.location.host,
                showPassword: false,
                loading: false,
                search: null,
                model: new Album(),
                label: {
                    url: this.$gettext("Service URL"),
                    user: this.$gettext("Username"),
                    pass: this.$gettext("Password"),
                    cancel: this.$gettext("Cancel"),
                    confirm: this.$gettext("Done"),
                }
            }
        },
        methods: {
            add() {
                this.loading = true;

                this.model.addLink().then((a) => {
                    this.$emit('close');
                }).finally(() => this.loading = false)
            },
            upload() {
                this.$emit('upload');
            },
            close() {
                this.$emit('close');
            },
            confirm() {
                this.$emit('close');
            },
        },
        watch: {
            show: function (show) {
                this.model = this.sele
            }
        },
    }
</script>
