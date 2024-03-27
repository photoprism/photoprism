<template>
  <div class="p-tab p-tab-discover-colors">
    <v-container grid-list-xs text-xs-center fluid>
      <p class="subheading pb-3"> This is a very first draft for a "Discover" area where you can find photos by color, by similarity, by season or just randomly. Feedback and contributions welcome. </p>

      <v-layout row wrap class="p-colors">
        <v-flex v-for="(color, index) in colors" :key="index" class="p-color pa-2" xs3 d-flex grow>
          <v-hover>
            <v-card slot-scope="{ hover }" :to="{ name: 'browse', query: { color: color.name } }" :dark="useDark(color)" :color="color.example" :flat="!hover" class="clickable py-1">
              <v-card-text class="px-0 py-5 body-2">{{ color.label }}</v-card-text>
            </v-card>
          </v-hover>
        </v-flex>
      </v-layout>
    </v-container>
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
