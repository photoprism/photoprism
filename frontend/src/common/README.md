## View Helper Guidelines

**Last Updated:** March 8, 2026

### Focus Management

PhotoPrism uses a shared view helper to maintain predictable focus across pages and dialogs:

- [`frontend/src/common/view.js`](https://github.com/photoprism/photoprism/blob/develop/frontend/src/common/view.js)

This helper tracks the currently active component, applies focus when views change, and traps focus inside open dialogs, ensuring that tabbing never leaks into the page behind an overlay. The following guidelines explain how to work with the helper when building UI functionality.

#### Session Storage & Web View Notes

When integrating third-party clients such as native mobile apps that embed the PhotoPrism web UI in a web view and pre-populate browser storage, keep the following behavior in mind:

- Storage keys are namespaced per app site via `storageNamespace` and use the format `pp:<storageNamespace>:<key>`. When no namespace is available, the fallback prefix is `pp:root:`.
- Session restore requires both `session.token` and `session.id`. A token-only write does not create an authenticated session state during startup.
- Optional session metadata keys are `session.provider`, `session.user`, `session.scope`, `session.data`, and `session.error`.
- The storage preference flag is always stored in namespaced `localStorage` under `session`. A value of `"true"` tells the app to use namespaced `sessionStorage` for the active session; any other value falls back to namespaced `localStorage`.
- `Session` must resolve `storageNamespace` from the actual client config shape used at runtime. In production, that value may be present on `config.values.storageNamespace` even when a direct `config.storageNamespace` property is absent.
- The OIDC callback template clears legacy global auth/session payload keys and namespaced session data keys from `localStorage` and `sessionStorage`, but preserves the `session` storage-preference flag so OIDC flows using ephemeral `sessionStorage` remain logged in after redirect.
- During migration, the OIDC callback also honors the legacy unnamespaced `localStorage["session"] === "true"` preference when choosing the preferred write target before rewriting the session into namespaced storage.
- Legacy compatibility keys (`authToken`, `sessionId`, `sessionData`, `user`, and `provider`) are only migrated by the startup bridge when both the legacy token and legacy session id are present. New integrations should not rely on those keys.
- Malformed JSON stored in `session.user` or `session.data` is discarded during startup instead of being reused.
- In the standard login UI, this preference is exposed through the `Stay signed in on this device` toggle. Checked means persistent namespaced `localStorage`; unchecked means ephemeral namespaced `sessionStorage`.
- The login page initializes that toggle from the current session storage mode when available, and otherwise falls back to the stored namespaced `session` preference flag.
- Logout and session reset clear the current app namespace from both `localStorage` and `sessionStorage`, and also remove deprecated raw legacy auth keys from both stores, without touching other namespaced instances on the same origin.
- Regression tests for session restore should exercise both preferred `sessionStorage` and a real `Config`-shaped object so namespace resolution bugs cannot silently fall back to `pp:root:`.

Preferred integration contract for new native or web view clients:

1. Read `storageNamespace` from the client config so the keys match the current PhotoPrism site.
2. Decide whether the session should be ephemeral or persistent:
   - Set `localStorage["pp:<storageNamespace>:session"] = "true"` for ephemeral auth stored in namespaced `sessionStorage`.
   - Set `localStorage["pp:<storageNamespace>:session"] = "false"` or leave it unset for persistent auth stored in namespaced `localStorage`.
3. Write both `session.token` and `session.id` into the selected namespaced store before loading the main app.
4. Optionally write `session.provider`, `session.user`, `session.scope`, and `session.data` if the embedding flow already has those values.

Example for a persistent session:

```js
const ns = "app.localssl.dev";
const prefix = `pp:${ns}:`;

window.localStorage.setItem(prefix + "session", "false");
window.localStorage.setItem(prefix + "session.token", "<access-token>");
window.localStorage.setItem(prefix + "session.id", "<session-id>");
```

Example for an ephemeral session:

```js
const ns = "app.localssl.dev";
const prefix = `pp:${ns}:`;

window.localStorage.setItem(prefix + "session", "true");
window.sessionStorage.setItem(prefix + "session.token", "<access-token>");
window.sessionStorage.setItem(prefix + "session.id", "<session-id>");
```

#### Tabindex Cheat Sheet

| Value      | When to use it                                          | Effect                                                                                         |
|:-----------|:--------------------------------------------------------|:-----------------------------------------------------------------------------------------------|
| `0`        | Interactive controls in the natural tab order           | Element participates in sequential keyboard focus                                              |
| `-1`       | Programmatic focus targets (dialog wrappers, sentinels) | Element can receive focus via script but is skipped while tabbing                              |
| *positive* | **Avoid**                                               | Custom tab order becomes hard to maintain; the view helper no longer knows the “first” element |

##### Tips

- Root page containers (`<div class="p-page ...">`) should use `tabindex="-1"` so the view helper can focus them when a route becomes active, then immediately move focus to the first interactive control.
- Leave buttons, inputs, and links at the default `tabindex="0"` (or no attribute) so the browser controls the natural order.

#### Dialog Implementation Checklist

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
     <v-btn variant="flat" color="button"
         class="action-cancel" @click.stop="close">
       {{ $gettext(`Cancel`) }}
     </v-btn>
     <v-btn variant="flat" color="highlight"
         class="action-confirm" @click.stop="confirm">
       {{ $gettext(`Delete`) }}
     </v-btn>
   </v-card-actions>
   ```

   The view helper resolves the first tabbable element (`Cancel` in this case) as the fallback when tabbing inside the dialog.

3. **Avoid per-dialog traps unless necessary**

   Only add local `@focusout` handlers if a dialog needs custom behaviour. If you do, always call `ev.preventDefault()` when you redirect focus so you do not fight the global handler.

#### Keyboard Event Handling

Dialogs and page shells often react to keyboard shortcuts (Escape to close, Enter to confirm, etc.). To keep those handlers compatible with text inputs and other interactive children:

- Attach listeners to the focusable container that the view helper manages – the page wrapper with `tabindex="-1"` or the dialog root (`<v-dialog ref="dialog">`).
- Prefer `@keyup` (for example, `@keyup.enter.exact="confirm"`) so elements inside the container receive `keydown` events first and can call `event.stopPropagation()` when they need to keep the key (such as pressing Enter inside a form field).
- **Persistent dialogs (`persistent` attribute)** must handle the Escape key with `@keydown.esc.exact="close"`. Vuetify’s built-in Escape handler plays a “rejection” shake animation when the dialog refuses to close; attaching a direct keydown listener overrides the built-in handler and suppresses the animation while still allowing inner inputs to cancel the event.
- Combine modifiers like `.exact` and `.stop` intentionally. Use `.stop` only when the handler fully resolves the action; otherwise allow events to bubble to ancestor traps.
- If a component must react on `keydown`, scope the listener to the specific control instead of the container, and document why the early trigger is required.
- When emitting from reusable components, forward the native event (`close(event)`) so parents can inspect `event.defaultPrevented` or `event.key` before acting.

Note: To override Vuetify’s built-in `<v-dialog>` Escape handler (and stop the “rejection” animation on persistent dialogs), attach a direct `@keydown.esc.exact="close"` listener; the global `onShortCut(ev)` hook is not sufficient on its own.

Example dialog wiring:

```vue
<v-dialog
  ref="dialog"
  persistent
  @keydown.esc.exact="close"
  @keyup.enter.exact="confirm"
>
  <v-card ref="content" tabindex="-1">
    <!-- dialog body -->
  </v-card>
</v-dialog>
```

Example page container:

```vue
<template>
  <div class="p-page p-settings" tabindex="-1" @keyup.esc.exact="maybeClose">
    <!-- page content -->
  </div>
</template>
```

Both snippets allow focused inputs to veto shortcuts by calling `event.stopPropagation()` or `event.preventDefault()` before the key reaches the container listener, keeping focus management predictable across the app.

##### Global Shortcut Forwarding

`common/view.js` registers a single `keydown` listener that forwards shortcut keys to the active component:

```js
// onKeyDown forwards global shortcuts (Escape, Ctrl/⌘ combos)
// to the active component when supported.
onKeyDown(ev) {
  if (!this.current || !ev || !(ev instanceof KeyboardEvent) || !ev.code) {
    return;
  } else if (!ev.ctrlKey && !ev.metaKey && ev.code !== "Escape") {
    return;
  } else if (typeof this.current?.onShortCut !== "function") {
    return;
  }

  if (this.current.onShortCut(ev)) {
    ev.preventDefault();
  }
}
```

- Implement `onShortCut(ev)` on pages or dialogs when you need to react to Ctrl / ⌘ combinations or global Escape handling. The helper only forwards events where `ev.ctrlKey` or `ev.metaKey` is `true`, or the Escape key is pressed, so it cannot be repurposed for arbitrary keys.
- Persistent dialogs that must suppress Vuetify’s rejection animation should still attach a direct `@keydown.esc.exact` handler; `onShortCut(ev)` alone does not override the built-in dialog behaviour.
- Return `true` from `onShortCut(ev)` after handling a shortcut to signal `preventDefault()`. Return `false` to fall back to the browser’s native behaviour.

#### Example: Delete Confirmation Dialog

```vue
<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    persistent
    max-width="350"
    class="p-dialog p-file-delete-dialog"
    @keydown.esc.exact="close"
    @keyup.enter.exact="confirm"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" tabindex="-1">
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon icon="mdi-delete-outline" size="54" color="primary"></v-icon>
        <p class="text-subtitle-1">{{ $gettext(`Are you sure?`) }}</p>
      </v-card-title>
      <v-card-actions class="action-buttons mt-1">
        <v-btn variant="flat" color="button"
               class="action-cancel" @click.stop="close">
          {{ $gettext(`Cancel`) }}
        </v-btn>
        <v-btn color="highlight" variant="flat"
               class="action-confirm" @click.stop="confirm">
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

#### Troubleshooting Checklist

##### Focus Escapes the Dialog When Tabbing
- Verify the dialog calls `$view.enter(this)` / `$view.leave(this)`.
- Confirm the dialog template has `ref="dialog"`; if you teleport manually, expose `contentEl`.
- Ensure there is at least one control with `tabindex="0"` inside the card. Pure static content cannot trap focus.

##### Focus Lands on the Overlay Instead of a Button
- Check for stray `tabindex="-1"` on child elements. Only the outer container should use `-1`.
- Use the browser console with `trace` logging enabled (`this.$config.get("trace")`) to see which elements receive `document.focusin/out`.

##### Custom Focusout Handler Keeps Fighting the Trap
- Make sure the handler checks `this.$view.isActive(this)` and calls `ev.preventDefault()` when redirecting focus.
- Consider removing the custom handler if the global trap already matches the desired behaviour.

##### Nested Dialogs (Dialog Inside Dialog)
- Each dialog must have `ref="dialog"` so the helper can distinguish them.
- The helper chooses the currently active component (`this.$view.getCurrent()`) as the trap owner, so opening a second dialog automatically pauses the first one’s trap.
