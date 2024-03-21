<template>
  <transition name="fade-transition">
    <v-btn v-if="show" color="transparent" dark fab fixed class="p-scroll-top" @click.stop="scrollToTop">
      <v-icon>arrow_upward</v-icon>
    </v-btn>
  </transition>
</template>

<script>
export default {
  name: "PScrollTop",
  data() {
    return {
      show: false,
      maxY: 0,
    };
  },
  created() {
    window.addEventListener("scroll", this.onScroll);
  },
  destroyed() {
    window.removeEventListener("scroll", this.onScroll);
  },
  methods: {
    onScroll: function () {
      if (window.scrollY > this.maxY) {
        this.maxY = window.scrollY;
        this.show = false;
      } else if (window.scrollY < 300) {
        this.show = false;
        this.maxY = 0;
      } else if (this.maxY - window.scrollY > 75) {
        this.show = true;
      }
    },
    scrollToTop: function () {
      return this.$vuetify.goTo(0);
    },
  },
};
</script>
