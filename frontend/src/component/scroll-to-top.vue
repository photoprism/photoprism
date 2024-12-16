<template>
  <transition name="fade-transition">
    <button v-if="showButton" type="button" class="p-scroll-to-top" @click.stop="scrollToTop">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M13 20h-2V8l-5.5 5.5-1.42-1.42L12 4.16l7.92 7.92-1.42 1.42L13 8z"></path></svg>
      <translate>Back to top</translate>
    </button>
  </transition>
</template>

<script>
export default {
  name: "PScrollToTop",
  data() {
    return {
      showButton: false,
      maxScrollY: 0,
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
      if (window.scrollY > this.maxScrollY) {
        this.maxScrollY = window.scrollY;
        this.showButton = false;
      } else if (window.scrollY < 300) {
        this.showButton = false;
        this.maxScrollY = 0;
      } else if (this.maxScrollY - window.scrollY > 75) {
        this.showButton = true;
      }
    },
    scrollToTop() {
      window.scrollTo({ top: 0, behavior: "smooth" });
    },
  },
};
</script>
