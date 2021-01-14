<template>
  <div class="p-page p-page-about">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>About</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://photoprism.app/" target="_blank" class="action-info" :title="$gettext('About')">
        <v-icon size="26">info</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container fluid class="pa-4">
      <h2 class="pb-2 subheading">Funding</h2>
      <p class="body-1">This <a href="https://photoprism.app/" target="_blank" class="text-link">software</a> is about freedom and privacy but not necessarily about free beer. We feel like it
        would be a mistake to state there will be no costs, because clearly we have huge expenses, your server hardware
        will have a price tag, and then maybe you'd like to have some extra features that need to be developed.</p>
      <p class="body-1">You're welcome to support us via <a href="https://github.com/sponsors/photoprism" target="_blank" class="text-link">GitHub
        Sponsors</a>,
        especially if you have feature requests, or need help with using our software.
        In addition, you can find us on <a href="https://www.patreon.com/photoprism" target="_blank" class="text-link">Patreon</a>
        and <a href="https://www.paypal.me/photoprism" target="_blank" class="text-link">PayPal</a>.</p>
      <p class="body-1">Your continuous support helps...</p>
      <ul class="pb-3 body-1">
        <li>to pay for hosting, hardware and external services like satellite maps</li>
        <li>developing new features, and keeping them free for everyone 🌈</li>
      </ul>
      <p class="body-1">Also, please <a href="https://github.com/photoprism/photoprism/stargazers" target="_blank" class="text-link">leave a star</a> on
        GitHub if you like this project.
        It provides additional motivation to keep going.</p>
      <p class="body-1">Thank you very much! ❤️</p>

      <h2 class="py-2 subheading">Roadmap</h2>
      <p class="body-1">Our vision is to provide the most user-friendly solution for browsing, organizing and sharing
        your personal photo collection.
        The <a href="https://github.com/photoprism/photoprism/projects/5" target="_blank" class="text-link">roadmap</a> shows what tasks are in progress,
        what needs testing, and which feature requests are going to be implemented next.</p>
      <p class="body-1">Please give ideas you like a thumbs-up, so that we know what is most popular.
        Ideas backed by one or more eligible <a
            href="https://github.com/photoprism/photoprism/blob/develop/SPONSORS.md" target="_blank" class="text-link">sponsors</a> will be prioritized as
        well.</p>

      <h2 class="py-2 subheading">
        <translate>Trademarks</translate>
      </h2>
      <p class="body-1 pb-2">
        <translate>PhotoPrism® is a registered trademark of Michael Mayer.</translate>
        <translate>You may use it as required to describe our software, run your own server, for educational purposes, but not for offering commercial goods, products, or services without prior written permission. In other words, please ask.</translate>
      </p>

      <p class="text-xs-center pt-3 mb-2">
        <router-link to="/about/license">
          <img src="/static/img/badge-agpl.svg" alt="License AGPL v3" style="max-width:100%;"/>
        </router-link>
        <a target="_blank" href="https://docs.photoprism.org/" rel="nofollow"><img src="/static/img/badge-docs.svg"
                                                                                   alt="Documentation"
                                                                                   style="max-width:100%;"></a>
        <a target="_blank" href="https://gitter.im/browseyourlife/community" rel="nofollow"><img
            src="/static/img/badge-chat.svg" alt="Community Chat" style="max-width:100%;"></a>
        <a target="_blank" href="https://twitter.com/browseyourlife" rel="nofollow"><img
            src="/static/img/badge-twitter.svg" alt="Twitter" style="max-width:100%;"></a>
      </p>
    </v-container>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
export default {
  name: 'PPageAbout',
  data() {
    return {};
  },
  methods: {},
};
</script>
