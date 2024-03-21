<template>
  <div class="auth-header">
    <div id="auth-logo" class="logo text-xs-center">
      <img :src="$config.getIcon()" :alt="config.name" />
    </div>
  </div>
</template>
<script>
export default {
  name: "PAuthHeader",
  props: {},
  data() {
    return {
      sponsor: this.$config.isSponsor(),
      config: this.$config.values,
      rtl: this.$rtl,
    };
  },
  methods: {},
};
</script>
