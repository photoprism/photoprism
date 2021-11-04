<template>
  <v-slider
      v-model="zoomFactor"
      :min="1"
      :max="5"
      :step="1"
      :label="$gettext('Zoom')"
      :hide-details="true"
      class="zoom-factor"
      @change="zoomFactorChange"
  />
</template>

<script>

import Event from "pubsub-js";

export default {
  name: "ZoomFactor",
  data() {
    return {
      zoomFactor: null
    };
  },
  mounted() {
    let current = this.getZoomFactor();
    console.log("Current zoom factor", current);
    this.$set(this, 'zoomFactor', current);

    Event.subscribe('zoom.factor.refresh', () => {
      this.zoomFactorChange();
    });
  },
  methods: {
    getZoomFactor() {
      return this.zoomFactor ?? localStorage.getItem('photoprism.zoom.factor') ?? 1;
    },
    zoomFactorChange() {
      Event.publish('zoom.factor.change', this.zoomFactor);
      localStorage.setItem('photoprism.zoom.factor', this.zoomFactor);
    }
  }
};

</script>