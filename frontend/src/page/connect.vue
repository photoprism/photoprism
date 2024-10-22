<template>
  <div class="p-page p-page-upgrade">
    <v-toolbar flat color="secondary" :dense="$vuetify.display.smAndDown">
      <v-toolbar-title>
        <translate>Membership</translate>
        <v-icon v-if="rtl">mdi-chevron-left</v-icon>
        <v-icon v-else>mdi-chevron-right</v-icon>
        <span v-if="busy">
          <translate>Busy, please wait…</translate>
        </span>
        <span v-else-if="success">
          <translate>Successfully Connected</translate>
        </span>
        <span v-else-if="error">
          <translate>Error</translate>
        </span>
        <span v-else>
          <translate>Upgrade</translate>
        </span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://link.photoprism.app/personal-editions" target="_blank" class="action-upgrade" :title="$gettext('Learn more')">
        <v-icon size="26" color="secondary-dark" v-html="'$vuetify.icons.prism'"></v-icon>
      </v-btn>
    </v-toolbar>
    <v-form ref="form" v-model="valid" autocomplete="off" class="px-6 pt-6 pb-0" lazy-validation @submit.prevent>
      <v-row v-if="busy">
        <v-col cols="12" class="d-flex text-sm-center pa-2">
          <v-progress-linear color="secondary-dark flex-grow-1" :indeterminate="true"></v-progress-linear>
        </v-col>
      </v-row>
      <v-row v-else-if="error">
        <v-col cols="12" class="text-sm-left pa-2">
          <!-- TODO: change this icon -->
          <v-alert color="error" icon="gpp_bad" class="mt-6" variant="outlined">
            {{ error }}
          </v-alert>
        </v-col>
        <v-col cols="12" class="pa-2">
          <v-btn color="primary-button lighten-2" :block="$vuetify.display.xs" class="ml-0" variant="outlined" :disabled="busy" @click.stop="reset">
            <translate>Cancel</translate>
          </v-btn>
          <v-btn color="primary-button" :block="$vuetify.display.xs" class="text-white ml-0" href="https://www.photoprism.app/contact" target="_blank" variant="flat">
            <translate>Contact Us</translate>
          </v-btn>
        </v-col>
      </v-row>
      <v-row v-else-if="success">
        <v-col cols="12" class="pa-2 d-flex">
          <p class="text-subtitle-1 text-left flex-grow-1">
            <translate>Your account has been successfully connected.</translate>
            <span v-if="$config.values.restart">
              <translate>Please restart your instance for the changes to take effect.</translate>
            </span>
          </p>
        </v-col>
        <v-col cols="12" class="d-flex grow pa-2">
          <v-btn href="https://my.photoprism.app/dashboard" target="_blank" color="primary-button lighten-2 flex-grow-1" :block="$vuetify.display.xs" class="ml-0" variant="outlined" :disabled="busy">
            <translate>Manage Account</translate>
          </v-btn>
          <v-btn v-if="$config.values.restart && !$config.values.disable.restart" color="primary-button" :block="$vuetify.display.xs" class="text-white ml-0" variant="flat" :disabled="busy" @click.stop.p.prevent="onRestart">
            <translate>Restart</translate>
            <!-- TODO: change this icon -->
            <v-icon :right="!rtl" :left="rtl" dark>restart_alt</v-icon>
          </v-btn>
          <v-btn v-if="$config.getTier() < 4" href="https://my.photoprism.app/dashboard/membership" target="_blank" color="primary-button" :block="$vuetify.display.xs" class="text-white ml-0" variant="flat" :disabled="busy">
            <translate>Upgrade Now</translate>
            <v-icon v-if="rtl" left dark>mdi-chevron-left</v-icon>
            <v-icon v-else right dark>mdi-chevron-right</v-icon>
          </v-btn>
        </v-col>
      </v-row>
      <v-row v-else>
        <v-col v-if="$config.getTier() < 4" cols="12" class="d-flex align-center justify-center px-2 pt-1 pb-6 text-subtitle-1 text-selectable">
          <translate>Become a member today, support our mission and enjoy our member benefits!</translate>
          <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
        </v-col>
        <v-col cols="12" class="grow align-center justify-center px-2 py-1">
          <v-alert color="secondary-dark" variant="outlined">
            <p class="text-subtitle-1 text-selectable">
              <strong><translate>To upgrade, you can either enter an activation code or click "Register" to sign up on our website:</translate></strong>
            </p>
            <!-- TODO: check property return-masked-value TEST -->
            <v-text-field v-model="form.token" flat solo hide-details return-masked-value :mask="tokenMask" autocomplete="off" color="secondary-dark" background-color="secondary-light" :label="$gettext('Activation Code')" type="text"> </v-text-field>
            <div class="action-buttons text-left mt-6">
              <v-btn v-if="$config.getTier() >= 4" href="https://my.photoprism.app/dashboard" target="_blank" color="primary-button lighten-2" :block="$vuetify.display.xs" class="ml-0" variant="outlined" :disabled="busy">
                <translate>Manage Account</translate>
              </v-btn>
              <v-btn v-else color="primary-button lighten-2" :block="$vuetify.display.xs" class="ml-0" variant="outlined" :disabled="busy" @click.stop="compare">
                <translate>Compare Editions</translate>
              </v-btn>

              <v-btn v-if="!form.token.length" color="primary-button" class="text-white ml-0 action-proceed" :block="$vuetify.display.xs" variant="flat" :disabled="busy" @click.stop="connect">
                <translate>Register</translate>
                <v-icon v-if="rtl" left dark>mdi-chevron-left</v-icon>
                <v-icon v-else right dark>mdi-chevron-right</v-icon>
              </v-btn>
              <v-btn v-else color="primary-button" :block="$vuetify.display.xs" class="text-white ml-0 action-activate" variant="flat" :disabled="busy || form.token.length !== tokenMask.length" @click.stop="activate">
                <translate>Activate</translate>
                <v-icon v-if="rtl" left dark>mdi-chevron-left</v-icon>
                <v-icon v-else right dark>mdi-chevron-right</v-icon>
              </v-btn>
            </div>
          </v-alert>
        </v-col>
        <v-col cols="12" class="px-2 pt-6 pb-0 text-body-1 text-selectable">
          <translate>You are welcome to contact us at membership@photoprism.app for questions regarding your membership.</translate>
          <translate>By using the software and services we provide, you agree to our terms of service, privacy policy, and code of conduct.</translate>
        </v-col>
      </v-row>
    </v-form>
    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import * as options from "options/options";
