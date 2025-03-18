<template>
  <footer class="p-about-footer text-ltr">
    <div class="flex-fill text-sm-start">
      <strong>
        <router-link to="/about" class="text-link text-selectable">{{ about }} {{ getMembership() }}</router-link>
      </strong>
      <span :title="version" class="body-link text-selectable">
        <span class="cursor-copy" @click.stop.prevent="$util.copyText(about, version)">Build</span>
        <a href="https://docs.photoprism.app/release-notes/" target="_blank" class="body-link text-truncate">{{
          build
        }}</a>
      </span>
    </div>
    <div class="hidden-xs text-sm-end">
      <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE" target="_blank" class="text-link"
        >3rd-party software packages</a
      >
      <a href="https://www.photoprism.app/about/team/" target="_blank" class="body-link">© 2018-2025 PhotoPrism UG</a>
    </div>
  </footer>
</template>

<script>
export default {
  name: "PAboutFooter",
  data() {
    const about = this.$config.getAbout();
    const membership = this.$config.getMembership();
    const customer = this.$config.getCustomer();

    return {
      rtl: this.$isRtl,
      about: about,
      membership: membership,
      customer: customer,
      version: this.$config.getVersion(),
      isDemo: this.$config.isDemo(),
    };
  },
  computed: {
    build() {
      if (this.$vuetify.display.xs) {
        return this.$config.getVersion().split("-").slice(0, 1).join("-");
      } else {
        return this.$config.getVersion().split("-").slice(0, 2).join("-");
      }
    },
  },
  methods: {
    getMembership() {
      if (this.isDemo) {
        return "Demo";
      }

      const tier = this.$config.getTier();
      const edition = this.$config.getEdition();

      if (edition === "plus" && tier > 7) {
        return "Plus";
      } else if (edition === "plus" && tier > 5) {
        return "Essentials+";
      } else if (tier > 3) {
        return "Essentials";
      }

      return "CE";
    },
  },
};
</script>
