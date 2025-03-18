<template>
  <div class="p-tab p-tab-discover-colors">
    <div class="pa-2 text-center">
      <p class="text-subtitle-1 pb-6">
        This is a very first draft for a "Discover" area where you can find photos by color, by similarity, by season or
        just randomly. Feedback and contributions welcome.
      </p>

      <v-row class="p-colors">
        <v-col v-for="color in colors" :key="color.name" class="p-color pa-2 d-flex grow" cols="3">
          <v-hover v-slot="{ hover }">
            <v-card
              :to="{ name: 'browse', query: { color: color.name } }"
              :theme="useDark(color)"
              :color="color.example"
              :text="!hover"
              class="clickable py-1"
            >
              <v-card-text class="px-0 py-12 text-body-2">{{ color.label }}</v-card-text>
            </v-card>
          </v-hover>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script>
export default {
  name: "PTabDiscoverColors",
  data() {
    return {
      readonly: this.$config.get("readonly"),
      colors: this.$config.values.colors,
      labels: {},
    };
  },
  created() {},
  methods: {
    useDark(color) {
      switch (color.name) {
        case "white":
          return false;
        default:
          return true;
      }
    },
  },
};
</script>
