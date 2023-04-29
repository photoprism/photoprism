<template>
  <div class="p-page p-page-upgrade">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>Membership</translate>
        <v-icon v-if="rtl">navigate_before</v-icon>
        <v-icon v-else>navigate_next</v-icon>
        <span v-if="busy">
          <translate>Busy, please wait…</translate>
        </span>
        <span v-else-if="success">
          <translate>Successfully Connected</translate>
        </span>
        <span v-else-if="error">
          <translate>Invalid</translate>
        </span>
        <span v-else>
          <translate>Upgrade</translate>
        </span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://link.photoprism.app/personal-editions" target="_blank" class="action-upgrade"
             :title="$gettext('Learn more')">
        <v-icon size="26" color="secondary-dark" v-html="'$vuetify.icons.prism'"></v-icon>
      </v-btn>
    </v-toolbar>
    <v-form ref="form" v-model="valid" autocomplete="off" class="px-3 pt-3 pb-0" lazy-validation>
      <v-layout v-if="busy" row wrap>
        <v-flex xs12 d-flex class="text-sm-center pa-2">
          <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
        </v-flex>
      </v-layout>
      <v-layout v-else-if="error" row wrap>
        <v-flex xs12 class="text-sm-left pa-2">
          <v-alert
              :value="true"
              color="error"
              icon="gpp_bad"
              class="mt-3"
              outline
          >
            {{ error }}
          </v-alert>
        </v-flex>
        <v-flex xs12 class="pa-2">
          <v-btn color="primary-button lighten-2" :block="$vuetify.breakpoint.xsOnly"
                 class="ml-0"
                 outline
                 :disabled="busy"
                 @click.stop="reset">
            <translate>Cancel</translate>
          </v-btn>
          <v-btn color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                 class="white--text ml-0"
                 href="https://www.photoprism.app/contact"
                 target="_blank"
                 depressed>
            <translate>Contact Us</translate>
          </v-btn>
        </v-flex>
      </v-layout>
      <v-layout v-else-if="success" row wrap>
        <v-flex xs12 d-flex class="pa-2">
          <p class="subheading text-xs-left">
            <translate>Your account has been successfully connected.</translate>
            <span v-if="$config.values.restart">
            <translate>Please restart your instance for the changes to take effect.</translate>
            </span>
          </p>
        </v-flex>
        <v-flex xs12 grow class="pa-2">
          <v-btn href="https://my.photoprism.app/dashboard" target="_blank" color="primary-button lighten-2" :block="$vuetify.breakpoint.xsOnly"
                 class="ml-0" outline :disabled="busy">
              <translate>Manage Account</translate>
          </v-btn>
          <v-btn v-if="$config.values.restart" color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                 class="white--text ml-0" depressed :disabled="busy" @click.stop.p.prevent="onRestart">
            <translate>Restart</translate>
            <v-icon :right="!rtl" :left="rtl" dark>restart_alt</v-icon>
          </v-btn>
          <v-btn v-if="membership === ''" href="https://my.photoprism.app/get-started" target="_blank" color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                 class="white--text ml-0" depressed :disabled="busy">
            <translate>Upgrade Now</translate>
            <v-icon v-if="rtl" left dark>navigate_before</v-icon>
            <v-icon v-else right dark>navigate_next</v-icon>
          </v-btn>
        </v-flex>
      </v-layout>
      <v-layout v-else row wrap>
        <v-flex xs12 grow align-center justify-center class="px-2 py-1">
          <v-alert
              :value="true"
              color="secondary-dark"
              outline
          >
          <p class="subheading text-selectable">
            <strong><translate>To upgrade, you can either enter an activation code or click "Register" to sign up on our website:</translate></strong>
          </p>
          <v-text-field v-model="form.token" flat solo hide-details return-masked-value :mask="tokenMask"
                        browser-autocomplete="off"
                        color="secondary-dark"
                        background-color="secondary-light" :label="$gettext('Activation Code')" type="text">
          </v-text-field>
          <div class="action-buttons text-xs-left mt-3">
            <v-btn v-if="membership && membership !== 'ce'" href="https://my.photoprism.app/dashboard" target="_blank" color="primary-button lighten-2" :block="$vuetify.breakpoint.xsOnly"
                   class="ml-0"
                   outline
                   :disabled="busy">
              <translate>Manage Account</translate>
            </v-btn>
            <v-btn v-else color="primary-button lighten-2" :block="$vuetify.breakpoint.xsOnly"
                   class="ml-0"
                   outline
                   :disabled="busy"
                   @click.stop="compare">
              <translate>Compare Editions</translate>
            </v-btn>

            <v-btn v-if="!form.token.length" color="primary-button"
                   class="white--text ml-0 action-proceed" :block="$vuetify.breakpoint.xsOnly"
                   depressed
                   :disabled="busy"
                   @click.stop="connect">
              <translate>Register</translate>
              <v-icon v-if="rtl" left dark>navigate_before</v-icon>
              <v-icon v-else right dark>navigate_next</v-icon>
            </v-btn>
            <v-btn v-else color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                   class="white--text ml-0 action-activate"
                   depressed
                   :disabled="busy || form.token.length !== tokenMask.length"
                   @click.stop="activate">
              <translate>Activate</translate>
              <v-icon v-if="rtl" left dark>navigate_before</v-icon>
              <v-icon v-else right dark>navigate_next</v-icon>
            </v-btn>
          </div>
          </v-alert>
        </v-flex>
        <v-flex xs12 class="px-2 pt-3 pb-0">
          <p class="body-1 text-selectable">
            <translate>You are welcome to contact us at membership@photoprism.app for questions regarding your membership.</translate>
            <translate>By using the software and services we provide, you agree to our terms of service, privacy policy, and code of conduct.</translate>
          </p>
        </v-flex>
        <v-flex v-show="showInfo" xs12 class="px-2 pt-3 pb-0">
          <h3 class="title pb-3">
            <translate>Frequently Asked Questions</translate>
          </h3>
          <p class="subheading text-selectable">
            <translate>Why are some features only available to members?</translate>
          </p>
          <p class="body-1 text-selectable">
            <translate>PhotoPrism is 100% self-funded and independent.</translate>
            <translate>Voluntary donations do not cover the cost of a team working full time to provide you with updates, documentation, and support.</translate>
            <translate>It is your decision whether you want to sign up to enjoy additional benefits.</translate>
          </p>
          <p class="subheading text-selectable">
            <translate>What functionality is generally available?</translate>
          </p>
          <p class="body-1 text-selectable">
            <translate>Our team decides this on an ongoing basis depending on the support effort required, server and licensing costs, and whether the features are generally needed by everyone or mainly requested by organizations and advanced users.</translate>
            <translate>As this helps us provide more features to the public, we encourage all users to support our mission.</translate>
          </p>
          <p class="body-1">
            <a href="https://www.photoprism.app/oss/faq" class="text-link" target="_blank"><translate>Learn more</translate> ›</a>
          </p>
        </v-flex>
      </v-layout>
    </v-form>
    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import * as options from "options/options";
