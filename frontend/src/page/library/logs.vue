<template>
  <v-container fluid fill-height class="pa-0 ma-0 p-tab p-tab-logs">
    <v-layout row wrap fill-height class="pa-0 ma-3">
      <v-flex grow xs12 class="pa-2 terminal elevation-0 p-logs" style="overflow: auto">
        <p v-if="logs.length === 0" class="p-log-empty">
          <translate>Nothing to see here yet.</translate>
        </p>
        <p v-for="(log, index) in logs" :key="index.id" class="p-log-message text-selectable" :class="'p-log-' + log.level">
          {{ formatTime(log.time) }} {{ level(log) }} <span>{{ log.message }}</span>
        </p>
      </v-flex>
    </v-layout>
  </v-container>
</template>

<script>
import { DateTime } from "luxon";

export default {
  name: "PTabLogs",
  data() {
    return {
      logs: this.$log.logs,
    };
  },
  created() {},
  methods: {
    level(log) {
      return log.level.substring(0, 4).toUpperCase();
    },
    formatTime(s) {
      if (!s) {
        return this.$gettext("Unknown");
      }

      return DateTime.fromISO(s).toFormat("yyyy-LL-dd HH:mm:ss");
    },
  },
};
</script>
