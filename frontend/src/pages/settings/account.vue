<template>
  <div class="p-tab p-settings-account">
    <v-form dense ref="form" class="form-password" accept-charset="UTF-8">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <v-text-field
                      hide-details required
                      :disabled="busy"
                      browser-autocomplete="off"
                      :label="$gettext('Current Password')"
                      color="secondary-dark"
                      type="password"
                      placeholder="••••••••"
                      v-model="oldPassword"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-text-field
                      required counter persistent-hint
                      :disabled="busy"
                      browser-autocomplete="off"
                      :label="$gettext('New Password')"
                      color="secondary-dark"
                      type="password"
                      placeholder="••••••••"
                      v-model="newPassword"
                      :hint="$gettext('At least 6 characters.')"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-text-field
                      required counter persistent-hint
                      :disabled="busy"
                      browser-autocomplete="off"
                      :label="$gettext('Retype Password')"
                      color="secondary-dark"
                      type="password"
                      placeholder="••••••••"
                      v-model="confirmPassword"
                      :hint="$gettext('Please confirm your new password.')"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <p class="caption pa-0">
                <translate>Note: Updating the password will not revoke access from already authenticated users.</translate>
              </p>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-btn depressed color="secondary-dark"
                     @click.stop="confirm"
                     :disabled="disabled()"
                     class="action-confirm white--text ma-0">
                <translate>Change</translate>
                <v-icon right dark>vpn_key</v-icon>
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
        name: 'p-settings-account',
        data() {
            return {
                oldPassword: "",
                newPassword: "",
                confirmPassword: "",
                busy: false,
            };
        },
        methods: {
            disabled() {
                return (this.busy || this.oldPassword === "" || this.newPassword.length < 6 || (this.newPassword !== this.confirmPassword));
            },
            confirm() {
                this.busy = true;
                this.$session.getUser().changePassword(this.oldPassword, this.newPassword).then(() => {
                    this.$notify.success(this.$gettext("Password changed"));
                }).finally(() => this.busy = false)
            },
        },
    };
</script>
