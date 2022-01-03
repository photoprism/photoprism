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
    <v-container fluid class="pa-4">
      <p class="body-2">
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

      <p class="body-1 pb-2">
        <translate>Your continued support helps us provide regular updates and services like world maps.</translate>
        <translate>Requests endorsed by silver and gold sponsors receive a golden label and will be prioritized.</translate>
        <translate>Thank you very much!</translate>
      </p>

      <h2 class="mt-4 mb-2 subheading">
        <translate>PhotoPrism® is a registered trademark</translate>
      </h2>
      <p class="body-1">
        Our <a target="_blank" href="https://photoprism.app/trademark">Trademark and Brand Guidelines</a>,
        which may be updated from time to time, describe how our Brand Assets may be used.
        It is important to us that any permitted use of our Brand Assets is fair and meets the highest standards.
      </p>
      <p class="body-1 pb-2">
        To ensure this, we reserve the right to object to any use or misuse in any jurisdiction worldwide.
        Visit <a target="_blank" href="https://photoprism.app/trademark">photoprism.app/trademark</a> to learn more.
      </p>

      <p class="text-xs-center pt-3 mb-2">
        <router-link to="/about/license">
          <img :src="$config.staticUri + '/img/badge-agpl.svg'" alt="License AGPL v3" style="max-width:100%;"/>
        </router-link>
        <a target="_blank" href="https://docs.photoprism.app/" rel="nofollow"><img :src="$config.staticUri + '/img/badge-docs.svg'"
                                                                                   alt="Documentation"
                                                                                   style="max-width:100%;"></a>
        <a target="_blank" href="https://gitter.im/browseyourlife/community" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-chat.svg'" alt="Community Chat" style="max-width:100%;"></a>
        <a target="_blank" href="https://twitter.com/photoprism_app" rel="nofollow"><img
            :src="$config.staticUri + '/img/badge-twitter.svg'" alt="Twitter" style="max-width:100%;"></a>
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
