<template>
  <v-icon
    :icon="icon"
    :aria-label="label"
    @mousedown.prevent
    @click="onClick"
    @keydown.enter.space.prevent.stop="onClick"
    @keyup.enter.space.stop
  ></v-icon>
</template>

<script>
// PInputAction renders a clickable field icon, e.g. in a #prepend-inner or #append-inner slot, with its own
// accessible name; Vuetify derives the name of an icon set through *-icon props from the field label.
export default {
  name: "PInputAction",
  props: {
    icon: {
      type: String,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
  },
  emits: ["click"],
  methods: {
    // onClick forwards a pointer click or an Enter/Space keypress as a click event.
    onClick(ev) {
      this.$emit("click", ev);
    },
  },
};
</script>
