<template>
  <div class="p-tab p-tab-discover-memories">
    <!-- Sub-tab strip: This Day / This Month -->
    <v-tabs v-model="activeMode" density="compact" class="memories-subtabs bg-transparent" grow>
      <v-tab value="day" @click="switchMode('day')">
        {{ $gettext(`This Day`) }}
      </v-tab>
      <v-tab value="month" @click="switchMode('month')">
        {{ $gettext(`This Month`) }}
      </v-tab>
    </v-tabs>

    <!-- View toggle: cards / mosaic -->
    <div class="memories-toolbar d-flex align-center justify-end px-3 py-1">
      <v-btn-toggle v-model="view" density="compact" rounded="0" mandatory>
        <v-btn value="cards" size="small" :title="$gettext('Cards')">
          <v-icon>mdi-view-module</v-icon>
        </v-btn>
        <v-btn value="mosaic" size="small" :title="$gettext('Mosaic')">
          <v-icon>mdi-view-comfy</v-icon>
        </v-btn>
      </v-btn-toggle>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="pa-6 text-center">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <!-- Empty state -->
    <div v-else-if="!loading && yearGroups.length === 0" class="pa-6 text-center">
      <v-icon size="48" class="mb-3 text-medium-emphasis">mdi-image-search-outline</v-icon>
      <p class="text-subtitle-1 text-medium-emphasis">
        <template v-if="activeMode === 'day'">
          {{ $gettext(`No memories found for this day in past years.`) }}
        </template>
        <template v-else>
          {{ $gettext(`No memories found for this month in past years.`) }}
        </template>
      </p>
    </div>

    <!-- Year groups -->
    <div v-else class="memories-groups pb-4">
      <div v-for="group in yearGroups" :key="group.year" class="memories-year-group">
        <!-- Sticky year header -->
        <div class="memories-year-header px-4 py-2 d-flex align-center">
          <span class="text-h6 font-weight-bold">{{ group.year }}</span>
          <span class="ml-2 text-caption text-medium-emphasis">
            {{ yearsAgoLabel(group.year) }} &bull; {{ group.photos.length }} {{ $gettext(`photos`) }}
          </span>
        </div>

        <!-- Mosaic view -->
        <div v-if="view === 'mosaic'" class="memories-mosaic pa-1">
          <div class="v-row ma-0">
            <div
              v-for="photo in group.photos"
              :key="photo.UID"
              class="v-col-6 v-col-sm-4 v-col-md-3 v-col-lg-2 pa-1"
            >
              <v-img
                :src="photo.thumbnailUrl('tile_224')"
                :alt="photo.Title"
                aspect-ratio="1"
                cover
                class="memories-thumb rounded cursor-pointer"
                @click="openPhoto(photo)"
              >
                <template #error>
                  <div class="memories-thumb-error d-flex align-center justify-center fill-height bg-surface-variant">
                    <v-icon>mdi-image-off</v-icon>
                  </div>
                </template>
              </v-img>
            </div>
          </div>
        </div>

        <!-- Cards view -->
        <div v-else class="memories-cards pa-2">
          <div class="v-row">
            <div
              v-for="photo in group.photos"
              :key="photo.UID"
              class="v-col-12 v-col-sm-6 v-col-md-4 v-col-lg-3 v-col-xl-2"
            >
              <v-card class="memories-card" elevation="1" @click="openPhoto(photo)">
                <v-img
                  :src="photo.thumbnailUrl('tile_224')"
                  :alt="photo.Title"
                  aspect-ratio="1"
                  cover
                  class="cursor-pointer"
                >
                  <template #error>
                    <div class="d-flex align-center justify-center fill-height bg-surface-variant">
                      <v-icon>mdi-image-off</v-icon>
                    </div>
                  </template>
                </v-img>
                <v-card-text v-if="photo.Title" class="pa-2 text-caption text-truncate">
                  {{ photo.Title }}
                </v-card-text>
              </v-card>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Photo } from "model/photo";

const BATCH = 120; // photos per year-query

export default {
  name: "PTabDiscoverMemories",

  data() {
    const now = new Date();
    return {
      loading: false,
      activeMode: "day", // "day" | "month"
      view: "cards",     // "cards" | "mosaic"
      currentDay: now.getDate(),
      currentMonth: now.getMonth() + 1,
      currentYear: now.getFullYear(),
      yearGroups: [],    // [{ year, photos[] }]
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    switchMode(mode) {
      if (this.activeMode !== mode) {
        this.activeMode = mode;
        this.load();
      }
    },

    async load() {
      this.loading = true;
      this.yearGroups = [];

      try {
        // Build base query params
        const params = {
          count: BATCH,
          offset: 0,
          order: "oldest",
          quality: 0, // include all quality levels
        };

        if (this.activeMode === "day") {
          // Exact day+month match across all years
          params.day = this.currentDay;
          params.month = this.currentMonth;
        } else {
          // Full month across all years
          params.month = this.currentMonth;
        }

        const resp = await Photo.search(params);
        if (!resp || !resp.models || resp.models.length === 0) {
          return;
        }

        // Group by year, exclude current year
        const groups = {};
        for (const photo of resp.models) {
          const year = photo.Year;
          if (!year || year <= 0 || year === this.currentYear) continue;
          if (!groups[year]) groups[year] = [];
          groups[year].push(photo);
        }

        // Sort years newest first
        this.yearGroups = Object.keys(groups)
          .map(Number)
          .sort((a, b) => b - a)
          .map((year) => ({ year, photos: groups[year] }));
      } catch (e) {
        if (process.env.NODE_ENV !== "production") {
          console.warn("Memories load failed", e);
        }
      } finally {
        this.loading = false;
      }
    },

    yearsAgoLabel(year) {
      const diff = this.currentYear - year;
      if (diff === 1) return this.$gettext("1 year ago");
      return diff + " " + this.$gettext("years ago");
    },

    openPhoto(photo) {
      // Navigate to browse view filtered to this photo's date
      // so the user can see it in full context with all the normal controls
      const takenAt = photo.TakenAtLocal || photo.TakenAt;
      if (takenAt) {
        const d = new Date(takenAt);
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, "0");
        const day = String(d.getDate()).padStart(2, "0");
        this.$router.push({
          name: "browse",
          query: { q: `taken:${y}-${m}-${day}` },
        });
      } else {
        this.$router.push({ name: "browse", query: { q: photo.UID } });
      }
    },
  },
};
</script>

<style lang="scss">
.p-tab-discover-memories {
  .memories-subtabs {
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  }

  .memories-toolbar {
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    min-height: 40px;
  }

  .memories-year-header {
    position: sticky;
    top: 0;
    z-index: 2;
    background: rgb(var(--v-theme-background));
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  }

  .memories-thumb {
    width: 100%;
    cursor: pointer;
    transition: opacity 0.15s;

    &:hover {
      opacity: 0.85;
    }
  }

  .memories-card {
    cursor: pointer;
    transition: box-shadow 0.15s;

    &:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2) !important;
    }
  }
}
</style>
