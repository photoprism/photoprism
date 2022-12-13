<template>
  <div class="p-page p-page-upgrade">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>Upgrade</translate>
        <v-icon v-if="rtl">navigate_before</v-icon>
        <v-icon v-else>navigate_next</v-icon>
        <span v-if="busy">
          <translate>Busy, please wait…</translate>
        </span>
        <span v-else-if="success">
          <translate>Verified</translate>
        </span>
        <span v-else-if="error">
          <translate>Failed</translate>
        </span>
        <span v-else>
          <translate>Support Our Mission</translate>
      </span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://link.photoprism.app/personal-editions" target="_blank" class="action-upgrade"
             :title="$gettext('Learn more')">
        <v-icon size="26">diamond</v-icon>
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
          <v-btn color="secondary-light" :block="$vuetify.breakpoint.xsOnly"
                 class="ml-0"
                 depressed
                 :disabled="busy"
                 @click.stop="reset">
            <translate>Cancel</translate>
          </v-btn>
          <v-btn color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                 class="white--text ml-0"
                 href="https://photoprism.app/contact"
                 target="_blank"
                 depressed>
            <translate>Contact Us</translate>
          </v-btn>
        </v-flex>
      </v-layout>
      <v-layout v-else-if="success" row wrap>
        <v-flex xs12 d-flex class="text-sm-left pa-2">
          <v-alert
              :value="true"
              color="success"
              icon="verified"
              class="mt-3"
              outline
          >
            <translate>Successfully Connected</translate>
          </v-alert>
        </v-flex>
        <v-flex xs12 grow class="pa-2">
          <v-btn href="https://my.photoprism.app/dashboard" target="_blank" color="secondary-light"
                 class="ml-0" :block="$vuetify.breakpoint.xsOnly"
                 depressed
                 :disabled="busy">
              <translate>Manage account</translate>
          </v-btn>
          <v-btn href="https://my.photoprism.app/get-started" target="_blank" color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                 class="white--text ml-0"
                 depressed
                 :disabled="busy">
            <translate>Get started</translate>
            <v-icon v-if="rtl" left dark>navigate_before</v-icon>
            <v-icon v-else right dark>navigate_next</v-icon>
          </v-btn>
        </v-flex>
      </v-layout>
      <v-layout v-else row wrap>
        <v-flex xs12 class="px-2 pt-2 pb-0">
          <p class="subheading text-selectable">
            <strong><translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate></strong>
          </p>
          <p class="subheading text-selectable">
            <translate>To upgrade, you may either enter an activation code or click on "Proceed" to sign up on our website:</translate>
          </p>
        </v-flex>
        <v-flex xs12 class="pa-2">
          <v-text-field v-model="form.token" flat solo hide-details return-masked-value :mask="tokenMask"
                        browser-autocomplete="off"
                        color="secondary-dark"
                        background-color="secondary-light" :label="$gettext('Activation Code')" type="text">
          </v-text-field>
        </v-flex>
        <v-flex xs12 grow class="px-2 pb-2 pt-1">
          <v-btn v-if="!form.token.length" color="primary-button"
                 class="white--text ml-0 action-proceed" :block="$vuetify.breakpoint.xsOnly"
                 depressed
                 :disabled="busy"
                 @click.stop="connect">
            <translate>Proceed</translate>
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
        </v-flex>
        <v-flex xs12 class="px-2 pt-3 pb-0">
          <p class="body-1 text-selectable">
            <translate>Feel free to contact us at hello@photoprism.app if you have any questions.</translate>
          </p>
        </v-flex>
        <v-flex v-show="showInfo" xs12 class="px-2 pt-3 pb-0">
          <h3 class="pb-3">
            <translate>Frequently Asked Questions</translate>
          </h3>
          <p class="body-2 text-selectable">
            <translate>Shouldn't free software be free of costs?</translate>
          </p>
          <p class="body-1 text-selectable">
            <translate>Think of “free software” as in “free speech,” not as in “free beer.” The Free Software Foundation sometimes calls it “libre software,” borrowing the French or Spanish word for “free” as in freedom, to show they do not mean the software is gratis.</translate>
          </p>
          <p class="body-2 text-selectable">
            <translate>Why are some features only available to sponsors?</translate>
          </p>
          <p class="body-1 text-selectable">
            <translate>PhotoPrism is 100% self-funded. Voluntary donations do not cover the cost of a team working full time to provide you with updates, documentation, and support. It is your decision whether you want to sign up to enjoy additional benefits.</translate>
          </p>
          <p class="body-2 text-selectable">
            <translate>What functionality is generally available?</translate>
          </p>
          <p class="body-1 text-selectable">
            <translate>Our team decides this on an ongoing basis depending on the support effort required, server and licensing costs, and whether the features are generally needed by everyone or mainly requested by organizations and advanced users. As this allows us to make more features available to the public, we encourage all users to support our mission.</translate>
          </p>
        </v-flex>
        <v-flex v-show="showInfo" xs12 class="pa-2">
          <v-btn color="secondary-light" :block="$vuetify.breakpoint.xsOnly"
                 class="ml-0"
                 depressed
                 :disabled="busy"
                 @click.stop="compare">
            <translate>Compare Editions</translate>
            <v-icon :right="!rtl" :left="rtl" dark>compare_arrows</v-icon>
          </v-btn>
        </v-flex>
      </v-layout>
    </v-form>
    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import * as options from "options/options";
import Api from "common/api";

export default {
  name: 'PPageConnect',
  data() {
    const token = this.$route.params.token ? this.$route.params.token : "";
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
      showInfo: !token,
      rtl: this.$rtl,
      tokenMask: 'nnnn-nnnn-nnnn',
      form: {
        token,
      },
    };
  },
  created() {
    this.$config.load().then(() => {
      if (this.$config.isPublic() || !this.$session.isAdmin()) {
        this.$router.push({name: "home"});
      }
    });
  },
  methods: {
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
  },
};
</script>
