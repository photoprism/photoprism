<template>
  <v-dialog :value="show" persistent max-width="575" class="modal-dialog sponsor-dialog" @keydown.esc="close">
    <v-card raised elevation="24">
      <v-card-title class="px-2 pb-0">
        <v-row class="px-2">
          <v-col cols="10">
            <h3 class="title mb-0">
              <translate>Support Our Mission</translate>
            </h3>
          </v-col>
          <v-col cols="2" class="text-right">
            <v-icon size="26" color="secondary-dark" v-html="'$vuetify.icons.prism'"></v-icon>
          </v-col>
        </v-row>
      </v-card-title>
      <v-card-text class="px-2">
        <v-row class="px-2">
          <v-col cols="12" class="py-1">
            <p class="body-2">
              <translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate>
            </p>
            <p class="body-1">
              <translate>Being 100% self-funded and independent, we can promise you that we will never sell your data and that we will always be transparent about our software and services.</translate>
            </p>
            <p class="body-1">
              <translate>You are welcome to contact us at membership@photoprism.app for questions regarding your membership.</translate>
            </p>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions class="pt-0 px-2">
        <v-row class="px-2">
          <v-col cols="12"  class="text-right py-2">
            <v-btn depressed color="secondary-light" class="action-close compact" @click.stop="close">
              <translate>No thanks</translate>
            </v-btn>
            <v-btn v-if="isPublic || !isAdmin" href="https://link.photoprism.app/personal-editions" target="_blank" depressed color="primary-button" class="white--text action-about compact">
              <translate>Learn more</translate>
            </v-btn>
            <v-btn v-else depressed color="primary-button" class="white--text action-upgrade compact" @click.stop="upgrade">
              <translate>Upgrade Now</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PSponsorDialog",
  props: {
    show: Boolean,
  },
  data() {
    return {
      isPublic: this.$config.isPublic(),
      isAdmin: this.$session.isAdmin(),
      isDemo: this.$config.isDemo(),
      isSponsor: this.$config.isSponsor(),
      host: window.location.host,
      rtl: this.$rtl,
    };
  },
  methods: {
    close() {
      this.$emit("close");
    },
    upgrade() {
      this.$router.push({ name: "upgrade" });
      this.$emit("close");
    },
  },
};
</script>
