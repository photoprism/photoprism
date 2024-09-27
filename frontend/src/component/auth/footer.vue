<template>
  <div class="auth-footer">
    <footer>
      <v-row align="start" class="pa-0 ma-0">
        <v-col cols="12" sm="6" class="pa-0 body-2 text-selectable text-xs-center white--text text-sm-left">
          {{ about }}
        </v-col>

        <v-col v-if="legalInfo" cols="12" sm="6" class="pa-0 body-2 text-xs-center text-sm-right white--text">
          <a v-if="legalUrl" :href="legalUrl" target="_blank" class="text-link" :style="`color: ${colors.link}!important`">{{ legalInfo }}</a>
          <span v-else>{{ legalInfo }}</span>
        </v-col>
        <v-col v-else-if="caption" cols="12" sm="6" class="pa-0 body-2 text-selectable text-xs-center text-sm-right white--text">
          <strong>{{ caption }}</strong>
        </v-col>
        <v-col v-else cols="12" sm="6" class="pa-0 body-2 text-selectable text-xs-center text-sm-right white--text">
          <router-link to="/about" class="text-link">
            <span class="white--text">Made with ❤️ in Berlin</span>
          </router-link>
        </v-col>
      </v-row>
    </footer>
  </div>
</template>
<script>
export default {
  name: "PAuthFooter",
  props: {
    colors: {
      type: Object,
      default: () => {
        return {
          accent: "#05dde1",
          primary: "#00a6a9",
          secondary: "#505050",
          link: "#c8e3e7",
        };
      },
    },
  },
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
