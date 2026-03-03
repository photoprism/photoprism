<template>
  <div class="p-photos p-photo-view-timeline">
    <div v-if="photos.length === 0" class="pa-3">
      <v-alert color="surface-variant" :icon="isSharedView ? 'mdi-image-off' : 'mdi-lightbulb-outline'" class="no-results" variant="outlined">
        <div v-if="filter.order === 'edited'" class="font-weight-bold">
          {{ $gettext(`No recently edited pictures`) }}
        </div>
        <div v-else class="font-weight-bold">
          {{ $gettext(`No pictures found`) }}
        </div>
        <div class="mt-2">
          {{ $gettext(`Try again using other filters or keywords.`) }}
          <template v-if="!isSharedView">
            {{ $gettext(`In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.`) }}
          </template>
        </div>
      </v-alert>
    </div>
    <div v-else ref="container" class="timeline-container" @scroll="onScroll">
      <!-- Timeline Navigation Sidebar -->
      <div class="timeline-nav">
        <div class="timeline-nav-title">{{ $gettext('Timeline') }}</div>
        <div class="timeline-years">
          <button
            v-for="year in availableYears"
            :key="year"
            :class="{ active: currentYear === year }"
            class="timeline-year-btn"
            @click="scrollToYear(year)"
          >
            {{ year }}
          </button>
        </div>
      </div>

      <!-- Timeline Content -->
      <div class="timeline-content">
        <div
          v-for="(group, year) in groupedPhotos"
          :key="year"
          :ref="el => { if (el) yearRefs[year] = el }"
          class="timeline-year-section"
          :data-year="year"
        >
          <div class="timeline-year-header">
            <span class="timeline-year-label">{{ year }}</span>
            <span class="timeline-year-count">{{ group.length }} {{ $gettext('photos') }}</span>
          </div>

          <div class="timeline-months">
            <div
              v-for="(monthGroup, month) in groupByMonth(group)"
              :key="month"
              class="timeline-month-section"
            >
              <div class="timeline-month-header">
                <span class="timeline-month-label">{{ formatMonth(month) }}</span>
                <span class="timeline-month-count">{{ monthGroup.length }}</span>
              </div>

              <div class="timeline-photos">
                <div
                  v-for="m in monthGroup"
                  :key="m.ID"
                  class="timeline-photo-item"
                  @click="openPhoto(m)"
                >
                  <div
                    :title="m.Title"
                    :style="`background-image: url(${m.thumbnailUrl('tile_500')})`"
                    class="timeline-photo-thumb"
                  >
                    <div v-if="m.Type === 'video'" class="timeline-photo-video">
                      <i class="mdi mdi-play-circle" />
                      <span>{{ m.getDurationInfo() }}</span>
                    </div>
                    <div class="timeline-photo-date">
                      {{ formatDay(m.Day, m.Month) }}
                    </div>
                  </div>
                  <div class="timeline-photo-info">
                    <div v-if="m.Title" class="timeline-photo-title">{{ m.Title }}</div>
                    <div class="timeline-photo-meta">
                      <span v-if="m.CameraID > 1">{{ m.getCameraInfo() }}</span>
                      <span v-if="m.LocationID">{{ m.locationInfo() }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Photo } from "model/photo";

