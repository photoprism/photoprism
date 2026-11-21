<template>
  <div class="p-page p-page-discover" tabindex="-1">
    <div class="p-page__navigation">
      <v-tabs
        v-model="active"
        elevation="0"
        show-arrows
        class="bg-transparent"
        bg-color="secondary"
        :height="$vuetify.display.smAndDown ? 48 : 64"
      >
        <v-tab id="tab-discover-day" ripple @click="changePath('/discover')">
          <span class="d-none d-md-inline">{{ $gettext(`This Day in the Past`) }}</span>
          <span class="d-inline d-md-none">{{ $gettext(`This Day`) }}</span>
        </v-tab>
        <v-tab id="tab-discover-month" ripple @click="changePath('/discover/month')">
          <span class="d-none d-md-inline">{{ $gettext(`This Month in the Past`) }}</span>
          <span class="d-inline d-md-none">{{ $gettext(`This Month`) }}</span>
        </v-tab>
        <v-tab id="tab-discover-random" ripple @click="changePath('/discover/random')">
          {{ $gettext(`Random`) }}
        </v-tab>
      </v-tabs>

      <v-toolbar :density="$vuetify.display.smAndDown ? 'compact' : 'default'" class="page-toolbar" color="secondary">
        <v-spacer></v-spacer>
        <v-btn
          v-if="isRandomTab"
          :title="$gettext('Shuffle')"
          :disabled="shuffling"
          :loading="shuffling"
          icon="mdi-shuffle-variant"
          class="action-shuffle me-1"
          @click.stop="shuffle"
        ></v-btn>
        <v-btn-toggle
          :model-value="view"
          :title="$gettext('Toggle View')"
          :density="$vuetify.display.smAndDown ? 'comfortable' : 'default'"
          base-color="secondary"
          variant="flat"
          rounded="pill"
          mandatory
          border
        >
          <v-btn value="cards" icon="mdi-view-column" class="ps-1 action-view-cards" @click="setView('cards')"></v-btn>
          <v-btn value="mosaic" icon="mdi-view-comfy" class="pe-1 action-view-mosaic" @click="setView('mosaic')"></v-btn>
        </v-btn-toggle>
      </v-toolbar>
    </div>

    <v-tabs-window v-model="active">
      <v-tabs-window-item>
        <p-tab-discover-past mode="day" :view="view"></p-tab-discover-past>
      </v-tabs-window-item>
      <v-tabs-window-item>
        <p-tab-discover-past mode="month" :view="view"></p-tab-discover-past>
      </v-tabs-window-item>
      <v-tabs-window-item>
        <p-tab-discover-random ref="random" :view="view" @loading="onRandomLoading"></p-tab-discover-random>
      </v-tabs-window-item>
    </v-tabs-window>
  </div>
</template>

<script>
import tabPast from "page/discover/past.vue";
import tabRandom from "page/discover/random.vue";
import { getAppStorage } from "common/storage";

const appStorage = getAppStorage();

export default {
  name: "PPageDiscover",
  components: {
    "p-tab-discover-past": tabPast,
    "p-tab-discover-random": tabRandom,
  },
  props: {
    tab: {
      type: Number,
      default: 0,
    },
  },
  data() {
    return {
      active: this.tab,
      view: this.storedView(),
      randomLoading: false,
    };
  },
  computed: {
    isRandomTab() {
      return this.active === 2;
    },
    shuffling() {
      return this.isRandomTab && this.randomLoading;
    },
  },
  watch: {
    tab(value) {
      this.active = value;
    },
    // active remeasures the toolbar after a tab change so year headers stay below it.
    active() {
      this.$nextTick(() => this.syncNavOffset());
    },
  },
  mounted() {
    this.$view.enter(this);
    this.$nextTick(() => this.syncNavOffset());
    window.addEventListener("resize", this.syncNavOffset);
  },
  unmounted() {
    window.removeEventListener("resize", this.syncNavOffset);
    this.$view.leave(this);
  },
  methods: {
    // storedView returns the Cards/Mosaic preference for Discover without touching Browse.
    storedView() {
      const stored = appStorage.getItem("discover.view");

      if (stored === "mosaic" || stored === "cards") {
        return stored;
      }

      return window.innerWidth < 960 ? "mosaic" : "cards";
    },
    // setView persists the Cards/Mosaic toggle for Discover.
    setView(view) {
      this.view = view;
      appStorage.setItem("discover.view", view);
    },
    // syncNavOffset keeps year headers below the sticky tab and toolbar bar.
    syncNavOffset() {
      const nav = this.$el?.querySelector?.(".p-page__navigation");
      const height = nav?.offsetHeight || 0;

      if (height > 0) {
        this.$el.style.setProperty("--p-discover-sticky-top", `${height}px`);
      }
    },
    // onRandomLoading tracks Shuffle so the toolbar can show a spinner.
    onRandomLoading(loading) {
      this.randomLoading = loading === true;
    },
    // shuffle asks the Random tab to draw a new set.
    shuffle() {
      this.$refs.random?.shuffle();
    },
    // changePath updates the Discover tab route without stacking history.
    changePath(path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    },
  },
};
</script>
