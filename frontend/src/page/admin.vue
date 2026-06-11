<template>
  <div class="p-page p-page-admin"></div>
</template>
<script>
// PPageAdmin is a Community Edition stub: the admin area only exists as a
// Plus/Pro/Portal overlay, so in CE it redirects to the user's default route.
export default {
  name: "PPageAdmin",
  created() {
    this.$router.replace({ name: this.$session.getDefaultRoute() });
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
