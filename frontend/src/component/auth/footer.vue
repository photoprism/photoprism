<template>
  <div class="auth-footer">
    <footer>
      <v-layout wrap align-top pa-0 ma-0>
        <v-flex xs12 sm6 class="pa-0 body-2 text-selectable text-xs-center white--text text-sm-left">
          {{ about }}
        </v-flex>

        <v-flex v-if="legalInfo" xs12 sm6 class="pa-0 body-2 text-xs-center text-sm-right white--text">
          <a v-if="legalUrl" :href="legalUrl" target="_blank" class="text-link"
             :style="`color: ${colors.link}!important`">{{ legalInfo }}</a>
          <span v-else>{{ legalInfo }}</span>
        </v-flex>
        <v-flex v-else-if="caption" xs12 sm6
                class="pa-0 body-2 text-selectable text-xs-center text-sm-right white--text">
          <strong>{{ caption }}</strong>
        </v-flex>
        <v-flex v-else xs12 sm6 class="pa-0 body-2 text-selectable text-xs-center text-sm-right white--text">
          <router-link to="/about" class="text-link"><span class="white--text">Made with ❤️ in Berlin</span>
          </router-link>
        </v-flex>
      </v-layout>
    </footer>
  </div>
</template>
<script>

export default {
  name: 'PAuthFooter',
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
  methods: {}
};

</script>