<template>
    <transition name="fade-transition">
        <v-btn
                color="grey darken-2"
                dark
                fab
                fixed
                right
                top
                @click.stop="scrollToTop"
                v-if="scrolled"
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
                scrolled: false
            };
        },
        methods: {
            onScroll: function () {
                this.scrolled = window.scrollY > 300;
            },
            scrollToTop: function () {
                console.log("scrollY:", window.scrollY);
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
