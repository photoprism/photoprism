<template>
  <div class="p-tab p-tab-discover-todo">
    <div class="pa-2 text-center">
      <p class="text-subtitle-1 pb-6">
        This is a very first draft for a "Discover" area where you can find photos by color, by similarity, by season or
        just randomly. Feedback and contributions welcome.
      </p>
    </div>
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
