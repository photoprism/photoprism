<template>
  <div id="photoprism-help" class="p-page p-page-help">
    <v-toolbar flat color="secondary" :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
      <v-toolbar-title>
        <translate>Help</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://www.photoprism.app/" target="_blank" class="action-info" :title="$gettext('About')">
        <!-- TODO: change this icon -->
        <v-icon size="26">info</v-icon>
      </v-btn>
    </v-toolbar>

    <v-container fluid class="px-6 py-2">
      <p-help-websockets></p-help-websockets>
    </v-container>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import PHelpWebsockets from "./help/websockets.vue";

export default {
  name: "PPageHelp",
  components: {
    PHelpWebsockets,
  },
  data() {
    return {
      topic: this.$route.params.pathMatch,
    };
  },
};
</script>
