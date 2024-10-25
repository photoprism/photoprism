<template>
  <v-dialog :model-value="show" persistent max-width="500" class="p-account-add-dialog" @keydown.esc="cancel">
    <v-card elevation="24">
      <v-card-title class="pa-2">
        <v-row>
          <v-col cols="12" class="pa-2">
            <h3 class="text-h5 pa-0">
              <translate>Add Account</translate>
            </h3>
          </v-col>
        </v-row>
      </v-card-title>
      <v-card-text class="pb-0 pt-0 px-2">
        <v-row>
          <v-col cols="12" class="pa-2">
            <v-text-field v-model="model.AccURL" hide-details autofocus variant="filled" flat :label="$gettext('Service URL')" placeholder="https://www.example.com/" color="secondary-dark" autocorrect="off" autocapitalize="none"></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-text-field v-model="model.AccUser" hide-details variant="filled" flat :label="$gettext('Username')" placeholder="optional" color="secondary-dark" autocorrect="off" autocapitalize="none"></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-text-field
              v-model="model.AccPass"
              hide-details
              variant="filled"
              flat
              autocomplete="new-password"
              autocapitalize="none"
              :label="$gettext('Password')"
              placeholder="optional"
              color="secondary-dark"
              :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showPassword ? 'text' : 'password'"
              @click:append="showPassword = !showPassword"
            ></v-text-field>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions class="pt-1 pb-2 px-2">
        <v-row class="pa-2">
          <v-col cols="12" class="text-left text-caption">
            <translate>Note: Only WebDAV servers, like Nextcloud or PhotoPrism, can be configured as remote service for backup and file upload.</translate>
            <translate>Support for additional services, like Google Drive, will be added over time.</translate>
          </v-col>
          <v-col cols="12" class="text-right pt-2">
            <v-btn variant="flat" color="secondary-light" class="action-cancel ml-2" @click.stop="cancel">
              <span>{{ label.cancel }}</span>
            </v-btn>
            <v-btn variant="flat" theme="dark" color="primary-button" class="action-confirm compact mr-0" @click.stop="confirm">
              <span>{{ label.confirm }}</span>
            </v-btn>
          </v-col>
        </v-row>
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
