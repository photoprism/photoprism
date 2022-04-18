<template>
  <div class="p-tab p-settings-account">
    <v-form ref="form" dense class="form-password" accept-charset="UTF-8">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="oldPassword"
                  hide-details required
                  type="password"
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Current Password')"
                  class="input-current-password"
                  color="secondary-dark"
                  placeholder="••••••••"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="newPassword"
                  required counter persistent-hint
                  type="password"
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('New Password')"
                  class="input-new-password"
                  color="secondary-dark"
                  placeholder="••••••••"
                  :hint="$gettext('Must have at least 8 characters.')"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="confirmPassword"
                  required counter persistent-hint
                  type="password"
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Retype Password')"
                  class="input-retype-password"
                  color="secondary-dark"
                  placeholder="••••••••"
                  :hint="$gettext('Please confirm your new password.')"
                  @keyup.enter.native="confirm"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <p class="caption pa-0">
                <translate>Note: Updating the password will not revoke access from already authenticated users.</translate>
              </p>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-btn depressed color="primary-button"
                     :disabled="disabled()"
                     class="action-confirm white--text ma-0"
                     @click.stop="confirm">
                <translate>Change</translate>
                <v-icon :right="!rtl" :left="rtl" dark>keyboard_return</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>

export default {
  name: 'PSettingsAccount',
  data() {
    const isDemo = this.$config.get("demo");

    return {
      demo: isDemo,
      oldPassword: "",
      newPassword: "",
      confirmPassword: "",
      busy: false,
      rtl: this.$rtl,
    };
  },
  methods: {
    disabled() {
      return (this.demo || this.busy || this.oldPassword === "" || this.newPassword.length < 8 || (this.newPassword !== this.confirmPassword));
    },
    confirm() {
      this.busy = true;
      this.$session.getUser().changePassword(this.oldPassword, this.newPassword).then(() => {
        this.$notify.success(this.$gettext("Password changed"));
      }).finally(() => this.busy = false);
    },
  },
};
</script>
