<template>
  <transition name="fade-transition">
    <button v-if="showButton" type="button" :style="`top: ${offsetTop}px`" class="p-scroll" @click.stop="scrollToTop">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
        <path d="M13 20h-2V8l-5.5 5.5-1.42-1.42L12 4.16l7.92 7.92-1.42 1.42L13 8z"></path>
      </svg>
      {{ $gettext(`Back to top`) }}
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
    hidePanel: {
      type: Function,
      default: () => {},
    },
    hidePanelDistance: {
      type: Number,
      default: 250,
    },
    resetDistance: {
      type: Number,
      default: 250,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      wait: false,
      waitTime: 100, // ms
      panelHidden: false,
      showButton: false,
      showButtonDistance: 100,
      maxScrollY: 0,
      marginTop: 24,
      offsetTop: 72,
    };
  },
  created() {
    window.addEventListener("scroll", this.onScroll, { passive: true });
  },
  beforeUnmount() {
    window.removeEventListener("scroll", this.onScroll, false);
  },
  methods: {
    onScroll() {
      const scrollY = window.scrollY;

      if (scrollY > this.maxScrollY) {
        if (this.panelHidden && scrollY - this.maxScrollY < this.showButtonDistance) {
          return;
        }

        // Remember the maximum vertical scroll position.
        this.maxScrollY = scrollY;

        // Hide "Back to top" button if currently shown.
        this.resetButton();

        // Hide the expansion panel when scrolling more than specified in this.hidePanelDistance.
        if (scrollY > this.hidePanelDistance) {
          this.onHidePanel();
        }

        // Trigger this.loadMore() callback if within load distance (infinite scrolling).
        if (document.documentElement.scrollHeight - scrollY < this.loadDistance) {
          this.onLoadMore();
        }
      } else if (scrollY < this.resetDistance) {
        // Hide "Back to top" button and reset maximum scroll position.
        this.reset();
      } else if (this.maxScrollY - scrollY > this.showButtonDistance) {
        // Show "Back to top" button if it is not already shown.
        this.onShowButton();
      }
    },
    reset() {
      this.resetButton();
      this.resetHidePanel();
      this.resetScrollY();
    },
    getOffsetTop() {
      const el = document.querySelector(".p-page > .p-page__navigation");

      if (el && el instanceof HTMLElement && el.offsetHeight) {
        return el.offsetHeight;
      }

      return 48;
    },
    onShowButton() {
      // Show "Back to top" button if it is not already visible.
      if (!this.showButton && !this.loading && !this.wait) {
        this.offsetTop = this.getOffsetTop() + this.marginTop;
        this.showButton = true;
      }
    },
    resetButton() {
      // Hide the "Back to top" button if it is visible.
      if (this.showButton) {
        this.showButton = false;
      }
    },
    onHidePanel() {
      // Hide expansion panel when scrolling down and it's not already hidden.
      if (!this.panelHidden) {
        this.hidePanel();
        this.panelHidden = true;
      }
    },
    resetHidePanel() {
      if (this.panelHidden) {
        this.panelHidden = false;
      }
    },
    resetScrollY() {
      // Reset maximum vertical scroll position.
      if (this.maxScrollY > 0) {
        this.maxScrollY = -1;
      }
    },
    onLoadMore() {
      // Call this.loadMore() for infinite scrolling,
      // unless this.loadDisabled or this.wait is set.
      if (!this.loadDisabled && !this.wait) {
        this.onWait();
        this.$nextTick(() => {
          this.loadMore();
        });
      }
    },
    onWait() {
      // Helps ensure that callback functions like this.loadMore()
      // are only called once within this.waitTime.
      if (!this.wait) {
        this.wait = true;
        setTimeout(() => {
          this.wait = false;
        }, this.waitTime);
      }
    },
    scrollToTop() {
      // Scroll smoothly to the top of the browser window.
      window.scrollTo({ top: 0, behavior: "smooth" });
    },
  },
};
</script>
