<template>
  <v-card flat tile class="ma-0 pa-0 application p-about-footer">
    <v-card-actions class="px-4 py-2">
      <v-layout wrap align-top pt-3>
        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-selectable">
          PhotoPrism® {{ $config.get("version") }}<br>
          <router-link to="/about">100% self-funded and independent</router-link>
        </v-flex>

        <v-flex xs12 sm6 class="px-0 pb-2 body-1 text-xs-left text-sm-right">
          <a href="https://photoprism.app/team/" target="_blank">© 2018-2022 Michael Mayer</a>
          <br>
          <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE"
             target="_blank">
            3rd-party software packages</a>
        </v-flex>
      </v-layout>
    </v-card-actions>
  </v-card>
</template>

<script>
export default {
  name: 'PAboutFooter',
  data() {
    return {};
  },
  methods: {},
};
</script>
