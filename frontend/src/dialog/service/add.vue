<template>
  <v-dialog :model-value="show" persistent max-width="500" class="p-account-add-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-swap-horizontal</v-icon>
        <h6 class="text-h6">
          <translate>Add Account</translate>
        </h6>
      </v-card-title>
      <v-card-text class="dense">
        <v-row dense>
          <v-col cols="12">
            <v-text-field v-model="model.AccURL" hide-details autofocus :label="$gettext('Service URL')" placeholder="https://www.example.com/" autocorrect="off" autocapitalize="none"></v-text-field>
          </v-col>
          <v-col cols="12" sm="6">
            <v-text-field v-model="model.AccUser" hide-details :label="$gettext('Username')" placeholder="optional" autocorrect="off" autocapitalize="none"></v-text-field>
          </v-col>
          <v-col cols="12" sm="6">
            <v-text-field
              v-model="model.AccPass"
              hide-details
              autocomplete="new-password"
              autocapitalize="none"
              :label="$gettext('Password')"
              placeholder="optional"
              :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showPassword ? 'text' : 'password'"
              @click:append-inner="showPassword = !showPassword"
            ></v-text-field>
          </v-col>
          <v-col cols="12" class="text-left text-caption">
            <translate>Note: Only WebDAV servers, like Nextcloud or PhotoPrism, can be configured as remote service for backup and file upload.</translate>
            <translate>Support for additional services, like Google Drive, will be added over time.</translate>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <span>{{ label.cancel }}</span>
        </v-btn>
        <v-btn variant="flat" color="highlight" class="action-confirm" @click.stop="confirm">
          <span>{{ label.confirm }}</span>
        </v-btn>
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
