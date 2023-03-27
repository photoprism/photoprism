<template>
  <div class="p-page p-page-about">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        {{ $config.getAbout() }} {{ getMembership() }}
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://www.photoprism.app/" target="_blank" class="action-info" :title="$gettext('Learn more')">
        <v-icon size="26" v-html="'$vuetify.icons.prism'"></v-icon>
      </v-btn>
    </v-toolbar>
    <v-container fluid class="px-4 pt-4 pb-1">
      <p class="subheading py-2 text-selectable">
        <strong><translate>At PhotoPrism, we believe that every moment captured through a photograph is precious, and our mission is to enable people to cherish those moments for generations to come.</translate></strong>
      </p>

      <template v-if="canUpgrade">
        <h3 class="subheading py-2"><translate>Support Our Mission</translate></h3>
        <p class="text-selectable">
          <span v-if="membership !== 'essentials'"><translate>Become a member today to enjoy additional features and support our mission!</translate></span>
          <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
          <translate>Being 100% self-funded and independent, we can promise you that we will never sell your data and that we will always be transparent about our software and services.</translate>
        </p>
        <p v-if="isAdmin && !isPublic" class="text-xs-center my-4">
          <v-btn
              to="/upgrade"
              color="primary-button"
              class="white--text px-3 py-2 action-upgrade"
              round depressed
          >
            <translate>Upgrade Now</translate>
            <v-icon v-if="rtl" left dark>navigate_before</v-icon>
            <v-icon v-else right dark>navigate_next</v-icon>
          </v-btn>
        </p>
      </template>

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
    const membership = this.$config.getMembership();
    return {
      rtl: this.$rtl,
      membership: membership,
      canUpgrade: membership === 'ce' || membership === 'essentials',
      isPublic: this.$config.isPublic(),
      isAdmin: this.$session.isAdmin(),
      isDemo: this.$config.isDemo(),
      isSponsor: this.$config.isSponsor(),
    };
  },
  methods: {
    getMembership() {
      const m = this.$config.getMembership();
      switch (m) {
        case "ce":
        case "cloud":
          return "";
        case "essentials":
          return "Essentials";
        default:
          return "Plus";
      }
    }
  },
};
</script>
