<template>
  <div class="p-tab p-tab-discover-memories">
    <div class="pa-2 text-center">
      <p class="text-subtitle-1 pb-2">
        {{ $gettext("On This Day") }}
      </p>

      <p class="text-body-1 pb-6">
        {{ memoryDateLabel }}
      </p>

      <div v-if="loading" class="py-8">
        <v-progress-circular color="primary" indeterminate></v-progress-circular>
      </div>

      <v-alert v-else-if="error" color="warning" variant="outlined" icon="mdi-alert-outline" class="mx-auto mb-6 text-start memories-alert">
        {{ $gettext("Something went wrong, try again") }}
      </v-alert>

      <v-alert v-else-if="memoryGroups.length === 0" color="surface-variant" variant="outlined" icon="mdi-image-off-outline" class="mx-auto mb-6 text-start memories-alert">
        {{ $gettext("No pictures found") }}
      </v-alert>

      <template v-else>
        <div class="d-flex justify-center mb-6">
          <v-btn color="button" rounded variant="flat" :to="{ name: 'browse', query: browseQuery() }">
            {{ $gettext("Browse Pictures") }}
          </v-btn>
        </div>

        <v-row class="text-start">
          <v-col v-for="group in memoryGroups" :key="group.year" cols="12" sm="6" lg="4">
            <v-card class="memory-card clickable" :to="{ name: 'browse', query: browseQuery(group.year) }" variant="tonal" color="surface-variant">
              <v-card-text class="pb-2">
                <div class="d-flex align-center justify-space-between">
                  <div class="text-h6">{{ group.year }}</div>
                  <v-chip size="small" color="primary" variant="tonal">
                    {{ group.count }}
                  </v-chip>
                </div>
              </v-card-text>

              <div class="memory-preview-grid px-4 pb-4">
                <div v-for="photo in group.preview" :key="photo.UID" class="memory-preview" :style="`background-image: url(${photo.thumbnailUrl('tile_500')})`"></div>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </template>
    </div>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import { Photo } from "model/photo";

export default {
  name: "PTabDiscoverMemories",
  data() {
    return {
      loading: true,
      error: false,
      memoryDate: DateTime.local().startOf("day"),
      memoryGroups: [],
    };
  },
  computed: {
    memoryDateLabel() {
      return this.memoryDate.toLocaleString({ month: "long", day: "numeric" });
    },
  },
  created() {
    this.loadMemories();
  },
  methods: {
    browseQuery(year) {
      const query = {
        month: this.memoryDate.month,
        day: this.memoryDate.day,
        order: "oldest",
      };

      if (year) {
        query.year = year;
      }

      return query;
    },
    buildGroups(models) {
      const groups = new Map();
      const currentYear = this.memoryDate.year;

      models.forEach((model) => {
        const year = Number(model?.Year);

        if (!Number.isInteger(year) || year <= 0 || year >= currentYear) {
          return;
        }

        if (!groups.has(year)) {
          groups.set(year, {
            year,
            count: 0,
            preview: [],
          });
        }

        const group = groups.get(year);
        group.count += 1;

        if (group.preview.length < 4) {
          group.preview.push(model);
        }
      });

      return Array.from(groups.values()).sort((a, b) => b.year - a.year);
    },
    async loadMemories() {
      const params = {
        count: 200,
        day: this.memoryDate.day,
        merged: true,
        month: this.memoryDate.month,
        order: "oldest",
      };

      if (this.$config.getSettings()?.features?.private) {
        params.public = "true";
      }

      const memories = [];
      let offset = 0;

      try {
        for (let page = 0; page < 25; page++) {
          const response = await Photo.search({ ...params, offset });
          const models = Array.isArray(response.models) ? response.models : [];
          const total = Number.isFinite(response.count) ? response.count : models.length;
          const responseOffset = Number.isFinite(response.offset) ? response.offset : offset;
          const limit = Number.isFinite(response.limit) && response.limit > 0 ? response.limit : models.length;

          memories.push(...models);

          if (!models.length || limit <= 0 || memories.length >= total) {
            break;
          }

          offset = responseOffset + limit;
        }

        this.memoryGroups = this.buildGroups(memories);
      } catch {
        this.error = true;
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>

<style scoped>
.memories-alert {
  max-width: 720px;
}

.memory-card {
  height: 100%;
}

.memory-preview-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.memory-preview {
  aspect-ratio: 1;
  background-position: center;
  background-size: cover;
  border-radius: 12px;
}
</style>
