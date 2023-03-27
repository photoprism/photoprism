<template>
  <v-card flat tile class="application footer">
    <v-card-actions class="footer-actions">
      <v-layout wrap align-top pt-3>
        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-selectable text-xs-left">
          <strong><router-link to="/about" class="text-link text-selectable">{{ about }}&nbsp;{{ getMembership() }}</router-link></strong>
          <span class="body-link text-selectable">Build&nbsp;<a href="https://docs.photoprism.app/release-notes/" target="_blank" :title="version" class="body-link">{{ build }}</a></span>
        </v-flex>

        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-xs-center text-sm-right">
          <div class="hidden-xs-only">
            <a v-if="evaluation" href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE"
               target="_blank" class="text-link">3rd-party software packages</a>
            <a v-else href="https://my.photoprism.app/" target="_blank" class="text-link">Licensed to {{ customer }}</a>
            <a href="https://www.photoprism.app/about/team/" target="_blank" class="body-link">© 2018-2023 PhotoPrism UG</a>
          </div>
        </v-flex>
      </v-layout>
    </v-card-actions>
  </v-card>
</template>

<script>
export default {
  name: 'PAboutFooter',
  data() {
    const ver = this.$config.getVersion().split("-");
    const build = ver.slice(0, 2).join("-");
    const about = this.$config.getAbout();
    const membership = this.$config.getMembership();
    const customer = this.$config.getCustomer();
    const evaluation = !customer;

    return {
      rtl: this.$rtl,
      build: build,
      about: about,
      membership: membership,
      customer: customer,
      evaluation: evaluation,
      version: this.$config.getVersion(),
      sponsor: this.$config.isSponsor(),
    };
  },
  methods: {
    getMembership() {
      const m = this.$config.getMembership();
      switch (m) {
        case "":
        case "ce":
          return "Community Edition";
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
