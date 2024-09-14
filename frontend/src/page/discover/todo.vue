<template>
  <div class="p-tab p-tab-discover-todo">
    <v-container grid-list-xs text-xs-center fluid>
      <p class="subheading pb-3"> This is a very first draft for a "Discover" area where you can find photos by color, by similarity, by season or just randomly. Feedback and contributions welcome. </p>
    </v-container>
  </div>
</template>

<script>
export default {
  name: "PTabDiscoverTodo",
  data() {
    return {
      readonly: this.$config.get("readonly"),
      config: this.$config.values,
      labels: {},
    };
  },
  created() {},
  methods: {},
};
</script>
