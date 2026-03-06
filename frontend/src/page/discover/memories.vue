<template>
  <div class="p-tab p-tab-memories">
    <v-container fluid class="pa-4">
      <v-row class="mb-4">
        <v-col cols="12">
          <h2 class="text-h5 mb-2">{{ $gettext('Memories') }}</h2>
          <p class="text-body-2 text-grey">{{ $gettext('On this day in past years') }}</p>
        </v-col>
      </v-row>

      <v-row v-if="loading">
        <v-col cols="12" class="text-center">
          <v-progress-circular indeterminate color="primary"></v-progress-circular>
        </v-col>
      </v-row>

      <v-row v-else-if="memories.length === 0">
        <v-col cols="12" class="text-center">
          <v-icon icon="mdi-calendar-clock" size="64" color="grey-lighten-2"></v-icon>
          <p class="text-body-1 mt-4">{{ $gettext('No memories for today') }}</p>
          <p class="text-body-2 text-grey">{{ $gettext('Photos taken on this day in past years will appear here') }}</p>
        </v-col>
      </v-row>

      <div v-else>
        <v-row v-for="memory in memories" :key="memory.year" class="mb-6">
          <v-col cols="12">
            <v-card>
              <v-card-title class="d-flex align-center">
                <v-icon icon="mdi-calendar-star" class="mr-2"></v-icon>
                {{ memory.title }}
                <v-spacer></v-spacer>
                <v-chip size="small" color="primary">{{ memory.count }} {{ $gettext('photos') }}</v-chip>
              </v-card-title>
              <v-card-subtitle>
                {{ memory.month }}/{{ memory.day }}/{{ memory.year }}
              </v-card-subtitle>
              <v-card-text>
                <v-row>
                  <v-col 
                    v-for="(photo, index) in memory.photos.slice(0, 6)" 
                    :key="index"
                    cols="6" 
                    sm="4" 
                    md="3" 
                    lg="2"
                  >
                    <v-img
                      :src="photo.Thumb"
                      :alt="photo.Title"
                      aspect-ratio="1"
                      cover
                      class="rounded memory-thumbnail"
                      @click="openPhoto(memory.photos, index)"
                    ></v-img>
                  </v-col>
                </v-row>
                <v-btn
                  v-if="memory.photos.length > 6"
                  variant="text"
                  block
                  class="mt-2"
                  @click="viewAll(memory)"
                >
                  {{ $gettext('View all') }} {{ memory.photos.length }} {{ $gettext('photos') }}
                </v-btn>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </div>
    </v-container>
  </div>
</template>

<script>
import { Photo } from "model/photo";

export default {
  name: "PTabDiscoverMemories",
  data() {
    return {
      loading: false,
      memories: [],
    };
  },
  mounted() {
    this.loadMemories();
  },
  methods: {
    async loadMemories() {
      this.loading = true;
      try {
        const response = await this.$api.get("/memories");
        this.memories = response.data || [];
      } catch (error) {
        console.error("Failed to load memories:", error);
        this.$notify.error(this.$gettext("Failed to load memories"));
      } finally {
        this.loading = false;
      }
    },
    openPhoto(photos, index) {
      // Open photo viewer
      const photoModels = photos.map(p => new Photo(p));
      this.$viewer.open(photoModels, index);
    },
    viewAll(memory) {
      // Navigate to search results for this date
      const query = {
        month: memory.month,
        day: memory.day,
        year: memory.year,
      };
      this.$router.push({ path: "/browse", query });
    },
  },
};
</script>

<style scoped>
.memory-thumbnail {
  cursor: pointer;
  transition: transform 0.2s;
}
.memory-thumbnail:hover {
  transform: scale(1.05);
}
</style>
