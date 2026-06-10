<template>
  <div class="p-page p-page-cluster"></div>
</template>
<script>
// PPageCluster is a Community Edition stub: the cluster UI only exists as a
// Portal overlay, so in CE/Plus/Pro it redirects to the user's default route.
export default {
  name: "PPageCluster",
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
