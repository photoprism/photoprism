<template>
  <div class="auth-footer">
    <footer>
      <v-row align="start" class="pa-0 ma-0">
        <v-col xs="12" class="pa-0 text-subtitle-2 text-selectable text-start hidden-xs">
          {{ about }}
        </v-col>
        <v-col v-if="legalInfo" xs="12" class="pa-0 text-subtitle-2 text-center text-sm-end">
          <a v-if="legalUrl" :href="legalUrl" target="_blank" class="text-link">{{ legalInfo }}</a>
          <span v-else>{{ legalInfo }}</span>
        </v-col>
        <v-col v-else-if="caption" xs="12" class="pa-0 text-subtitle-2 text-selectable text-center text-sm-end">
          <strong>{{ caption }}</strong>
        </v-col>
        <v-col v-else xs="12" class="pa-0 text-subtitle-2 text-selectable text-center text-sm-end">
          <router-link to="/about" class="text-link text-ltr">
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
      rtl: this.$isRtl,
    };
  },
  methods: {},
};
</script>
