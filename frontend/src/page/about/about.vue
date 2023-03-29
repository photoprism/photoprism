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
      <p class="subheading py-1 pb-2 text-selectable">
        <strong><translate>At PhotoPrism, we believe that every moment captured through a photograph is precious, and our mission is to enable people to cherish those moments for generations to come.</translate></strong>
      </p>

      <template v-if="canUpgrade">
        <h3 class="subheading py-2"><translate>Support Our Mission</translate></h3>
        <p class="text-selectable">
          <span v-if="membership !== 'essentials'"><translate>Become a member today to enjoy additional features and support our mission!</translate></span>
          <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
          <translate>Being 100% self-funded and independent, we can promise you that we will never sell your data and that we will always be transparent about our software and services.</translate>
        </p>
        <p v-if="isSuperAdmin" class="text-xs-center my-4">
          <v-btn
              to="/upgrade"
              color="primary-button"
              class="white--text px-3 py-2 action-membership"
              round depressed
          >
            <translate>Upgrade Now</translate>
            <v-icon v-if="rtl" left dark>navigate_before</v-icon>
            <v-icon v-else right dark>navigate_next</v-icon>
          </v-btn>
        </p>
        <p v-else class="text-xs-center my-4">
          <v-btn
              href="https://link.photoprism.app/membership"
              target="_blank"
              color="primary-button"
              class="white--text px-3 py-2 action-membership"
              round depressed
          >
            <translate>Learn more</translate>
            <v-icon v-if="rtl" left dark>navigate_before</v-icon>
            <v-icon v-else right dark>navigate_next</v-icon>
          </v-btn>
        </p>
      </template>
      <template v-else-if="isSuperAdmin">
        <h3 class="subheading py-2"><translate>Thank You for Your Support!</translate> <v-icon size="20" color="primary">favorite</v-icon></h3>
        <p class="text-selectable">
          <translate>PhotoPrism is 100% self-funded and independent.</translate>
          <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
          <translate>Feel free to contact us at members@photoprism.app for questions regarding your membership.</translate>
        </p>
        <p class="text-xs-center my-4">
          <v-btn
              href="https://my.photoprism.app/dashboard" target="_blank"
              color="primary-button"
              class="white--text px-3 py-2 action-membership"
              round depressed
          >
            <translate>Manage Account</translate>
            <v-icon v-if="rtl" left dark>navigate_before</v-icon>
            <v-icon v-else right dark>navigate_next</v-icon>
          </v-btn>
        </p>
      </template>

      <div class="text-columns py-2">
        <h3 class="subheading pb-2">Getting Started</h3>
        <p class="text-selectable">
          Follow our <a href="https://docs.photoprism.app/user-guide/first-steps/" class="text-link" target="_blank">First Steps 👣</a> tutorial to learn how to navigate the user interface and ensure your library is indexed according to your individual preferences.
          Additional help and product-specific information can be found in our <a href="https://www.photoprism.app/plus/kb" class="text-link" target="_blank">Knowledge Base</a>.
        </p>

        <h3 class="subheading pb-2">User Guide</h3>
        <p>
          Visit <a href="https://link.photoprism.app/docs" class="text-link" target="_blank">docs.photoprism.app/user-guide</a> to learn how to sync, organize, and share your pictures. Our <a href="https://docs.photoprism.app/user-guide/" class="text-link" target="_blank">User Guide</a> also covers many advanced topics, such as <a href="https://docs.photoprism.app/user-guide/use-cases/google/" class="text-link" target="_blank">migrating from Google Photos</a> and <a href="https://docs.photoprism.app/user-guide/settings/advanced/#images" class="text-link" target="_blank">thumbnail quality settings</a>.
          Common issues can be quickly diagnosed and solved using the troubleshooting checklists we provide at <a href="https://docs.photoprism.app/getting-started/troubleshooting/" class="text-link" target="_blank">docs.photoprism.app/getting-started/troubleshooting</a>.
        </p>

        <h3 class="subheading pb-2">Getting Support</h3>
        <p>Before reporting a bug, please use our <a href="https://docs.photoprism.app/getting-started/troubleshooting/" class="text-link" target="_blank">Troubleshooting Checklists</a>
          to determine the cause of your problem. If you have a general question, need help, it could be a local configuration
          issue, or a misunderstanding in how the software works, you are welcome to ask in our <a href="https://link.photoprism.app/chat" class="text-link" target="_blank">Community Chat</a>
          or post your question in <a href="https://link.photoprism.app/discussions" class="text-link" target="_blank">GitHub Discussions</a></p>
        <p>When reporting a problem, always include the software versions you are using and <a href="https://www.photoprism.app/kb/reporting-bugs" class="text-link" target="_blank">other information about your environment</a>
          such as <a href="https://docs.photoprism.app/getting-started/troubleshooting/browsers/" class="text-link" target="_blank">browser, browser plugins</a>, operating system, storage type,
          memory size, and processor.</p>
        <p>We kindly ask you not to report bugs via GitHub Issues unless you are certain to have found a fully reproducible and previously unreported issue that must be fixed directly in the app.</p>

        <h3 class="subheading pb-2">Developer Guide</h3>
        <p>Our <a href="https://docs.photoprism.app/developer-guide/" class="text-link" target="_blank">Developer Guide</a> contains all the information you need to get started as a developer. It guides you from <a href="https://docs.photoprism.app/developer-guide/setup/" class="text-link" target="_blank">setting up your development environment</a> and <a href="https://docs.photoprism.app/developer-guide/pull-requests/" class="text-link" target="_blank">creating pull requests</a> to <a href="https://docs.photoprism.app/developer-guide/tests/" class="text-link" target="_blank">running tests</a> and <a href="https://docs.photoprism.app/developer-guide/translations-weblate/" class="text-link" target="_blank">adding translations</a>. Multiple subsections provide details on specific features and links to external resources for further information.</p>

        <h3 class="subheading pb-2">Terms &amp; Privacy</h3>
        <p>Visit <a href="https://www.photoprism.app/terms" class="text-link" target="_blank"><strong>photoprism.app/terms</strong></a> to learn how we work, what you can expect from us, and what we expect from you.
          What information we collect, how we use it, and under what circumstances we share it is explained in our <a href="https://www.photoprism.app/privacy" class="text-link" target="_blank">Privacy Policy</a>.</p>

        <p>Read our <a href="https://www.photoprism.app/privacy/gdpr" class="text-link" target="_blank">GDPR Compliance Statement</a> to learn more about the rights you have as a resident of the European Economic Area ("EEA"), our ongoing commitment to user privacy, and the General Data Protection Regulation ("GDPR").</p>

        <p>Our <a href="https://www.photoprism.app/trademark" class="text-link" target="_blank">Trademark and Brand Guidelines</a>, which may be updated from time to time, describe how our brand assets may be used. It is important to us that any permitted use of our brand assets is fair and meets the highest standards.</p>
      </div>

      <p class="text-xs-center pt-4 ma-0 pb-0">
        <router-link to="/license">
          <img :src="$config.staticUri + '/img/badge-agpl.svg'" alt="License AGPL v3" style="max-width:100%;"/>
        </router-link>
        <a target="_blank" href="https://link.photoprism.app/chat" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-chat.svg'" alt="Community Chat" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/discussions" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-ask-on-github.svg'" alt="GitHub Discussions" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/pixls-us" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-pixls-us.svg'" alt="PIXLS.US" style="max-width:100%;"></a>
        <a target="_blank" href="https://link.photoprism.app/mastodon" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-mastodon.svg'" alt="Twitter" style="max-width:100%;"></a>
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
    const isDemo = this.$config.isDemo();
    const isPublic = this.$config.isPublic();
    return {
      rtl: this.$rtl,
      membership: membership,
      canUpgrade: membership === 'ce' || membership === 'essentials',
      isDemo: isDemo,
      isPublic: isPublic,
      isSuperAdmin: this.$session.isSuperAdmin() && !isPublic && !isDemo,
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