export default {
  name: "PPhotoViewTimeline",
  props: {
    photos: {
      type: Array,
      default: () => [],
    },
    filter: {
      type: Object,
      default: () => ({}),
    },
    context: {
      type: String,
      default: "",
    },
    selectMode: {
      type: Boolean,
      default: false,
    },
    isSharedView: {
      type: Boolean,
      default: false,
    },
    openPhoto: {
      type: Function,
      default: () => {},
    },
    editPhoto: {
      type: Function,
      default: () => {},
    },
  },
  data() {
    return {
      currentYear: null,
      yearRefs: {},
    };
  },
  computed: {
    groupedPhotos() {
      const groups = {};
      this.photos.forEach((photo) => {
        const year = photo.Year || new Date(photo.CreatedAt).getFullYear();
        if (!groups[year]) {
          groups[year] = [];
        }
        groups[year].push(photo);
      });
      // Sort years descending
      return Object.keys(groups)
        .sort((a, b) => b - a)
        .reduce((acc, year) => {
          acc[year] = groups[year];
          return acc;
        }, {});
    },
    availableYears() {
      return Object.keys(this.groupedPhotos).map(Number).sort((a, b) => b - a);
    },
  },
  mounted() {
    this.updateCurrentYear();
    if (this.availableYears.length > 0) {
      this.currentYear = this.availableYears[0];
    }
  },
  methods: {
    groupByMonth(photos) {
      const groups = {};
      photos.forEach((photo) => {
        const month = photo.Month || new Date(photo.CreatedAt).getMonth() + 1;
        if (!groups[month]) {
          groups[month] = [];
        }
        groups[month].push(photo);
      });
      // Sort months descending within year
      return Object.keys(groups)
        .sort((a, b) => b - a)
        .reduce((acc, month) => {
          acc[month] = groups[month];
          return acc;
        }, {});
    },
    formatMonth(month) {
      const months = [
        this.$gettext("January"),
        this.$gettext("February"),
        this.$gettext("March"),
        this.$gettext("April"),
        this.$gettext("May"),
        this.$gettext("June"),
        this.$gettext("July"),
        this.$gettext("August"),
        this.$gettext("September"),
        this.$gettext("October"),
        this.$gettext("November"),
        this.$gettext("December"),
      ];
      return months[month - 1] || month;
    },
    formatDay(day, month) {
      if (!day) return "";
      return `${day}`;
    },
    scrollToYear(year) {
      const element = this.yearRefs[year];
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
        this.currentYear = year;
      }
    },
    onScroll() {
      this.updateCurrentYear();
    },
    updateCurrentYear() {
      if (!this.$refs.container) return;

      const container = this.$refs.container;
      const scrollTop = container.scrollTop;
      const containerRect = container.getBoundingClientRect();

      // Find which year section is currently in view
      for (const year of this.availableYears) {
        const element = this.yearRefs[year];
        if (element) {
          const rect = element.getBoundingClientRect();
          const relativeTop = rect.top - containerRect.top;
          if (relativeTop <= 100 && relativeTop + rect.height > 0) {
            this.currentYear = year;
            break;
          }
        }
      }
    },
  },
};
</script>

<style scoped>
.timeline-container {
  display: flex;
  height: calc(100vh - 120px);
  overflow: hidden;
}

.timeline-nav {
  width: 80px;
  background: var(--v-surface-variant);
  border-right: 1px solid var(--v-border-color);
  padding: 16px 8px;
  overflow-y: auto;
  flex-shrink: 0;
}

.timeline-nav-title {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--v-secondary-text);
  margin-bottom: 12px;
  text-align: center;
}

.timeline-years {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.timeline-year-btn {
  padding: 8px 4px;
  border: none;
  background: transparent;
  color: var(--v-secondary-text);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.timeline-year-btn:hover {
  background: var(--v-hover);
  color: var(--v-primary-text);
}

.timeline-year-btn.active {
  background: var(--v-primary);
  color: white;
}

.timeline-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.timeline-year-section {
  margin-bottom: 32px;
}

.timeline-year-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--v-border-color);
}

.timeline-year-label {
  font-size: 32px;
  font-weight: 300;
  color: var(--v-primary-text);
}

.timeline-year-count {
  font-size: 14px;
  color: var(--v-secondary-text);
}

.timeline-month-section {
  margin-bottom: 24px;
  padding-left: 16px;
  border-left: 3px solid var(--v-border-color);
}

.timeline-month-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.timeline-month-label {
  font-size: 18px;
  font-weight: 500;
  color: var(--v-primary-text);
}

.timeline-month-count {
  font-size: 12px;
  color: var(--v-secondary-text);
  background: var(--v-surface-variant);
  padding: 2px 8px;
  border-radius: 12px;
}

.timeline-photos {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.timeline-photo-item {
  cursor: pointer;
  transition: transform 0.2s;
}

.timeline-photo-item:hover {
  transform: translateY(-2px);
}

.timeline-photo-thumb {
  aspect-ratio: 1;
  background-size: cover;
  background-position: center;
  border-radius: 8px;
  position: relative;
  overflow: hidden;
}

.timeline-photo-date {
  position: absolute;
  top: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.timeline-photo-video {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.timeline-photo-info {
  padding: 8px 4px;
}

.timeline-photo-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--v-primary-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.timeline-photo-meta {
  font-size: 11px;
  color: var(--v-secondary-text);
  margin-top: 2px;
}

@media (max-width: 768px) {
  .timeline-nav {
    width: 60px;
    padding: 12px 4px;
  }

  .timeline-year-btn {
    font-size: 11px;
    padding: 6px 2px;
  }

  .timeline-content {
    padding: 16px;
  }

  .timeline-year-label {
    font-size: 24px;
  }

  .timeline-photos {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 12px;
  }
}
</style>
