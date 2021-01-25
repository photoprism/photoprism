<template>
  <v-dialog v-model="show" lazy persistent max-width="500" class="modal-dialog sponsor-dialog" @keydown.esc="close">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs9>
            <h3 class="title mb-0">
              <translate>Let's Join Forces</translate>
            </h3>
          </v-flex>
          <v-flex xs3 text-xs-right>
            <v-icon color="secondary-dark">favorite</v-icon>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text>
        <p class="body-1">
          This is an early-access feature to say thanks to our sponsors and contributors.
          We'll let you know how to enable it when you sign up on
          <a href="https://www.patreon.com/photoprism" class="text-link" target="_blank">Patreon</a> or
          <a href="https://github.com/sponsors/photoprism" class="text-link" target="_blank">GitHub Sponsors</a>.
        </p>
        <p class="body-1">
          Your continuous support helps...
        </p>
        <ul>
          <li>pay for operating expenses and external services like satellite maps</li>
          <li>developing new features and keeping them <span title="At a later time. Fair is fair.">free for everyone</span></li>
        </ul>
        <p class="body-1 pt-3">
          Please contact us at <a href="mailto:hello@photoprism.app" class="text-link" target="_blank">hello@photoprism.app</a>
          if you have questions or need help.
        </p>
      </v-card-text>
      <v-card-actions class="pt-0">
        <v-layout row wrap class="px-2">
          <v-flex xs12 sm6 text-xs-right class="py-2">
            <v-btn small depressed color="secondary-light"
                   class="action-close"
                   @click.stop="close">
              <translate>No thanks</translate>
            </v-btn>
          </v-flex>
          <v-flex xs12 sm6 text-xs-right class="py-2">
            <v-btn small depressed color="primary-button" class="white--text action-close"
                   @click.stop="signIn">
              <translate>I'm a sponsor</translate>
            </v-btn>
            <v-btn small depressed color="primary-button" class="white--text action-close"
                   @click.stop="signUp">
              Sign Up
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
      window.open("https://www.patreon.com/photoprism", "_blank");
      this.$emit('close');
    },
  },
};
</script>
