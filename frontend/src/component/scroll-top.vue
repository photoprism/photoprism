<template>
  <transition name="fade-transition">
    <v-btn
            v-if="show"
            color="accent darken-2"
            dark
            fab
            fixed
            @click.stop="scrollToTop"
            class="p-scroll-top"
    >
      <v-icon>arrow_upward</v-icon>
    </v-btn>
  </transition>
</template>

<script>
    export default {
        name: 'p-scroll-top',
        data() {
            return {
                show: false,
                maxY: 0,
            };
        },
        methods: {
            onScroll: function () {
                if(window.scrollY > this.maxY) {
                    this.maxY = window.scrollY;
                    this.show = false;
                } else if (window.scrollY < 300) {
                    this.show = false;
                    this.maxY = 0;
                } else if ((this.maxY - window.scrollY) > 75) {
                    this.show = true;
                }
            },
            scrollToTop: function () {
                return this.$vuetify.goTo(0);
            },
        },
        created() {
            window.addEventListener('scroll', this.onScroll);
        },
        destroyed() {
            window.removeEventListener('scroll', this.onScroll);
        }
    };
</script>
