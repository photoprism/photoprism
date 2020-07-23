<template>
  <div class="p-page p-page-help" id="photoprism-help">
    <v-toolbar flat color="secondary">
      <v-toolbar-title>
        <translate>Help</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon href="https://github.com/photoprism/photoprism" target="_blank" class="action-github" title="GitHub">
        <img src="/static/brands/github.svg" width="24" alt="GitHub">
      </v-btn>
    </v-toolbar>

    <v-container fluid class="px-4 py-2">
      <p-help-websockets></p-help-websockets>
    </v-container>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
    import PHelpWebsockets from "./help/websockets.vue"

    export default {
        name: 'p-page-help',
        components: {
            PHelpWebsockets
        },
        data() {
            return {
                topic: this.$route.params.pathMatch,
            };
        },
    };
</script>
