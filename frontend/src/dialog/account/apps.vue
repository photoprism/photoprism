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
              <translate>To generate a new app-specific password, please enter the name and authorization scope of the application and choose an expiration date:</translate>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                v-model="newApp.client_name"
                :disabled="busy"
                name="client_name"
                type="text"
                :label="$gettext('Name')"
                required
                autofocus
                hide-details
                box
                autocorrect="off"
                autocapitalize="none"
                autocomplete="off"
                browser-autocomplete="off"
                class="input-name text-selectable"
                color="secondary-dark"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-select v-model="newApp.scope" hide-details box :disabled="busy" :items="auth.ScopeOptions()" :label="$gettext('Scope')" :menu-props="{ maxHeight: 346 }" color="secondary-dark" background-color="secondary-light" class="input-scope"></v-select>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-select v-model="newApp.lifetime" :disabled="busy" :label="$gettext('Expires')" browser-autocomplete="off" hide-details box flat color="secondary-dark" class="input-expires" item-text="text" item-value="value" :items="options.Expires()"></v-select>
            </v-flex>
          </v-layout>
        </v-card-text>
        <v-card-actions class="pa-2">
          <v-layout row wrap class="pa-2">
            <v-flex xs12 text-xs-right>
              <v-btn depressed color="secondary-light" class="action-close ml-0" @click.stop="close">
                <translate>Close</translate>
              </v-btn>
              <v-btn depressed color="primary-button" disabled class="action-generate white--text compact mr-0" @click.stop="close">
                <translate>Generate</translate>
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
import Util from "common/util";
import * as auth from "options/auth";
import * as options from "options/options";

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
      auth,
      options,
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
      newApp: {
        grant_type: "session",
        password: "",
        client_name: "",
        scope: "*",
        lifetime: 0,
      },
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
