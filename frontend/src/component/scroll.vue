<template>
  <transition name="fade-transition">
    <button v-if="showButton" type="button" class="p-scroll" @click.stop="scrollToTop">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M13 20h-2V8l-5.5 5.5-1.42-1.42L12 4.16l7.92 7.92-1.42 1.42L13 8z"></path></svg>
      <translate>Back to top</translate>
    </button>
  </transition>
</template>

<script>
export default {
  name: "PScroll",
  props: {
    loadMore: {
      type: Function,
      default: () => {},
    },
    loadDistance: {
      type: Number,
      default: window.innerHeight * 3,
    },
    loadDisabled: {
      type: Boolean,
      default: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      maxScrollY: 0,
      wait: false,
      showButton: false,
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

        if (this.showButton) {
          this.showButton = false;
        }

        if (!this.loadDisabled && !this.wait && document.documentElement.scrollHeight - window.scrollY < this.loadDistance) {
          this.wait = true;
          this.loadMore();
          setTimeout(() => {
            this.wait = false;
          }, 1000);
        }
      } else if (window.scrollY < window.innerHeight) {
        if (this.showButton) {
          this.showButton = false;
        }
        if (this.maxScrollY !== 0) {
          this.maxScrollY = 0;
        }
      } else if (!this.showButton && !this.loading && !this.wait && this.maxScrollY - window.scrollY > 80) {
        this.showButton = true;
      }
    },
    scrollToTop() {
      window.scrollTo({ top: 0, behavior: "smooth" });
    },
  },
};
</script>
