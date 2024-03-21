<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="modal-dialog p-account-password-dialog" @keydown.esc="close">
    <v-form ref="form" dense class="form-password" accept-charset="UTF-8" @submit.prevent>
      <v-card raised elevation="24">
        <v-card-title primary-title class="pa-2">
          <v-layout row wrap class="pa-2">
            <v-flex xs9 class="text-xs-left">
              <h3 class="headline pa-0">
                <translate>Change Password</translate>
              </h3>
            </v-flex>
            <v-flex xs3 class="text-xs-right">
              <v-icon size="28" color="primary">lock</v-icon>
            </v-flex>
          </v-layout>
        </v-card-title>
        <v-card-text class="py-0 px-2">
          <v-layout wrap align-top>
            <v-flex v-if="oldRequired" xs12 class="px-2 pb-2 caption">
              <translate>Please note that changing your password will log you out on other devices and browsers.</translate>
            </v-flex>
            <v-flex v-if="oldRequired" xs12 class="px-2 py-1">
              <v-text-field
                v-model="oldPassword"
                hide-details
                required
                box
                flat
                type="password"
                :disabled="busy"
                :maxlength="maxLength"
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Current Password')"
                class="input-current-password"
                color="secondary-dark"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="px-2 py-1">
              <v-text-field
                v-model="newPassword"
                required
                counter
                persistent-hint
                box
                flat
                type="password"
                :disabled="busy"
                :minlength="minLength"
                :maxlength="maxLength"
                browser-autocomplete="new-password"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('New Password')"
                class="input-new-password"
                color="secondary-dark"
                :hint="$gettextInterpolate($gettext('Must have at least %{n} characters.'), { n: minLength })"
              ></v-text-field>
            </v-flex>

            <v-flex xs12 class="px-2 py-1">
              <v-text-field
                v-model="confirmPassword"
                required
                counter
                persistent-hint
                box
                flat
                type="password"
                :disabled="busy"
                :minlength="minLength"
                :maxlength="maxLength"
                browser-autocomplete="new-password"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Retype Password')"
                class="input-retype-password"
                color="secondary-dark"
                :hint="$gettext('Please confirm your new password.')"
                @keyup.enter.native="onConfirm"
              ></v-text-field>
            </v-flex>
          </v-layout>
        </v-card-text>
        <v-card-actions class="pt-1 pb-2 px-2">
          <v-layout row wrap class="pa-2">
            <v-flex xs12 text-xs-right>
              <v-btn depressed color="secondary-light" class="action-cancel ml-0" @click.stop="close">
                <translate>Cancel</translate>
              </v-btn>
              <v-btn depressed color="primary-button" class="action-confirm white--text compact mr-0" :disabled="isDisabled()" @click.stop="onConfirm">
                <translate>Save</translate>
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

export default {
  name: "PAccountPasswordDialog",
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
      oldPassword: "",
      newPassword: "",
      confirmPassword: "",
      minLength: this.$config.get("passwordLength"),
      maxLength: 72,
      rtl: this.$rtl,
    };
  },
  computed: {
    oldRequired() {
      if (!this.model) {
        return true;
      }

      const sessionUser = this.$session.getUser();

      return !sessionUser.SuperAdmin || this.model.getId() === sessionUser.getId();
    },
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$emit("close");
    }
  },
  methods: {
    isDisabled() {
      return this.isDemo || this.busy || (this.oldPassword === "" && this.oldRequired) || this.newPassword.length < this.minLength || this.newPassword.length > this.maxLength || this.newPassword !== this.confirmPassword;
    },
    onConfirm() {
      this.busy = true;
      this.model
        .changePassword(this.oldPassword, this.newPassword)
        .then(() => {
          this.$notify.success(this.$gettext("Password changed"));
          this.$emit("close");
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
