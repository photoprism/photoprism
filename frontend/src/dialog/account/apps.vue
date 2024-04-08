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
              <v-icon v-if="action === 'add'" size="28" color="primary">add</v-icon>
              <v-icon v-else-if="action === 'copy'" size="28" color="primary">password</v-icon>
              <v-icon v-else size="28" color="primary">devices</v-icon>
            </v-flex>
          </v-layout>
        </v-card-title>
        <!-- Confirm -->
        <template v-if="confirmAction !== ''">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-1">
                <translate>Enter your password to confirm the action and continue:</translate>
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
                  @keyup.enter.native="onConfirm"
                ></v-text-field>
              </v-flex>
            </v-layout>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-back ml-0" @click.stop="onBack">
                  <translate>Back</translate>
                </v-btn>
                <v-btn depressed color="primary-button" :disabled="!password || password.length < 4" class="action-confirm white--text compact mr-0" @click.stop="onConfirm">
                  <translate>Continue</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Copy -->
        <template v-else-if="action === 'copy'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-1">
                <translate>Please copy the following randomly generated app password and keep it in a safe place, as you will not be able to see it again:</translate>
              </v-flex>
              <v-flex xs12 class="pa-2">
                <v-text-field
                  v-model="appPassword"
                  type="text"
                  hide-details
                  readonly
                  solo
                  flat
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  browser-autocomplete="off"
                  append-icon="content_copy"
                  class="input-app-password text-selectable"
                  color="secondary-dark"
                  @click:append="onCopyAppPassword"
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
                <v-btn v-if="appPasswordCopied" depressed color="primary-button" :disabled="busy" class="action-done white--text compact mr-0" @click.stop="onDone">
                  <translate>Done</translate>
                </v-btn>
                <v-btn v-else depressed color="primary-button" class="action-copy white--text compact mr-0" @click.stop="onCopyAppPassword">
                  <translate>Copy</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Add -->
        <template v-else-if="action === 'add'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-1">
                <translate>To generate a new app-specific password, please enter the name and authorization scope of the application and select an expiration date:</translate>
              </v-flex>
              <v-flex xs12 class="pa-2">
                <v-text-field
                  v-model="app.client_name"
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
                <v-select v-model="app.scope" hide-details box :disabled="busy" :items="auth.ScopeOptions()" :label="$gettext('Scope')" :menu-props="{ maxHeight: 346 }" color="secondary-dark" background-color="secondary-light" class="input-scope"></v-select>
              </v-flex>
              <v-flex xs12 sm6 class="pa-2">
                <v-select v-model="app.expires_in" :disabled="busy" :label="$gettext('Expires')" browser-autocomplete="off" hide-details box flat color="secondary-dark" class="input-expires" item-text="text" item-value="value" :items="options.Expires()"></v-select>
              </v-flex>
            </v-layout>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-cancel ml-0" @click.stop="onCancel">
                  <translate>Cancel</translate>
                </v-btn>
                <v-btn depressed color="primary-button" :disabled="app.client_name === '' || app.scope === ''" class="action-generate white--text compact mr-0" @click.stop="onGenerate">
                  <translate>Generate</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Apps -->
        <template v-else>
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
            </v-layout>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-close ml-0" @click.stop="close">
                  <translate>Close</translate>
                </v-btn>
                <v-btn depressed color="primary-button" class="action-add white--text compact mr-0" @click.stop="onAdd">
                  <translate>Add</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
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
      action: "",
      confirmAction: "",
      user: this.$session.getUser(),
      apps: [],
      app: {
        client_name: "",
        scope: "*",
        expires_in: 0,
      },
      appPassword: "",
      appPasswordCopied: false,
    };
  },
  watch: {
    show: function (show) {
      if (show) {
        this.reset();
        this.find();
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
    onCopyAppPassword() {
      this.copyText(this.appPassword);
      this.appPasswordCopied = true;
    },
    reset(action) {
      if (!action) {
        action = "apps";
      }

      this.app = {
        client_name: "",
        scope: "*",
        expires_in: 0,
      };

      this.action = action;
      this.confirmAction = "";
      this.appPasswordCopied = false;

      console.log("reset", this.action, action);
    },
    onConfirm() {
      if (this.busy) {
        return;
      }

      switch (this.confirmAction) {
        case "onGenerate":
          this.onGenerate();
      }
    },
    onDone() {
      if (this.busy) {
        return;
      }

      this.appPassword = "";
      this.reset();
    },
    onCancel() {
      if (this.busy) {
        return;
      }

      this.reset();
    },
    onBack() {
      if (this.busy) {
        return;
      }

      this.confirmAction = "";
    },
    onAdd() {
      if (this.busy) {
        return;
      }

      this.action = "add";
      this.confirmAction = "";
    },
    onGenerate() {
      if (this.busy) {
        return;
      }

      if (this.confirmAction === "") {
        this.confirmAction = "onGenerate";
        return;
      }

      this.busy = true;
      this.$session
        .createApp(this.app.client_name, this.app.scope, this.app.expires_in, this.password)
        .then((app) => {
          this.appPassword = app.access_token;
          this.reset("copy");
        })
        .catch(() => {
          this.action = "add";
          this.confirmAction = "";
        })
        .finally(() => {
          this.busy = false;
        });
    },
    find() {
      this.$notify.blockUI();
      this.model
        .findApps()
        .then((resp) => {
          console.log("findApps", resp);
          this.apps = resp;
        })
        .finally(() => {
          this.$notify.unblockUI();
        });
    },
    close() {
      if (this.busy) {
        return;
      }

      this.appPassword = "";

      this.$emit("close");
    },
  },
};
</script>
