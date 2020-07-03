<template>
  <div class="p-page p-page-help" id="photoprism-help">
    <v-toolbar flat color="secondary">
      <v-toolbar-title>
        <translate>Help</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon @click.stop="bugReport" class="action-support">
        <v-icon>bug_report</v-icon>
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
        methods: {
            bugReport() {
                console.log("BUG REPORT");
            }
        },
    };
</script>
