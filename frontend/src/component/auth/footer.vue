<template>
  <div class="auth-footer">
    <footer v-if="sponsor">
      <v-layout wrap align-top pa-0 ma-0>
        <v-flex xs12 sm6 class="pa-0 body-2 text-selectable text-xs-center white--text text-sm-left">
          {{ $config.getAbout() }}
        </v-flex>

        <v-flex v-if="config.legalInfo" xs12 sm6 class="pa-0 body-2 text-xs-center text-sm-right white--text">
          <a v-if="config.legalUrl" :href="config.legalUrl" target="_blank" class="text-link"
             :style="`color: ${colors.link}!important`">{{ config.legalInfo }}</a>
          <span v-else>{{ config.legalInfo }}</span>
        </v-flex>
        <v-flex v-else xs12 class="pa-0 body-2 text-selectable text-xs-center white--text text-sm-right sm6">
          <strong>{{ config.siteCaption ? config.siteCaption : config.siteTitle }}</strong>
        </v-flex>
      </v-layout>
    </footer>
    <footer v-else>
      <v-layout wrap align-top pa-0 ma-0>
        <v-flex xs12 sm6 class="pa-0 body-2 text-xs-center text-sm-left white--text text-selectable">
          <strong>{{ config.siteTitle }}</strong> – {{ config.siteCaption }}
        </v-flex>
        <v-flex xs12 sm6 class="pa-0 body-2 text-selectable text-xs-center text-sm-right">
          <router-link to="/about" class="text-link"><span class="white--text">Made with ❤️ in Berlin</span></router-link>
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
    return {
      sponsor: this.$config.isSponsor(),
      config: this.$config.values,
      rtl: this.$rtl,
    };
  },
  methods: {}
};

</script>