<template>
  <v-container fluid fill-height class="pa-0 ma-0 p-tab p-tab-logs">
    <v-row class="pa-0 ma-6 d-flex align-stretch">
      <v-col  cols="12" class="d-flex grow pa-2 terminal elevation-0 p-logs" style="overflow: auto">
        <p v-if="logs.length === 0" class="p-log-empty flex-grow-1">
          <translate>Nothing to see here yet.</translate>
        </p>
        <p v-for="(log, index) in logs" :key="index.id" class="p-log-message text-selectable flex-grow-1" :class="'p-log-' + log.level">
          {{ formatTime(log.time) }} {{ level(log) }} <span>{{ log.message }}</span>
        </p>
      </v-col>
    </v-row>
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
