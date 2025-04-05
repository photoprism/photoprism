<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="575"
    class="p-dialog modal-dialog sponsor-dialog"
    @keydown.esc.exact="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-diamond-stone</v-icon>
        <h6 class="text-h6">{{ $gettext(`Support Our Mission`) }}</h6>
      </v-card-title>
      <v-card-text class="text-subtitle-2">{{
        $gettext(
          `Your continued support helps us provide regular updates and remain independent, so we can fulfill our mission and protect your privacy.`
        )
      }}</v-card-text>
      <v-card-text class="text-body-2">{{
        $gettext(
          `Being 100% self-funded and independent, we can promise you that we will never sell your data and that we will always be transparent about our software and services.`
        )
      }}</v-card-text>
      <v-card-text class="text-body-2">{{
        $gettext(`You are welcome to contact us at membership@photoprism.app for questions regarding your membership.`)
      }}</v-card-text>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-close" @click.stop="close">
          {{ $gettext(`No thanks`) }}
        </v-btn>
        <v-btn
          v-if="isPublic || !isAdmin"
          :href="links.compare"
          target="_blank"
          variant="flat"
          color="highlight"
          class="text-white action-about"
        >
          {{ $gettext(`Learn more`) }}
        </v-btn>
        <v-btn v-else variant="flat" color="highlight" class="text-white action-upgrade" @click.stop="upgrade">
          {{ $gettext(`Upgrade Now`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import links from "common/links";

export default {
  name: "PDialogSponsor",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      links,
      isPublic: this.$config.isPublic(),
      isAdmin: this.$session.isAdmin(),
      isDemo: this.$config.isDemo(),
      isSponsor: this.$config.isSponsor(),
      host: window.location.host,
      rtl: this.$isRtl,
    };
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
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
