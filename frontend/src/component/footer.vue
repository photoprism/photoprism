<template>
  <v-card flat tile class="ma-0 pa-0 application p-about-footer">
    <v-card-actions class="px-4 py-2">
      <v-layout wrap align-top pt-3>
        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-selectable text-xs-center text-sm-left">
          <template v-if="sponsor"><router-link to="/about" class="text-link"><translate>Thank you for supporting PhotoPrism®</translate></router-link></template>
          <strong v-else><router-link to="/about" class="text-link"><translate>PhotoPrism® needs your support</translate></router-link></strong>
          <br><a href="https://docs.photoprism.app/release-notes/" target="_blank">Build {{ $config.get("version") }}</a>
        </v-flex>

        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-xs-center text-sm-right">
          <span class="hidden-sm-and-down">
            <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE"
               target="_blank" class="text-link">3rd-party software packages</a><br>
          </span>
          <a href="https://photoprism.app/team/" target="_blank">© 2018-2022 PhotoPrism UG</a>
        </v-flex>
      </v-layout>
    </v-card-actions>
  </v-card>
</template>

<script>
export default {
  name: 'PAboutFooter',
  data() {
    return {
      rtl: this.$rtl,
      sponsor: this.$config.isSponsor(),
    };
  },
  methods: {},
};
</script>
