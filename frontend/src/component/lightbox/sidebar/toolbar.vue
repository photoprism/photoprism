<template>
  <div :class="['p-lightbox-sidebar-toolbar', 'd-flex', { 'p-lightbox-sidebar-toolbar--chip': chipMode }]">
    <v-btn
      v-if="editing && canUndo"
      :disabled="undoDisabled"
      icon="mdi-undo-variant"
      density="compact"
      variant="plain"
      size="x-small"
      :class="['meta-inline-undo', { 'meta-chip-undo': chipMode }]"
      :title="$gettext('Undo')"
      @mousedown.prevent
      @click.stop="$emit('undo')"
    ></v-btn>
    <v-btn
      v-if="editing"
      icon="mdi-content-save"
      density="compact"
      variant="plain"
      size="x-small"
      :class="['meta-inline-confirm', { 'meta-chip-confirm': chipMode }]"
      :title="$gettext('Save')"
      @mousedown.prevent
      @click.stop="$emit('confirm')"
    ></v-btn>
    <v-btn
      v-else
      icon="mdi-pencil-outline"
      density="compact"
      variant="plain"
      size="x-small"
      :title="$gettext('Edit')"
      class="meta-inline-pencil"
      @click.stop="$emit('start')"
    ></v-btn>
  </div>
</template>

<script>
export default {
  name: "PLightboxSidebarToolbar",
  props: {
    editing: {
      type: Boolean,
      default: false,
    },
    // Shows an Undo icon when true and `editing` is also true.
    canUndo: {
      type: Boolean,
      default: false,
    },
    // Renders the Undo button in a disabled (non-clickable) state — used
    // by inline-text editors that always show Undo while editing but
    // want it inactive until the value differs from the editOriginal.
    // Chip toolbars leave this at false (their parent v-if already
    // gates mounting on whether Undo would do anything).
    undoDisabled: {
      type: Boolean,
      default: false,
    },
    // Tags the buttons with `meta-chip-*` so the `hide-edit-save` /
    // `hide-edit-undo` rules in css/lightbox.css skip them — chip
    // sections have no keyboard alternative for commit / undo.
    chipMode: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["confirm", "start", "undo"],
};
</script>
