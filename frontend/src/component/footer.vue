<template>
  <v-card variant="flat" tile class="bg-background footer">
    <v-card-actions class="footer-actions">
      <v-row align="start" class="pt-6" no-gutters>
        <v-col cols="12" sm="6" class="text-start">
          <p class="text-body-2 text-selectable px-0 pb-2">
            <strong><router-link to="/about" class="text-link text-selectable">{{ about }}{{ getMembership() }}</router-link></strong>
            <span class="body-link text-selectable">Build&nbsp;<a href="https://docs.photoprism.app/release-notes/" target="_blank" :title="version" class="body-link">{{ build }}</a></span>
          </p>
        </v-col>
        <v-col cols="12" sm="6" class="hidden-xs text-center text-sm-right">
          <p class="px-0 pb-2 text-body-2">
            <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE" target="_blank" class="text-link">3rd-party software packages</a>
            <a href="https://www.photoprism.app/about/team/" target="_blank" class="body-link">© 2018-2024 PhotoPrism UG</a>
          </p>
        </v-col>
      </v-row>
    </v-card-actions>
  </v-card>
</template>

<script>
export default {
  name: "PAboutFooter",
  data() {
    const ver = this.$config.getVersion().split("-");
    const build = ver.slice(0, 2).join("-");
    const about = this.$config.getAbout();
    const membership = this.$config.getMembership();
    const customer = this.$config.getCustomer();

    return {
      rtl: this.$rtl,
      build: build,
      about: about,
      membership: membership,
      customer: customer,
      version: this.$config.getVersion(),
      isDemo: this.$config.isDemo(),
    };
  },
  methods: {
    getMembership() {
      if (this.isDemo) {
        return " Demo";
      }

      const tier = this.$config.getTier();
      if (tier < 4) {
        return " CE";
      } else if (tier === 4) {
        return " Essentials";
      }

      return "";
    },
  },
};
</script>
