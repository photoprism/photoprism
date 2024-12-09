<template>
  <div class="auth-footer">
    <footer>
      <v-row align="start" class="pa-0 ma-0">
        <v-col cols="12" sm="6" class="pa-0 text-subtitle-2 text-selectable text-center text-sm-left">
          {{ about }}
        </v-col>

        <v-col v-if="legalInfo" cols="12" sm="6" class="pa-0 text-subtitle-2 text-center text-sm-right">
          <a v-if="legalUrl" :href="legalUrl" target="_blank" class="text-link">{{ legalInfo }}</a>
          <span v-else>{{ legalInfo }}</span>
        </v-col>
        <v-col v-else-if="caption" cols="12" sm="6" class="pa-0 text-subtitle-2 text-selectable text-center text-sm-right">
          <strong>{{ caption }}</strong>
        </v-col>
        <v-col v-else cols="12" sm="6" class="pa-0 text-subtitle-2 text-selectable text-center text-sm-right">
          <router-link to="/about" class="text-link">
            <span class="text-white">Made with ❤️ in Berlin</span>
          </router-link>
        </v-col>
      </v-row>
    </footer>
  </div>
</template>
<script>
export default {
  name: "PAuthFooter",
  data() {
    const config = this.$config;
    return {
      about: config.getAbout(),
      sponsor: config.isSponsor(),
      caption: config.values.siteCaption ? config.values.siteCaption : config.values.siteTitle,
      legalUrl: config.values.legalUrl,
      legalInfo: config.values.legalInfo,
      config: config.values,
      rtl: this.$rtl,
    };
  },
  methods: {},
};
</script>
