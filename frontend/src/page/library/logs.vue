<template>
  <div class="p-tab p-tab-logs pa-4 fill-height">
    <v-row class="d-flex align-stretch">
      <v-col cols="12" class="grow pa-2 bg-terminal elevation-0 p-logs" style="overflow: auto">
        <div v-if="logs.length === 0" class="p-log-empty flex-grow-1">{{ $gettext(`Nothing to see here yet.`) }}</div>
        <div
          v-for="log in logs"
          :key="log.id"
          class="p-log-message text-selectable text-break"
          :class="'p-log-' + log.level"
        >
          {{ formatTime(log.time) }} {{ formatLevel(log.level) }} <span>{{ log.message }}</span>
        </div>
      </v-col>
    </v-row>
  </div>
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
  methods: {
    formatLevel(s) {
      if (!s) {
        return "INFO";
      }

      return s.substring(0, 4).toUpperCase();
    },
    formatTime(s) {
      if (!s) {
        return "0000-00-00 00:00:00";
      }

      return DateTime.fromISO(s).toFormat("yyyy-LL-dd HH:mm:ss");
    },
  },
};
</script>
