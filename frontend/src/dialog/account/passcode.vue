<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="modal-dialog p-account-passcode-dialog" @keydown.esc="close">
    <v-form ref="form" lazy-validation dense accept-charset="UTF-8" class="form-password" @submit.prevent>
      <v-card raised elevation="24">
        <v-card-title primary-title class="pa-2">
          <v-layout row wrap class="pa-2">
            <v-flex xs10 class="text-xs-left">
              <h3 class="headline pa-0">
                <translate>2-Factor Authentication</translate>
              </h3>
            </v-flex>
            <v-flex xs2 class="text-xs-right">
              <v-icon v-if="page === 'deactivate'" size="28" color="primary">verified_user</v-icon>
              <v-icon v-else size="28" color="primary">vpn_key</v-icon>
            </v-flex>
          </v-layout>
        </v-card-title>
        <!-- Setup -->
        <template v-if="page === 'setup'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-2">
                <translate>After entering your password for confirmation, you can set up two-factor authentication with a compatible authenticator app or device:</translate>
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
            <v-flex xs12 class="pa-2 body-1">
              <translate>Enabling two-factor authentication means that you will need a randomly generated verification code to log in, so even if someone gains access to your password, they will not be able to access your account.</translate>
            </v-flex>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-close ml-0" @click.stop="close">
                  <translate>Close</translate>
                </v-btn>
                <v-btn depressed color="primary-button" class="action-setup white--text compact mr-0" :disabled="setupDisabled()" @click.stop="onSetup">
                  <translate>Setup</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Confirm -->
        <template v-else-if="page === 'confirm'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-1">
                <translate>Scan the QR code with your authenticator app or use the setup key shown below and then enter the generated passcode for verification:</translate>
              </v-flex>
              <v-flex xs12 class="pa-2">
                <img :src="key.QRCode" class="width-100" alt="QR Code" />
              </v-flex>
            </v-layout>
            <v-flex xs12 class="pa-2 subheading text-xs-center">
              <pre class="clickable" @click.stop.prevent="copyText(key.Secret)">{{ key.Secret }}</pre>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                v-model="passcode"
                :disabled="busy"
                name="passcode"
                type="text"
                :label="$gettext('Verification Code')"
                mask="### ###"
                pattern="[0-9]*"
                inputmode="numeric"
                hide-details
                required
                solo
                flat
                autocorrect="off"
                autocapitalize="none"
                autocomplete="one-time-code"
                browser-autocomplete="one-time-code"
                class="input-passcode"
                color="secondary-dark"
                prepend-inner-icon="verified_user"
                @keyup.enter.native="onConfirm"
              ></v-text-field>
            </v-flex>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-cancel ml-0" @click.stop="close">
                  <translate>Cancel</translate>
                </v-btn>
                <v-btn depressed color="primary-button" class="action-confirm white--text compact mr-0" :disabled="passcode.length !== 6" @click.stop="onConfirm">
                  <translate>Confirm</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Activate -->
        <template v-else-if="page === 'activate'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-2">
                <translate>Use the following recovery code to access your account when you are unable to generate a valid passcode with your authenticator app or device:</translate>
              </v-flex>
              <v-flex xs12 class="pa-2">
                <v-text-field
                  v-model="key.RecoveryCode"
                  type="text"
                  mask="nnn nnn nnn nnn"
                  hide-details
                  readonly
                  solo
                  flat
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  browser-autocomplete="off"
                  append-icon="content_copy"
                  class="input-recoverycode"
                  color="secondary-dark"
                  @click:append="onCopyRecoveryCode"
                ></v-text-field>
              </v-flex>
              <v-flex xs12 class="pa-2 body-1">
                <translate>To avoid being locked out of your account, please download, print or copy this recovery code now and keep it in a safe place.</translate>
                <translate>It is a one-time use code that will disable 2FA for your account when you use it.</translate>
              </v-flex>
            </v-layout>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-cancel ml-0" @click.stop="close">
                  <translate>Cancel</translate>
                </v-btn>
                <v-btn v-if="recoveryCodeCopied" depressed color="primary-button" class="action-activate white--text compact mr-0" @click.stop="onActivate">
                  <translate>Activate</translate>
                </v-btn>
                <v-btn v-else depressed color="primary-button" class="action-copy white--text compact mr-0" @click.stop="onCopyRecoveryCode">
                  <translate>Copy</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Deactivate -->
        <template v-else-if="page === 'deactivate'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-2">
                <translate>Two-factor authentication has been enabled for your account.</translate>
              </v-flex>
              <v-flex xs12 class="pa-2 body-1">
                <translate>If you lose access to your authenticator app or device, you can use your recovery code to regain access to your account.</translate>
                <translate>It is a one-time use code that will disable 2FA for your account when you use it.</translate>
              </v-flex>
              <v-flex xs12 class="pa-2 body-1">
                <translate>To switch to a new authenticator app or device, first deactivate two-factor authentication and then reactivate it:</translate>
              </v-flex>
              <v-flex xs12 class="pa-2">
                <v-text-field
                  v-model="password"
                  :disabled="busy"
                  name="password"
                  :type="showPassword ? 'text' : 'password'"
                  hide-details
                  required
                  solo
                  flat
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="current-password"
                  browser-autocomplete="current-password"
                  :label="$gettext('Password')"
                  class="input-password text-selectable"
                  :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                  prepend-inner-icon="lock"
                  color="secondary-dark"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter.native="onDeactivate"
                ></v-text-field>
              </v-flex>
            </v-layout>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="primary-button" class="action-deactivate white--text compact ml-0" :disabled="setupDisabled()" @click.stop="onDeactivate">
                  <translate>Deactivate</translate>
                </v-btn>
                <v-btn depressed color="secondary-light" class="action-close mr-0" @click.stop="close">
                  <translate>Close</translate>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>
        </template>
        <!-- Not Available -->
        <template v-else-if="page === 'not_available'">
          <v-card-text class="py-0 px-2">
            <v-layout wrap align-top>
              <v-flex xs12 class="pa-2 body-2">
                <translate>Only locally managed accounts can be set up for authentication with 2FA.</translate>
              </v-flex>
            </v-layout>
          </v-card-text>
          <v-card-actions class="pa-2">
            <v-layout row wrap class="pa-2">
              <v-flex xs12 text-xs-right>
                <v-btn depressed color="secondary-light" class="action-close mr-0" @click.stop="close">
                  <translate>Close</translate>
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
import Util from "common/util";

