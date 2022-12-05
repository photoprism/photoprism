<template>
  <div class="p-page p-page-upgrade">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>Support Our Mission</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://link.photoprism.app/personal-editions" target="_blank" class="action-upgrade"
             :title="$gettext('Upgrade')">
        <v-icon size="26">diamond</v-icon>
      </v-btn>
    </v-toolbar>
    <v-form ref="form" v-model="valid" autocomplete="off" class="px-3 pt-3 pb-0" lazy-validation>
      <v-layout row wrap>
        <v-flex xs12 class="px-2 pt-2 pb-0">
          <p class="subheading text-selectable">
            <strong>
              <translate>Upgrade now and enjoy our member benefits!</translate>
            </strong>
          </p>
          <p class="subheading text-selectable">
            <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
            <translate>To upgrade, you may either enter an activation code or click on "Sign Up" to upgrade on our website:</translate>
          </p>
        </v-flex>
        <v-flex xs12 class="pa-2">
          <v-text-field v-model="form.token" flat solo hide-details return-masked-value :mask="tokenMask"
                        browser-autocomplete="off"
                        color="secondary-dark"
                        background-color="secondary-light" :label="$gettext('Activation Code')" type="text">
          </v-text-field>
        </v-flex>
        <v-flex xs12 grow class="px-2 py-1">
          <v-btn color="secondary-light" :block="$vuetify.breakpoint.xsOnly"
                 class="ml-0"
                 depressed
                 :disabled="busy"
                 @click.stop="compare">
            <translate>Compare Features</translate>
          </v-btn>
          <v-btn v-if="form.token.length < 1" color="primary-button"
                 class="white--text ml-0" :block="$vuetify.breakpoint.xsOnly"
                 depressed
                 :disabled="busy"
                 @click.stop="upgrade">
            <translate>Sign Up</translate>
            <v-icon :right="!rtl" :left="rtl" dark>navigate_next</v-icon>
          </v-btn>
          <v-btn v-else color="primary-button" :block="$vuetify.breakpoint.xsOnly"
                 class="white--text ml-0"
                 depressed
                 :disabled="busy || form.token.length < 4"
                 @click.stop="activate">
            <translate>Activate</translate>
            <v-icon :right="!rtl" :left="rtl" dark>navigate_next</v-icon>
          </v-btn>
        </v-flex>
        <v-flex xs12 class="px-2 pt-3 pb-0">
          <p class="body-1 text-selectable">
            <translate>Feel free to contact us at hello@photoprism.app if you have any questions.</translate>
          </p>
        </v-flex>
        <v-flex xs12 class="px-2 pt-3 pb-0">
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
            <translate>Our team evaluates this on an ongoing basis, depending on the support effort features and config options cause or have caused in the past, and whether they are generally needed by everyone or mainly requested by organizations and advanced users. As this allows us to make more features available to the public, we encourage all users to support our mission.</translate>
          </p>
          <p><a href="https://link.photoprism.app/membership" class="text-link" target="_blank"><translate>Learn more</translate> ›</a></p>
        </v-flex>
      </v-layout>
    </v-form>
    <p-about-footer></p-about-footer>
  </div>
</template>


<script>
import * as options from "options/options";

export default {
  name: 'PPageUpgrade',
  data() {
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
      rtl: this.$rtl,
      tokenMask: 'nnnn-nnnn-nnnn',
      form: {
        name: "hub",
        token: "",
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
    compare() {
      window.open('https://link.photoprism.app/personal-editions', '_blank').focus();
    },
    upgrade() {
      window.location = 'https://my.photoprism.app/register?upgrade=' + encodeURIComponent(window.location);
    },
    activate() {
      this.$router.push({name: "connect", params: {name: "hub", token: this.form.token}});
    }
  },
};
</script>
