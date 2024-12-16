<template>
  <transition name="fade-transition">
    <button v-if="show" type="button" class="p-scroll-top" @click.stop="scrollToTop">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M13 20h-2V8l-5.5 5.5-1.42-1.42L12 4.16l7.92 7.92-1.42 1.42L13 8z"></path></svg>
      <translate>Back to top</translate>
    </button>
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
    window.addEventListener("scroll", this.onScroll, { passive: true });
  },
  beforeUnmount() {
    window.removeEventListener("scroll", this.onScroll, { passive: true });
  },
  methods: {
    onScroll() {
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
    scrollToTop() {
      window.scrollTo({ top: 0, behavior: "smooth" });
    },
  },
};
</script>
