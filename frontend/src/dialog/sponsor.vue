<template>
  <v-dialog :model-value="show" persistent max-width="575" class="modal-dialog sponsor-dialog" @keydown.esc="close">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-diamond-stone</v-icon>
        <h6 class="text-h6"><translate>Support Our Mission</translate></h6>
      </v-card-title>
      <v-card-text class="text-subtitle-2"><translate>Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.</translate></v-card-text>
      <v-card-text class="text-body-2"><translate>Being 100% self-funded and independent, we can promise you that we will never sell your data and that we will always be transparent about our software and services.</translate></v-card-text>
      <v-card-text class="text-body-2"><translate>You are welcome to contact us at membership@photoprism.app for questions regarding your membership.</translate></v-card-text>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-close" @click.stop="close">
          <translate>No thanks</translate>
        </v-btn>
        <v-btn v-if="isPublic || !isAdmin" href="https://link.photoprism.app/personal-editions" target="_blank" variant="flat" color="primary-button" class="text-white action-about">
          <translate>Learn more</translate>
        </v-btn>
        <v-btn v-else variant="flat" color="primary-button" class="text-white action-upgrade" @click.stop="upgrade">
          <translate>Upgrade Now</translate>
        </v-btn>
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
