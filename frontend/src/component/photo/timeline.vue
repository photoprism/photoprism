<template>
  <aside v-if="!disabled" class="p-photo-timeline" :aria-label="$gettext('Timeline')" :aria-busy="loading ? 'true' : 'false'" data-testid="photo-timeline">
    <template v-if="visible">
      <div v-for="group in yearGroups" :key="group.year" class="p-photo-timeline__group">
        <div class="p-photo-timeline__year">{{ group.year }}</div>
        <div class="p-photo-timeline__months">
          <button
            v-for="bucket in group.buckets"
            :key="bucketKey(bucket)"
            type="button"
            class="p-photo-timeline__month"
            :class="{ 'is-active': isActive(bucket) }"
            :title="bucketAriaLabel(bucket)"
            :aria-label="bucketAriaLabel(bucket)"
            :aria-pressed="isActive(bucket) ? 'true' : 'false'"
            :data-testid="`timeline-month-${bucket.year}-${bucket.month}`"
            @click="selectBucket(bucket)"
          >
            <span class="p-photo-timeline__month-label">{{ monthLabel(bucket.month) }}</span>
            <span v-if="bucket.photoCount > 0" class="p-photo-timeline__month-count">{{ bucket.photoCount }}</span>
          </button>
        </div>
      </div>

      <div v-if="unknownDateCount > 0" class="p-photo-timeline__unknown" data-testid="timeline-unknown">
        {{ $gettext("Unknown") }}: {{ unknownDateCount }}
      </div>
    </template>
  </aside>
</template>

<script>
import Photo from "model/photo";
import { Info } from "luxon";

const MonthLabels = Info.months("short");

export default {
  name: "PPhotoTimeline",
  props: {
    filter: {
      type: Object,
      default: () => ({}),
    },
    staticFilter: {
      type: Object,
      default: () => ({}),
    },
    updateQuery: {
      type: Function,
      default: () => {},
    },
    refreshToken: {
      type: Number,
      default: 0,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["visibility"],
  data() {
    return {
      buckets: [],
      loading: false,
      requestId: 0,
      unknownDateCount: 0,
    };
  },
  computed: {
    requestSignature() {
      if (this.disabled) {
        return "disabled";
      }

      const params = this.timelineParams();

      return JSON.stringify({ ...params, refreshToken: this.refreshToken });
    },
    visible() {
      return !this.disabled && this.buckets.length > 0;
    },
    yearGroups() {
      const groups = [];
      const years = {};

      this.buckets.forEach((bucket) => {
        if (!years[bucket.year]) {
          years[bucket.year] = {
            year: bucket.year,
            buckets: [],
          };
          groups.push(years[bucket.year]);
        }

        years[bucket.year].buckets.push(bucket);
      });

      return groups;
    },
  },
  watch: {
    requestSignature: {
      immediate: true,
      handler() {
        if (this.disabled) {
          this.cancelTimeline();
          return;
        }

        this.loadTimeline();
      },
    },
  },
  beforeUnmount() {
    this.requestId++;
  },
  methods: {
    emitVisibility() {
      this.$emit("visibility", this.loading || this.buckets.length > 0);
    },
    cancelTimeline() {
      this.requestId++;
      this.loading = false;
      this.buckets = [];
      this.unknownDateCount = 0;
      this.emitVisibility();
    },
    timelineParams() {
      const params = {};

      if (this.filter && typeof this.filter === "object") {
        Object.assign(params, this.filter);
      }

      if (this.staticFilter && typeof this.staticFilter === "object") {
        Object.assign(params, this.staticFilter);
      }

      delete params.year;
      delete params.month;
      delete params.day;

      params.bucket = "month";

      return params;
    },
    loadTimeline() {
      const requestId = ++this.requestId;

      if (typeof Photo.timeline !== "function") {
        this.buckets = [];
        this.unknownDateCount = 0;
        this.loading = false;
        this.emitVisibility();
        return;
      }

      this.loading = true;
      this.emitVisibility();

      Promise.resolve()
        .then(() => Photo.timeline(this.timelineParams()))
        .then((response) => {
          if (requestId !== this.requestId) {
            return;
          }

          const data = response?.data || response || {};

          this.buckets = this.normalizeBuckets(data.buckets);
          this.unknownDateCount = Math.max(0, Number(data.unknownDateCount) || 0);
          this.emitVisibility();
        })
        .catch(() => {
          if (requestId !== this.requestId) {
            return;
          }

          this.buckets = [];
          this.unknownDateCount = 0;
          this.emitVisibility();
        })
        .finally(() => {
          if (requestId === this.requestId) {
            this.loading = false;
            this.emitVisibility();
          }
        });
    },
    normalizeBuckets(buckets) {
      if (!Array.isArray(buckets)) {
        return [];
      }

      return buckets
        .map((bucket) => ({
          key: bucket.key || bucket.Key || "",
          label: bucket.label || bucket.Label || "",
          year: Number(bucket.year ?? bucket.Year),
          month: Number(bucket.month ?? bucket.Month),
          photoCount: Number(bucket.photoCount ?? bucket.PhotoCount) || 0,
        }))
        .filter((bucket) => bucket.year > 0 && bucket.month > 0 && bucket.month <= 12);
    },
    bucketKey(bucket) {
      return bucket.key || `${bucket.year}-${bucket.month}`;
    },
    bucketTitle(bucket) {
      if (bucket.label) {
        return bucket.label;
      }

      return `${this.monthLabel(bucket.month)} ${bucket.year}`;
    },
    bucketAriaLabel(bucket) {
      return this.$gettext("Filter photos from %{date}, %{count} pictures", {
        date: this.bucketTitle(bucket),
        count: Number(bucket.photoCount) || 0,
      });
    },
    monthLabel(month) {
      return MonthLabels[month - 1] || String(month).padStart(2, "0");
    },
    isActive(bucket) {
      return Number(this.filter?.year) === bucket.year && Number(this.filter?.month) === bucket.month;
    },
    selectBucket(bucket) {
      this.updateQuery({
        year: bucket.year,
        month: bucket.month,
        day: 0,
      });
    },
  },
};
</script>