import Api from "common/api";
import { restart } from "common/server";

export default {
  name: "PPageConnect",
  data() {
    const token = this.$route.params.token ? this.$route.params.token : "";
    const membership = this.$config.getMembership();
    return {
      success: false,
      busy: false,
      valid: false,
      error: "",
      options: options,
      isPublic: this.$config.isPublic(),
      isAdmin: this.$session.isAdmin(),
      isDemo: this.$config.isDemo(),
      isSponsor: this.$config.isSponsor(),
      tier: this.$config.getTier(),
      membership: membership,
      showInfo: !token && membership === "ce",
      rtl: this.$rtl,
      tokenMask: "nnnn-nnnn-nnnn",
      form: {
        token,
      },
    };
  },
  created() {
    this.$config.load().then(() => {
      if (this.$config.isPublic() || !this.$session.isSuperAdmin()) {
        this.$router.push({ name: "home" });
      }
    });
  },
  methods: {
    onRestart() {
      restart(this.$router.resolve({ name: "about" }).href);
    },
    reset() {
      this.success = false;
      this.busy = false;
      this.error = "";
    },
    compare() {
      window.open("https://link.photoprism.app/personal-editions", "_blank").focus();
    },
    connect() {
      window.location = "https://my.photoprism.app/connect/" + encodeURIComponent(window.location);
    },
    activate() {
      if (!this.form.token || this.form.token.length !== this.tokenMask.length) {
        return;
      }

      const values = { Token: this.form.token };

      if (values.Token.length >= 4) {
        this.busy = true;
        this.$notify.blockUI();
        Api.put("connect/hub", values)
          .then(() => {
            this.$notify.success(this.$gettext("Connected"));
            this.success = true;
            this.busy = false;
            this.$config.update();
          })
          .catch((error) => {
            this.busy = false;
            if (error.response && error.response.data) {
              let data = error.response.data;
              this.error = data.message ? data.message : data.error;
            }

            if (!this.error) {
              this.error = this.$gettext("Invalid parameters");
            }
          })
          .finally(() => {
            this.$notify.unblockUI();
          });
      } else {
        this.$notify.error(this.$gettext("Invalid parameters"));
        this.$router.push({ name: "upgrade" });
      }
    },
    getMembership() {
      const m = this.$config.getMembership();
      switch (m) {
        case "":
        case "ce":
          return "Community";
        case "cloud":
          return "Cloud";
        case "essentials":
          return "Essentials";
        default:
          return "Plus";
      }
    },
  },
};
</script>
