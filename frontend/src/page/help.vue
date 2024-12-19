<template>
  <div class="p-page p-page-help">
    <v-toolbar flat color="secondary" :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
      <v-toolbar-title class="flex-grow-1">
        <translate>Help</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://docs.photoprism.app/getting-started/troubleshooting/" target="_blank" class="action-info mx-2" :title="$gettext('Learn more')">
        <v-icon icon="mdi-book-open-page-variant" size="26" color="surface-variant"></v-icon>
      </v-btn>
    </v-toolbar>

    <v-container fluid class="px-6 py-4">
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
