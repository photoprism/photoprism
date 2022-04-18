<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="modal-dialog sponsor-dialog" @keydown.esc="close">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs9>
            <h3 class="title mb-0">
              <translate>Become a sponsor</translate>
            </h3>
          </v-flex>
          <v-flex xs3 text-xs-right>
            <v-icon color="secondary-dark">$vuetify.icons.sponsor</v-icon>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text>
        <p class="body-1">
          <translate>This currently is a sponsor feature to thank everyone who supports the development of this application.</translate>
          <translate>We'll let you know how to enable it when you sign up on Patreon or GitHub Sponsors.</translate>
        </p>
        <p class="body-1">
          <translate>Your continued support helps us provide regular updates and services like world maps.</translate>
        </p>
        <p class="body-1">
          <translate>Feel free to contact us at hello@photoprism.app if you have any questions.</translate>
        </p>
      </v-card-text>
      <v-card-actions class="pt-0">
        <v-layout row wrap class="px-2">
          <v-flex xs12 sm4 text-xs-right text-sm-left class="py-2">
            <v-btn small depressed color="secondary-light"
                   class="action-close"
                   @click.stop="close">
              <translate>No thanks</translate>
            </v-btn>
          </v-flex>
          <v-flex xs12 sm8 text-xs-right class="py-2">
            <v-btn small depressed color="primary-button" class="white--text action-close"
                   @click.stop="signIn">
              <translate>I'm a sponsor</translate>
            </v-btn>
            <v-btn small depressed color="primary-button" class="white--text action-close"
                   @click.stop="signUp">
              <translate>Sign Up</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: 'PSponsorDialog',
  props: {
    show: Boolean,
  },
  data() {
    return {
      host: window.location.host,
      rtl: this.$rtl,
    };
  },
  methods: {
    close() {
      this.$emit('close');
    },
    signIn() {
      window.open("https://photoprism.app/contact", "_blank");
      this.$emit('close');
    },
    signUp() {
      window.open("https://link.photoprism.app/patreon", "_blank");
      this.$emit('close');
    },
  },
};
</script>
