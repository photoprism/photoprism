<template>
  <v-icon v-if="editing" icon="mdi-check" size="small" class="meta-inline-confirm" @mousedown.prevent @click.stop="$emit('confirm')"></v-icon>
  <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="$emit('start')"></v-icon>
</template>

<script>
export default {
  name: "PSidebarInlineToolbar",
  props: {
    editing: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["confirm", "start"],
};
</script>
