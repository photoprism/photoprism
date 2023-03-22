<template>
  <div class="p-page p-page-about">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>About</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://www.photoprism.app/" target="_blank" class="action-info" :title="$gettext('Learn more')">
        <v-icon size="26">diamond</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container fluid class="px-4 pt-4 pb-1">
      <h3 class="title text-selectable font-weight-bold py-2">
        PhotoPrism® - AI-Powered Photos App for the Decentralized Web
      </h3>

      <p class="body-2 pt-2 text-selectable">
        <translate>Our mission is to provide the most user- and privacy-friendly solution to keep your pictures organized and accessible.</translate>
      </p>

      <h3 class="subheading py-2"><translate>PhotoPrism+ Membership</translate></h3>
      <p class="text-selectable">
        <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
        <translate>Being 100% self-funded and independent, we can promise you that we will never sell your data and that we will always be transparent about our software and services.</translate>
      </p>

      <div v-if="isAdmin && !isSponsor && !isPublic">
        <p class="text-xs-center my-4">
          <v-btn
              to="/upgrade"
              color="primary-button"
              class="white--text px-3 py-2 action-upgrade"
              round depressed
          >
            <translate>Upgrade Now</translate>
            <v-icon :left="rtl" :right="!rtl" size="18" class="ml-2" dark>verified</v-icon>
          </v-btn>
        </p>
      </div>
      <div v-else>
        <p class="text-xs-center my-4">
          <v-btn
              href="https://link.photoprism.app/membership"
              target="_blank"
              color="primary-button"
              class="white--text px-3 py-2 action-upgrade"
              round depressed
          >
            <translate>Learn more</translate>
            <v-icon :left="rtl" :right="!rtl" size="18" class="ml-2" dark>diamond</v-icon>
          </v-btn>
        </p>
      </div>

      <h3 class="subheading py-2"><translate>User Guide</translate></h3>
      <p class="text-selectable">
        <translate>Visit docs.photoprism.app/user-guide to learn how to sync, organize, and share your pictures.</translate>
        <translate>Our User Guide also covers many advanced topics, such as migrating from Google Photos and thumbnail quality settings.</translate>
        <translate>Common issues can be quickly diagnosed and solved using the troubleshooting checklists we provide.</translate>
      </p>
      <p><a href="https://link.photoprism.app/docs" class="text-link" target="_blank"><translate>Read the Docs</translate> ›</a></p>

      <h3 class="subheading py-2"><translate>Knowledge Base</translate></h3>
      <p class="text-selectable"><translate>Browse the Knowledge Base for detailed information on specific product features, services, and related resources.</translate></p>
      <p><a href="https://www.photoprism.app/kb" class="text-link" target="_blank"><translate>Learn more</translate> ›</a></p>

      <h3 class="subheading py-2">
        <translate>Getting Support</translate>
      </h3>
      <p class="body-1 text-selectable">
        <a target="_blank" href="https://docs.photoprism.app/getting-started/troubleshooting/">
          <translate>Before submitting a support request, please use our Troubleshooting Checklists to determine the cause of your problem.</translate>
          <translate>If this doesn't help, or you have other questions:</translate>
        </a>
      </p>
      <ul class="body-1 mb-3">
        <li><a target="_blank" href="https://link.photoprism.app/reddit"><translate>you are welcome to join us on Reddit</translate></a></li>
        <li><a target="_blank" href="https://link.photoprism.app/discussions"><translate>post your question in GitHub Discussions</translate></a></li>
        <li><a target="_blank" href="https://link.photoprism.app/chat"><translate>or ask in our Community Chat</translate></a></li>
      </ul>
      <p class="body-1 text-selectable pb-2">
        <a target="_blank" href="https://www.photoprism.app/contact"><translate>In addition, sponsors receive direct technical support via email.</translate></a>
        <span v-if="!isSponsor">
          <translate>We'll do our best to answer all your questions. In return, we ask you to back us on Patreon or GitHub Sponsors.</translate>
        </span>
      </p>

      <p v-if="isSponsor" class="text-xs-center">
        <img src="https://cdn.photoprism.app/thank-you/colorful.png" width="100%" alt="THANK YOU">
      </p>

      <p class="text-xs-center pt-2 ma-0 pb-0">
        <router-link to="/license">
          <img :src="$config.staticUri + '/img/badge-agpl.svg'" alt="License AGPL v3" style="max-width:100%;"/>
        </router-link>
        <a target="_blank" href="https://docs.photoprism.app/"><img :src="$config.staticUri + '/img/badge-docs.svg'"
                                                                                   alt="Official Documentation"
                                                                                   style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/chat" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-chat.svg'" alt="Community Chat" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/discussions" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-ask-on-github.svg'" alt="GitHub Discussions" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/pixls-us" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-pixls-us.svg'" alt="PIXLS.US" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/twitter" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-twitter.svg'" alt="Twitter" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/reddit" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-reddit.svg'" alt="Reddit" style="max-width:100%;"></a>
      </p>

      <p class="caption mt-4 text-xs-center">
        PhotoPrism® is a <a href="https://www.photoprism.app/trademark" target="_blank" class="text-link">registered trademark</a>.
        By using the software and services we provide, you agree to our
        <a href="https://www.photoprism.app/terms" target="_blank" class="text-link">Terms of Service</a>,
        <a href="https://www.photoprism.app/privacy" target="_blank" class="text-link">Privacy Policy</a>, and
        <a href="https://www.photoprism.app/code-of-conduct" target="_blank" class="text-link">Code of Conduct</a>.
      </p>
    </v-container>
    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
export default {
  name: 'PPageAbout',
  data() {
    return {
      rtl: this.$rtl,
      isPublic: this.$config.isPublic(),
      isAdmin: this.$session.isAdmin(),
      isDemo: this.$config.isDemo(),
      isSponsor: this.$config.isSponsor(),
    };
  },
  methods: {},
};
</script>
