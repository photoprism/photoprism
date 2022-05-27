<template>
  <div class="p-page p-page-about">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>About</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://photoprism.app/" target="_blank" class="action-info" :title="$gettext('About')">
        <v-icon size="26">chat</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container fluid class="px-4 pt-4 pb-1">
      <p class="body-2 lh-17">
          <translate>Our mission is to provide the most user- and privacy-friendly solution to keep your pictures organized and accessible.</translate>
          <a href="https://link.photoprism.app/roadmap" target="_blank">
            <translate>The roadmap shows what tasks are in progress, what needs testing, and which feature requests are going to be implemented next.</translate>
          </a>
      </p>

      <p v-if="!sponsor" class="text-xs-center my-4">
        <v-btn
            href="https://photoprism.app/membership"
            target="_blank"
            color="primary-button"
            class="white--text px-3 py-2 action-sponsor"
            round depressed small
        >
          <translate>Become a sponsor</translate>
          <v-icon :left="rtl" :right="!rtl" size="16" class="ml-2" dark>star</v-icon>
        </v-btn>
      </p>

      <p class="body-1">
        <translate>Your continued support helps us provide regular updates and services like world maps.</translate>
        <translate>Sponsors get access to additional features, receive direct technical support via email, and can join our private chat room on matrix.org.</translate>
      </p>
      <p v-if="!sponsor" class="body-1 pb-2">
        <a target="_blank" href="https://link.photoprism.app/github">
        <translate>Also, please leave a star on GitHub if you like this project. It provides additional motivation to keep going.</translate>
        </a>
      </p>
      <h2 class="mt-4 mb-2 body-2 lh-17">
        <translate>Getting Support</translate>
      </h2>
      <p class="body-1">
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
      <p class="body-1">
        <a target="_blank" href="https://photoprism.app/contact"><translate>In addition, sponsors receive direct technical support via email.</translate></a>
      </p>

      <p v-if="!sponsor" class="body-1 pb-2">
        <translate>We'll do our best to answer all your questions. In return, we ask you to back us on Patreon or GitHub Sponsors.</translate>
      </p>

      <p v-if="sponsor" class="text-xs-center">
        <img src="https://cdn.photoprism.app/thank-you/colorful.png" width="100%" alt="THANK YOU">
      </p>

      <p class="text-xs-center pt-2 ma-0 pb-0">
        <router-link to="/about/license">
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
        PhotoPrism® is a <a href="https://photoprism.app/trademark" target="_blank" class="text-link">registered trademark</a>.
        By using the software and services we provide, you agree to our
        <a href="https://photoprism.app/terms" target="_blank" class="text-link">Terms of Service</a>,
        <a href="https://photoprism.app/privacy" target="_blank" class="text-link">Privacy Policy</a>, and
        <a href="https://photoprism.app/code-of-conduct" target="_blank" class="text-link">Code of Conduct</a>.
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
      sponsor: this.$config.isSponsor(),
    };
  },
  methods: {},
};
</script>
