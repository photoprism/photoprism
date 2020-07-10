<template>
  <v-dialog lazy v-model="show" persistent max-width="500" class="p-account-create-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-title primary-title>
        <div>
          <h3 class="headline mx-2 my-0"><translate>Add Server</translate></h3>
        </div>
      </v-card-title>
      <v-card-text class="pt-0">
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
          <v-flex xs12 text-xs-left class="pa-2 caption">
            <translate>Note: Only WebDAV servers, like Nextcloud or PhotoPrism, can be configured as remote service for backup and file upload.</translate>
            <translate>Support for additional services, like Google Drive, will be added over time.</translate>
          </v-flex>
          <v-flex xs12 text-xs-right class="px-2 pt-2 pb-0">
            <v-btn @click.stop="cancel" depressed color="secondary-light"
                   class="action-cancel mr-2">
              <span>{{ label.cancel }}</span>
            </v-btn>
            <v-btn depressed dark color="secondary-dark" @click.stop="confirm"
                   class="action-confirm ma-0">
              <span>{{ label.confirm }}</span>
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
                    this.$emit('confirm', a.UID);
                });
            },
        },
        watch: {
            show: function (show) {
            }
        },
    }
</script>
