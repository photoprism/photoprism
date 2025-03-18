<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="500"
    class="p-dialog modal-dialog p-settings-password"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" class="form-password" accept-charset="UTF-8" @submit.prevent>
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-lock</v-icon>
          <h6 class="text-h6">{{ $gettext(`Change Password`) }}</h6>
        </v-card-title>
        <v-card-text class="dense">
          <v-row align="start" dense>
            <v-col v-if="oldRequired" cols="12" class="text-caption">
              {{ $gettext(`Please note that changing your password will log you out on other devices and browsers.`) }}
            </v-col>
            <v-col v-if="oldRequired" cols="12">
              <v-text-field
                ref="password"
                v-model="oldPassword"
                :type="showPassword ? 'text' : 'password'"
                :disabled="busy"
                :maxlength="maxLength"
                :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                :label="$gettext('Current Password')"
                tabindex="1"
                hide-details
                autocorrect="off"
                autocapitalize="none"
                autocomplete="current-password"
                class="input-current-password"
                @click:append-inner="showPassword = !showPassword"
              ></v-text-field>
            </v-col>

            <v-col cols="12">
              <v-text-field
                v-model="newPassword"
                :disabled="busy"
                :minlength="minLength"
                :maxlength="maxLength"
                :label="$gettext('New Password')"
                :hint="$gettextInterpolate($gettext('Must have at least %{n} characters.'), { n: minLength })"
                tabindex="2"
                counter
                persistent-hint
                type="password"
                autocorrect="off"
                autocapitalize="none"
                autocomplete="new-password"
                class="input-new-password"
              ></v-text-field>
            </v-col>

            <v-col cols="12">
              <v-text-field
                v-model="confirmPassword"
                :disabled="busy"
                :minlength="minLength"
                :maxlength="maxLength"
                :label="$gettext('Retype Password')"
                :hint="$gettext('Please confirm your new password.')"
                tabindex="3"
                counter
                persistent-hint
                type="password"
                autocorrect="off"
                autocapitalize="none"
                autocomplete="new-password"
                class="input-retype-password"
                @keyup.enter="onConfirm"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn tabindex="5" variant="flat" color="button" class="action-cancel" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn
            tabindex="4"
            variant="flat"
            color="highlight"
            class="action-confirm"
            :disabled="isDisabled()"
            @click.stop="onConfirm"
          >
            {{ $gettext(`Save`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import User from "model/user";

export default {
  name: "PSettingsPassword",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
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
      showPassword: false,
      rtl: this.$isRtl,
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
    afterEnter() {
      this.$view.enter(this, this.$refs?.password);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    isDisabled() {
      return (
        this.isDemo ||
        this.busy ||
        (this.oldPassword === "" && this.oldRequired) ||
        this.newPassword.length < this.minLength ||
        this.newPassword.length > this.maxLength ||
        this.newPassword !== this.confirmPassword
      );
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
