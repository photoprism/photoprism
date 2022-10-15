<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="p-account-create-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-title primary-title>
        <div>
          <h3 class="headline mx-2 my-0">
            <translate>Add Server</translate>
          </h3>
        </div>
      </v-card-title>
      <v-card-text class="pt-0">
        <v-layout row wrap>
          <v-flex xs12 class="pa-2">
            <v-text-field
                v-model="model.AccURL"
                hide-details autofocus
                :label="$gettext('Service URL')"
                placeholder="https://www.example.com/"
                color="secondary-dark"
                autocorrect="off"
                autocapitalize="none"
            ></v-text-field>
          </v-flex>
          <v-flex xs12 sm6 class="pa-2">
            <v-text-field
                v-model="model.AccUser"
                hide-details
                :label="$gettext('Username')"
                placeholder="optional"
                color="secondary-dark"
                autocorrect="off"
                autocapitalize="none"
            ></v-text-field>
          </v-flex>
          <v-flex xs12 sm6 class="pa-2">
            <v-text-field
                v-model="model.AccPass"
                hide-details
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Password')"
                placeholder="optional"
                color="secondary-dark"
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
            <v-btn depressed color="secondary-light" class="action-cancel mr-2"
                   @click.stop="cancel">
              <span>{{ label.cancel }}</span>
            </v-btn>
            <v-btn depressed dark color="primary-button" class="action-confirm ma-0"
                   @click.stop="confirm">
              <span>{{ label.confirm }}</span>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
<script>
import Service from "model/service";
import * as options from "options/options";

export default {
  name: 'PAccountCreateDialog',
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
      }
    };
  },
  watch: {
    show: function () {
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
};
</script>