export default {
  name: "PAccountPasscodeDialog",
  props: {
    show: Boolean,
    model: {
      type: Object,
      default: () => this.$session.getUser(),
    },
  },
  data() {
    return {
      busy: false,
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      passcode: "",
      password: "",
      recoveryCodeCopied: false,
      showPassword: false,
      minLength: this.$config.get("passwordLength"),
      maxLength: 72,
      rtl: this.$rtl,
      key: {},
    };
  },
  computed: {
    page() {
      if (this.model?.AuthProvider !== "default" && this.model?.AuthProvider !== "local" && this.model?.AuthProvider !== "ldap") {
        return "not_available";
      } else if (this.model?.AuthMethod === "2fa") {
        return "deactivate";
      } else if (this.key?.Type === "totp") {
        if (!this.key?.VerifiedAt) {
          return "confirm";
        } else if (!this.key?.ActivatedAt) {
          return "activate";
        } else {
          return "deactivate";
        }
      }

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
      this.passcode = "";
      this.password = "";
      this.showPassword = false;
      this.recoveryCodeCopied = false;
      this.updateUser();
    },
    updateUser() {
      this.$emit("updateUser");
    },
    setupDisabled() {
      return this.isDemo || this.busy || this.password.length < this.minLength;
    },
    onSetup() {
      if (this.busy || this.password === "") {
        return;
      }
      this.busy = true;
      this.model
        .createPasscode(this.password)
        .then((resp) => {
          this.key = resp;
        })
        .finally(() => {
          this.busy = false;
        });
    },
    onConfirm() {
      if (this.busy || this.passcode === "") {
        return;
      }
      this.busy = true;
      this.model
        .confirmPasscode(this.passcode)
        .then((resp) => {
          this.key = resp;
          this.$notify.success(this.$gettext("Successfully verified"));
        })
        .finally(() => {
          this.busy = false;
          this.passcode = "";
          this.password = "";
          this.showPassword = false;
          this.recoveryCodeCopied = false;
        });
    },
    onCopyRecoveryCode() {
      this.copyText(this.key.RecoveryCode);
      this.recoveryCodeCopied = true;
    },
    onActivate() {
      if (this.busy) {
        return;
      }

      this.busy = true;
      this.model
        .activatePasscode()
        .then((resp) => {
          this.key = resp;
          this.$notify.success(this.$gettext("Successfully activated"));
        })
        .finally(() => {
          this.busy = false;
          this.reset();
        });
    },
    onDeactivate() {
      if (this.busy || this.password === "") {
        return;
      }
      this.busy = true;
      this.model
        .deactivatePasscode(this.password)
        .then(() => {
          this.$notify.success(this.$gettext("Settings saved"));
          this.reset();
          this.key = {};
        })
        .finally(() => {
          this.busy = false;
        });
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
