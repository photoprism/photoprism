# View Helper Guidelines

**Last Updated:** November 11, 2025

## Focus Management

PhotoPrism maintains predictable keyboard focus across pages and dialogs by using a shared view helper:

- [`frontend/src/common/view.js`](https://github.com/photoprism/photoprism/blob/develop/frontend/src/common/view.js)

This helper tracks the currently active component, applies focus when views change, and traps focus inside open dialogs, ensuring that tabbing never leaks into the page behind an overlay. The following guidelines explain how to work with the helper when building UI functionality.

### Tabindex Cheat Sheet

| Value      | When to use it                                          | Effect                                                                                         |
|------------|---------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `0`        | Interactive controls in the natural tab order           | Element participates in sequential keyboard focus                                              |
| `-1`       | Programmatic focus targets (dialog wrappers, sentinels) | Element can receive focus via script but is skipped while tabbing                              |
| *positive* | **Avoid**                                               | Custom tab order becomes hard to maintain; the view helper no longer knows the “first” element |

**Tips**

- Root page containers (`<div class="p-page ...">`) should use `tabindex="-1"` so the view helper can focus them when a route becomes active, then immediately move focus to the first interactive control.
- Leave buttons, inputs, and links at the default `tabindex="0"` (or no attribute) so the browser controls the natural order.

### Dialog Implementation Checklist

Vuetify dialogs are teleported to the overlay container, so consistent refs and lifecycle hooks are essential.

1. **Add refs and focus hooks**

   ```vue
   <v-dialog
     ref="dialog"
     :model-value="visible"
     persistent
     max-width="350"
     @after-enter="afterEnter"
     @after-leave="afterLeave"
   >
     <v-card ref="content" tabindex="-1">
       <!-- dialog body -->
     </v-card>
   </v-dialog>
   ```

   ```js
   export default {
     methods: {
       afterEnter() {
         this.$view.enter(this);
       },
       afterLeave() {
         this.$view.leave(this);
       },
     },
   };
   ```

   - `ref="dialog"` lets the view helper grab the teleported overlay via `ref.contentEl`.
   - The `$view.enter/leave` calls are mandatory so the helper knows when to trap or release focus.

2. **Keep the first focusable control at `tabindex="0"`**

   ```vue
   <v-card-actions class="action-buttons">
     <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
       {{ $gettext(`Cancel`) }}
     </v-btn>
     <v-btn variant="flat" color="highlight" class="action-confirm" @click.stop="confirm">
       {{ $gettext(`Delete`) }}
     </v-btn>
   </v-card-actions>
   ```

   The view helper resolves the first tabbable element (`Cancel` in this case) as the fallback when tabbing inside the dialog.

3. **Avoid per-dialog traps unless necessary**

   Only add local `@focusout` handlers if a dialog needs custom behaviour. If you do, always call `ev.preventDefault()` when you redirect focus so you do not fight the global handler.

### Example: Delete Confirmation Dialog

```vue
<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    persistent
    max-width="350"
    class="p-dialog p-file-delete-dialog"
    @keydown.esc.exact="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" tabindex="-1">
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon icon="mdi-delete-outline" size="54" color="primary"></v-icon>
        <p class="text-subtitle-1">{{ $gettext(`Are you sure you want to permanently delete this file?`) }}</p>
      </v-card-title>
      <v-card-actions class="action-buttons mt-1">
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
          {{ $gettext(`Cancel`) }}
        </v-btn>
        <v-btn color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          {{ $gettext(`Delete`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "PFileDeleteDialog",
  props: {
    visible: Boolean,
  },
  emits: ["close", "confirm"],
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      this.$emit("confirm");
    },
  },
};
</script>
```

This pattern ensures:

- The dialog registers with the view helper as soon as it appears (`afterEnter`).
- Focus defaults to the `Cancel` button (first tabbable control).
- Tabbing continues to cycle between `Cancel` and `Delete` until the dialog closes.

### Troubleshooting Checklist

**Focus escapes the dialog when tabbing**
- Verify the dialog calls `$view.enter(this)` / `$view.leave(this)`.
- Confirm the dialog template has `ref="dialog"`; if you teleport manually, expose `contentEl`.
- Ensure there is at least one control with `tabindex="0"` inside the card. Pure static content cannot trap focus.

**Focus lands on the overlay instead of a button**
- Check for stray `tabindex="-1"` on child elements. Only the outer container should use `-1`.
- Use the browser console with `trace` logging enabled (`this.$config.get("trace")`) to see which elements receive `document.focusin/out`.

**Custom focusout handler keeps fighting the trap**
- Make sure the handler checks `this.$view.isActive(this)` and calls `ev.preventDefault()` when redirecting focus.
- Consider removing the custom handler if the global trap already matches the desired behaviour.

**Nested dialogs (dialog inside dialog)**
- Each dialog must have `ref="dialog"` so the helper can distinguish them.
- The helper chooses the currently active component (`this.$view.getCurrent()`) as the trap owner, so opening a second dialog automatically pauses the first one’s trap.
