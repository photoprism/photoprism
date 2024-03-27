<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="modal-dialog p-account-apps-dialog" @keydown.esc="close">
    <v-form ref="form" lazy-validation dense class="form-password" accept-charset="UTF-8" @submit.prevent>
      <v-card raised elevation="24">
        <v-card-title primary-title class="pa-2">
          <v-layout row wrap class="pa-2">
            <v-flex xs9 class="text-xs-left">
              <h3 class="headline pa-0">
                <translate>Apps and Devices</translate>
              </h3>
            </v-flex>
            <v-flex xs3 class="text-xs-right">
              <v-icon size="28" color="primary">add</v-icon>
            </v-flex>
          </v-layout>
        </v-card-title>
        <!-- Setup -->
        <v-card-text class="py-0 px-2">
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2 body-2">
              <translate>To create a new app-specific password, please enter the name and type of the application and select an expiration date:</translate>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                :disabled="busy"
                name="appname"
                type="text"
                :label="$gettext('App Name')"
                hide-details
                required
                autofocus
                solo
                flat
                autocorrect="off"
                autocapitalize="none"
                autocomplete="off"
                browser-autocomplete="off"
                prepend-inner-icon="devices"
                class="input-appname text-selectable"
                color="secondary-dark"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                v-model="password"
                :disabled="busy"
                name="password"
                :type="showPassword ? 'text' : 'password'"
                :label="$gettext('Password')"
                hide-details
                required
                autofocus
                solo
                flat
                autocorrect="off"
                autocapitalize="none"
                autocomplete="current-password"
                browser-autocomplete="current-password"
                class="input-password text-selectable"
                :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                prepend-inner-icon="lock"
                color="secondary-dark"
                @click:append="showPassword = !showPassword"
                @keyup.enter.native="onSetup"
              ></v-text-field>
            </v-flex>
          </v-layout>
        </v-card-text>
        <v-card-actions class="pa-2">
          <v-layout row wrap class="pa-2">
            <v-flex xs12 text-xs-right>
              <v-btn depressed color="secondary-light" class="action-close ml-0" @click.stop="close">
                <translate>Close</translate>
              </v-btn>
              <v-btn depressed color="primary-button" class="action-confirm white--text compact mr-0" @click.stop="close">
                <translate>Create</translate>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import User from "model/user";
import Util from "../../common/util";

export default {
  name: "PAccountAppsDialog",
  props: {
    show: Boolean,
    model: {
      type: Object,
      default: () => new User(null),
    },
  },
  data() {
    return {
      busy: false,
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      password: "",
      showPassword: false,
      minLength: this.$config.get("passwordLength"),
      maxLength: 72,
      rtl: this.$rtl,
      passwords: [],
      user: this.$session.getUser(),
    };
  },
  computed: {
    page() {
      return "setup";
    },
  },
  watch: {
    show: function (show) {
      if (show) {
        this.reset();
      }
    },
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$emit("close");
    }
  },
  methods: {
    async copyText(text) {
      if (!text) {
        return;
      }

      try {
        await Util.copyToMachineClipboard(text);
        this.$notify.success(this.$gettext("Copied to clipboard"));
      } catch (error) {
        this.$notify.error(this.$gettext("Failed copying to clipboard"));
      }
    },
    reset() {
      this.password = "";
      this.showPassword = false;
      this.updateUser();
    },
    updateUser() {
      this.$notify.blockUI();
      this.$session
        .refresh()
        .then(() => {
          this.user = this.$session.getUser();
        })
        .finally(() => {
          this.$notify.unblockUI();
        });
    },
    confirm() {
      this.close();
    },
    close() {
      if (this.busy) {
        return;
      }

      this.$emit("close");
    },
  },
};
</script>
