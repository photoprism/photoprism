<template>
  <v-dialog :model-value="show" persistent max-width="500" class="modal-dialog p-account-password-dialog" @keydown.esc="close">
    <v-form ref="form" class="form-password" accept-charset="UTF-8" @submit.prevent>
      <v-card elevation="24">
        <v-card-title>
          <v-row class="pa-2">
            <v-col cols="9" class="text-left">
              <h3 class="text-h5 pa-0">
                <translate>Change Password</translate>
              </h3>
            </v-col>
            <v-col cols="3" class="text-right">
              <v-icon size="28" color="primary">mdi-lock</v-icon>
            </v-col>
          </v-row>
        </v-card-title>
        <v-card-text class="py-0">
          <v-row align="start">
            <v-col v-if="oldRequired" cols="12" class="pb-2 text-caption">
              <translate>Please note that changing your password will log you out on other devices and browsers.</translate>
            </v-col>
            <v-col v-if="oldRequired" cols="12" class="py-1">
              <v-text-field
                v-model="oldPassword"
                hide-details
                required
                variant="solo"
                flat
                type="password"
                autocorrect="off"
                autocapitalize="none"
                autocomplete="current-password"
                :disabled="busy"
                :maxlength="maxLength"
                :label="$gettext('Current Password')"
                class="input-current-password"
                color="surface-variant"
              ></v-text-field>
            </v-col>

            <v-col cols="12" class="py-1">
              <v-text-field
                v-model="newPassword"
                required
                counter
                persistent-hint
                variant="solo"
                flat
                type="password"
                :disabled="busy"
                :minlength="minLength"
                :maxlength="maxLength"
                autocorrect="off"
                autocapitalize="none"
                autocomplete="new-password"
                :label="$gettext('New Password')"
                class="input-new-password"
                color="surface-variant"
                :hint="$gettextInterpolate($gettext('Must have at least %{n} characters.'), { n: minLength })"
              ></v-text-field>
            </v-col>

            <v-col cols="12" class="py-1">
              <v-text-field
                v-model="confirmPassword"
                required
                counter
                persistent-hint
                variant="solo"
                flat
                type="password"
                :disabled="busy"
                :minlength="minLength"
                :maxlength="maxLength"
                autocorrect="off"
                autocapitalize="none"
                autocomplete="new-password"
                :label="$gettext('Retype Password')"
                class="input-retype-password"
                color="surface-variant"
                :hint="$gettext('Please confirm your new password.')"
                @keyup.enter="onConfirm"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="pt-1 pb-2">
          <v-row class="pa-2">
            <v-col cols="12" class="text-right">
              <v-btn variant="flat" color="secondary-light" class="action-cancel ml-0" @click.stop="close">
                <translate>Cancel</translate>
              </v-btn>
              <v-btn variant="flat" color="primary-button" class="action-confirm compact mr-0" :disabled="isDisabled()" @click.stop="onConfirm">
                <translate>Save</translate>
              </v-btn>
            </v-col>
          </v-row>
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
