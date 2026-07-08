<template>
  <div class="p-page p-page-instances"></div>
</template>
<script>
// PPageInstances is a Community Edition stub: the instance selector only exists
// as a Portal overlay, so in CE/Plus/Pro it redirects to the user's default route.
export default {
  name: "PPageInstances",
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
