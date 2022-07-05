<template>
  <v-card flat tile class="application footer">
    <v-card-actions class="footer-actions">
      <v-layout wrap align-top pt-3>
        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-selectable text-xs-left">
          <strong><router-link to="/about" class="text-link">{{ $config.getEdition() }}</router-link></strong>
          <span class="body-link text-selectable">Build <a href="https://docs.photoprism.app/release-notes/" target="_blank" :title="version" class="body-link">{{ build }}</a></span>
        </v-flex>

        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-xs-center text-sm-right">
          <div class="hidden-xs-only">
            <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE"
               target="_blank" class="text-link">3rd-party software packages</a>
            <a href="https://photoprism.app/team/" target="_blank" class="body-link">© 2018-2022 PhotoPrism UG</a>
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

    return {
      rtl: this.$rtl,
      build: build,
      version: this.$config.getVersion(),
      sponsor: this.$config.isSponsor(),
    };
  },
  methods: {},
};
</script>
