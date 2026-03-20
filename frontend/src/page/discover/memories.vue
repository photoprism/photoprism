<template>
  <div class="p-tab p-tab-discover-memories">
    <div class="pa-2 text-center">
      <p class="text-subtitle-1 pb-4">
        {{ $gettext(`Photos taken on this day in previous years.`) }}
      </p>

      <div v-if="loading" class="pa-8 text-center">
        <v-progress-circular indeterminate color="primary"></v-progress-circular>
      </div>

      <div v-else-if="memories.length === 0" class="pa-8 text-center">
        <p class="text-body-1 text-medium-emphasis">
          {{ $gettext(`No memories found for today. Take some photos and come back next year!`) }}
        </p>
      </div>

      <v-row v-else class="p-memories">
        <v-col v-for="memory in memories" :key="memory.Year" cols="12" sm="6" md="4" class="pa-2">
          <v-card
            :to="{ name: 'browse', query: { month: memory.Month, day: memory.Day, year: memory.Year, order: 'oldest' } }"
            class="clickable"
            elevation="2"
          >
            <v-card-title class="text-h6">
              {{ yearsAgoLabel(memory.Year) }}
            </v-card-title>
            <v-card-subtitle>
              {{ memory.Year }} &mdash; {{ formattedDate(memory.Month, memory.Day) }}
            </v-card-subtitle>
            <v-card-text>
              <span class="text-body-2">
                {{ memory.PhotoCount }}
                {{ memory.PhotoCount === 1 ? $gettext(`photo`) : $gettext(`photos`) }}
              </span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script>
import $api from "common/api";

export default {
  name: "PTabDiscoverMemories",
  data() {
    return {
      loading: true,
      memories: [],
    };
  },
  created() {
    this.loadMemories();
  },
  methods: {
    loadMemories() {
      this.loading = true;
      $api
        .get("memories/on-this-day")
        .then((r) => {
          this.memories = r.data && Array.isArray(r.data) ? r.data : [];
        })
        .catch(() => {
          this.memories = [];
        })
        .finally(() => {
          this.loading = false;
        });
    },
    yearsAgoLabel(year) {
      const currentYear = new Date().getFullYear();
      const diff = currentYear - year;
      if (diff === 1) {
        return this.$gettext(`1 year ago today`);
      }
      return `${diff} ` + this.$gettext(`years ago today`);
    },
    formattedDate(month, day) {
      const date = new Date(2000, month - 1, day);
      return date.toLocaleDateString(undefined, { month: "long", day: "numeric" });
    },
  },
};
</script>
