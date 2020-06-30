<template>
  <div class="p-tab p-settings-account">
    <v-form lazy-validation dense
            ref="form" class="form-password" accept-charset="UTF-8">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="px-2 pt-2 pb-4">
              <v-text-field
                      hide-details required
                      :disabled="busy"
                      browser-autocomplete="off"
                      label="Current Password"
                      color="secondary-dark"
                      type="password"
                      placeholder="••••••••"
                      v-model="oldPassword"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-text-field
                      hide-details required
                      :disabled="busy"
                      browser-autocomplete="off"
                      label="New Password"
                      color="secondary-dark"
                      type="password"
                      placeholder="••••••••"
                      v-model="newPassword"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="pa-2">
              <v-text-field
                      hide-details required
                      :disabled="busy"
                      browser-autocomplete="off"
                      label="Confirm Password"
                      color="secondary-dark"
                      type="password"
                      placeholder="••••••••"
                      v-model="confirmPassword"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 text-xs-left class="px-2 pt-4 pb-2">
              <v-btn depressed dark color="secondary-dark" @click.stop="confirm"
                     class="action-confirm ma-0" :disabled="busy">
                <translate>Change Password</translate>
                <v-icon right dark>vpn_key</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>
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
                return (this.oldPassword === "" || this.newPassword === "" || (this.newPassword !== this.confirmPassword));
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
