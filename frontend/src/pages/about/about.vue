<template>
  <div class="p-page p-page-about">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>About</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://photoprism.app/" target="_blank" class="action-info" :title="$gettext('About')">
        <v-icon size="26">info_outline</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container fluid class="px-4 pt-4 pb-1">
      <p class="body-2 lh-17">
          <translate>Our vision is to provide the most user- and privacy-friendly solution to keep your pictures organized and accessible.</translate>
          <a href="https://github.com/photoprism/photoprism/projects/5" target="_blank">
            <translate>The roadmap shows what tasks are in progress, what needs testing, and which feature requests are going to be implemented next.</translate>
          </a>
      </p>

      <p class="text-xs-center my-4">
        <v-btn
            href="https://www.patreon.com/photoprism"
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
      <p class="body-1 pb-2">
        <a target="_blank" href="https://github.com/photoprism/photoprism">
        <translate>Also, please leave a star on GitHub if you like this project. It provides additional motivation to keep going.</translate>
        </a>
      </p>
      <h2 class="mt-4 mb-2 subheading">
        <translate>Getting Support</translate>
      </h2>
      <p class="body-1">
        <a target="_blank" href="https://docs.photoprism.app/getting-started/troubleshooting/">
        <translate>Before submitting a support request, please use our Troubleshooting Checklists to determine the cause of your problem.</translate>
        <translate>If this doesn't help, or you have other questions:</translate>
        </a>
      </p>
      <ul class="body-1 mb-3">
        <li><a target="_blank" href="https://www.reddit.com/r/photoprism/"><translate>you are welcome to join us on Reddit</translate></a></li>
        <li><a target="_blank" href="https://github.com/photoprism/photoprism/discussions"><translate>post your question in GitHub Discussions</translate></a></li>
        <li><a target="_blank" href="https://gitter.im/browseyourlife/community"><translate>or ask in our Community Chat</translate></a></li>
      </ul>
      <p class="body-1">
        <a target="_blank" href="https://photoprism.app/contact"><translate>In addition, sponsors receive direct technical support via email.</translate></a>
      </p>
      <p class="body-1 pb-2">
        <translate>We'll do our best to answer all your questions. In return, we ask you to back us on Patreon or GitHub Sponsors.</translate>
      </p>

      <p class="text-xs-center pt-2 ma-0 pb-0">
        <router-link to="/about/license">
          <img :src="$config.staticUri + '/img/badge-agpl.svg'" alt="License AGPL v3" style="max-width:100%;"/>
        </router-link>
        <a target="_blank" href="https://docs.photoprism.app/" rel="nofollow"><img :src="$config.staticUri + '/img/badge-docs.svg'"
                                                                                   alt="Documentation"
                                                                                   style="max-width:100%;"></a>
        <a target="_blank" href="https://gitter.im/browseyourlife/community" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-chat.svg'" alt="Community Chat" style="max-width:100%;"></a>
        <a target="_blank" href="https://github.com/photoprism/photoprism/discussions" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-ask-on-github.svg'" alt="GitHub Discussions" style="max-width:100%;"></a>
        <a target="_blank" href="https://discuss.pixls.us/photoprism" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-pixls-us.svg'" alt="PIXLS.US" style="max-width:100%;"></a>
        <a target="_blank" href="https://twitter.com/photoprism_app" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-twitter.svg'" alt="Twitter" style="max-width:100%;"></a>
        <a target="_blank" href="https://www.reddit.com/r/photoprism/" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-reddit.svg'" alt="Reddit" style="max-width:100%;"></a>
      </p>

      <p class="caption mt-4">
        <a target="_blank" href="https://photoprism.app/trademark">
        PhotoPrism® is a registered trademark. Our Trademark and Brand Guidelines,
        which may be updated from time to time, describe how our Brand Assets may be used.
        It is important to us that any permitted use of our Brand Assets is fair and meets the highest standards.
        Visit photoprism.app/trademark to learn more.
        </a>
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
    };
  },
  methods: {},
};
</script>