import Api from "common/api";
import {restart} from "common/server";

export default {
  name: 'PPageConnect',
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
      tokenMask: 'nnnn-nnnn-nnnn',
      form: {
        token,
      },
    };
  },
  created() {
    this.$config.load().then(() => {
      if (this.$config.isPublic() || !this.$session.isSuperAdmin()) {
        this.$router.push({name: "home"});
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
      window.open('https://link.photoprism.app/personal-editions', '_blank').focus();
    },
    connect() {
      window.location = 'https://my.photoprism.app/connect/' + encodeURIComponent(window.location);
    },
    activate() {
      if (!this.form.token || this.form.token.length !== this.tokenMask.length) {
        return;
      }

      const values = {Token: this.form.token};

      if (values.Token.length >= 4) {
        this.busy = true;
        this.$notify.blockUI();
        Api.put("connect/hub", values).then(() => {
          this.$notify.success(this.$gettext("Connected"));
          this.success = true;
          this.busy = false;
          this.$config.update();
        }).catch((error) => {
          this.busy = false;
          if (error.response && error.response.data) {
            let data = error.response.data;
            this.error = data.message ? data.message : data.error;
          }

          if(!this.error) {
            this.error = this.$gettext("Invalid parameters");
          }
        }).finally(() => {
          this.$notify.unblockUI();
        });
      } else {
        this.$notify.error(this.$gettext("Invalid parameters"));
        this.$router.push({name: "upgrade"});
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
