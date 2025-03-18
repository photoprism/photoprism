<template>
  <p-page-about></p-page-about>
</template>

<script>
import PPageAbout from "./about/about.vue";

export default {
  name: "PPageAdmin",
  components: { PPageAbout },
  data() {
    return {
      rtl: this.$isRtl,
    };
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {},
};
</script>
