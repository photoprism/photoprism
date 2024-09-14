<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="p-account-add-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pa-2">
        <v-layout row wrap>
          <v-flex xs12 class="pa-2">
            <h3 class="headline pa-0">
              <translate>Add Account</translate>
            </h3>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text class="pb-0 pt-0 px-2">
        <v-layout row wrap>
          <v-flex xs12 class="pa-2">
            <v-text-field v-model="model.AccURL" hide-details autofocus box flat :label="$gettext('Service URL')" placeholder="https://www.example.com/" color="secondary-dark" autocorrect="off" autocapitalize="none"></v-text-field>
          </v-flex>
          <v-flex xs12 sm6 class="pa-2">
            <v-text-field v-model="model.AccUser" hide-details box flat :label="$gettext('Username')" placeholder="optional" color="secondary-dark" autocorrect="off" autocapitalize="none"></v-text-field>
          </v-flex>
          <v-flex xs12 sm6 class="pa-2">
            <v-text-field
              v-model="model.AccPass"
              hide-details
              box
              flat
              browser-autocomplete="new-password"
              autocapitalize="none"
              :label="$gettext('Password')"
              placeholder="optional"
              color="secondary-dark"
              :append-icon="showPassword ? 'visibility' : 'visibility_off'"
              :type="showPassword ? 'text' : 'password'"
              @click:append="showPassword = !showPassword"
            ></v-text-field>
          </v-flex>
        </v-layout>
      </v-card-text>
      <v-card-actions class="pt-1 pb-2 px-2">
        <v-layout row wrap class="pa-2">
          <v-flex xs12 text-xs-left class="caption">
            <translate>Note: Only WebDAV servers, like Nextcloud or PhotoPrism, can be configured as remote service for backup and file upload.</translate>
            <translate>Support for additional services, like Google Drive, will be added over time.</translate>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-2">
            <v-btn depressed color="secondary-light" class="action-cancel ml-2" @click.stop="cancel">
              <span>{{ label.cancel }}</span>
            </v-btn>
            <v-btn depressed dark color="primary-button" class="action-confirm compact mr-0" @click.stop="confirm">
              <span>{{ label.confirm }}</span>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import Service from "model/service";
import * as options from "options/options";

export default {
  name: "PAccountAddDialog",
  props: {
    show: Boolean,
  },
  data() {
    return {
      options: options,
      showPassword: false,
      loading: false,
      search: null,
      model: new Service(false),
      label: {
        cancel: this.$gettext("Cancel"),
        confirm: this.$gettext("Connect"),
      },
    };
  },
  watch: {
    show: function () {},
  },
  methods: {
    cancel() {
      this.$emit("cancel");
    },
    confirm() {
      this.loading = true;

      this.model.save().then((a) => {
        this.loading = false;
        this.$notify.success(this.$gettext("Account created"));
        this.$emit("confirm", a.UID);
      });
    },
  },
};
</script>
