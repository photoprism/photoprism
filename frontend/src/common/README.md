## View Helper Guidelines

**Last Updated:** October 15, 2026

### Export Helpers

`export.js` exposes `canExport(config)` for a client `Config` instance and
`buildExportGroups(files, thumbs, thumbLimit)` for the full photo-detail `Files`
array and client config's `thumbs` list and effective `thumbLimit` in pixels.
It performs no requests and does not modify its inputs.
Consumers must load full photo details; search-result file summaries are insufficient.

- Export requires `settings.features.export === true` and `files.access_all` or
  `files.full_access`. The shared `Config.allow()` evaluator supplies the ACL check.
  Download flags, archive filters, and role names do not affect this capability.
- Each group has an original UID `id`, an `original` source snapshot, an `options`
  array, matched raster `thumbnailSource` (or `null`), `previewMissing`, and a
  `selected` option ID (empty when nothing is selected).
  Sources carry `uid`, `root`, `name`, `originalName`, `hash`, `format`, `mediaType`,
  `width`, `height`, `bytes`, `codec`, and `duration` (nanoseconds). Unknown
  dimensions, byte counts, and durations are `null`; unknown codec is empty.
- Options carry `id`, `kind` (`original`, `preview`, or `thumbnail`), `label`,
  `source`, `format`, `width`, and `height`; thumbnail options also carry the server
  `size` code. Labels are descriptors for the consuming UI to translate.
- Available media in Originals (`Root: /`) create groups in list order. Metadata
  sidecars and records with missing files, errors, deletion markers, or absent
  UIDs, names, or hashes are excluded. Videos and animations can use uniquely
  owned existing still previews; matching occurs before any grouping or filtering.
- JPEG/PNG originals supply their own thumbnails. Other still formats use an
  available JPEG/PNG matched by directory and full name (`photo.heic.jpg`), then
  by a stem owned by exactly one original (`photo.jpg`). Within each match level,
  vectors prefer PNG and other media prefer JPG, falling back to the other format.
  Multiple candidates at the chosen format priority or a
  candidate belonging to another original leave only Original and set
  `previewMissing`. Unavailable named originals still count toward ambiguity.
- Preview uses the existing file's format and dimensions. Thumbnail dimensions
  use that source's already oriented dimensions, nearest-pixel rounding, a
  one-pixel minimum, and no enlargement. Only advertised `fit_*` sizes whose box
  width and height are within `thumbLimit` are used, before fitting or merging
  outputs. The server reports the pre-generated limit when on-demand rendering
  is disabled and the larger of its cached and on-demand limits otherwise.
  Callers without a limit default to 1920 pixels.
- Equal output dimensions merge using the smallest box area, then size code.
  Thumbnail options sort by output area. A JPEG at source dimensions is labeled
  Full Size and omitted when Original JPEG or Preview JPEG already supplies it.
  A PNG source keeps that JPEG conversion.
- Only the first group can preselect a surviving `fit_1920` Full HD thumbnail.
  Full Size, a merged-away Full HD size, or missing Full HD leaves it empty.

`export-media.js` adds video, animated-image, motion-photo, and Apple Live Photo
choices to the same groups; `buildMediaExportGroups(files, thumbs, probes, thumbLimit)`
applies the same serving limit to their still previews. Static WebP, AVIF, TIFF, and JPEG XL originals are
eligible with any uniquely owned existing JPG/PNG preview. Audio is omitted.
Indexed images with multiple `Frames` or a positive `Duration` normalize to
animated sources. Animated WebP and APNG require a separate still preview. Motion JPEG/PNG originals can supply their own thumbnails.

Apple JPEG/HEIC and MOV records pair only when a metadata `appleKey` uniquely
identifies one image and one movie. After successful probes establish that both
lack a key, a unique pair with the same directory and basename requires a verified
movie duration of more than zero and at most 3.1 seconds. Missing named siblings
count against basename uniqueness. `InstanceID` is not consulted. At the first
member's position, the still and MOV appear as two groups with their own file UIDs,
still first and MOV immediately afterward; other groups retain their relative order.
The still retains its Original, preview, and JPEG choices. The MOV offers only its
Original and may use the still's matched raster for its header thumbnail. Unpaired
or ambiguous MOVs retain normal video choices. Preview ownership is resolved
across the entire stack before pairing. Only the first displayed group preselects
a surviving Full HD thumbnail.

Video and animation groups offer MP4 AVC when probe capabilities identify a ready
or preparable output and the source codec is not AVC/H.264. Container extensions
alone do not identify codecs. Motion photos offer only a verified native Video
Part; its real container and codec are retained. Missing capabilities leave the
Original and available still choices intact.
Zero-valued byte counts in video capabilities mean the output size is unknown;
the dialog omits that size until a positive byte count is available.

`export-download.js` freezes selected sources, names, and group order. It orders
Originals, Preview, thumbnails from smallest to largest, then video derivatives.
Final names are sanitized after appending the output extension. Unicode
whitespace is removed to match server cleaning; normal internal ASCII spaces
are preserved. Case-insensitive collisions across originals and derivatives
receive `_2`, `_3`, and subsequent suffixes. Original names use their own last `OriginalName`/`Name` component and
actual extension. Other outputs use the still/group basename with `_preview`,
`_fit_*`, `_avc.mp4`, or `_video` and the native container extension.

`ExportQueue` refreshes sessions and constructs URLs with current tokens. Ready
outputs are handed directly to the browser immediately, with 250 ms between
handoffs. A single preparation runs alongside them; a waiting conversion does
not block ready files. Each fixed output has `waiting`, `preparing`, `ready`,
`handed`, or `error` state. Retry refreshes authorization and schedules only the
failed output. A ready manual handoff is removed from the automatic queue;
explicit manual downloads of already handed outputs do not increment the count.

`component/photo/export/dialog.vue` loads full photo details and, when eligible
probe sources exist, refreshes the session and probes them sequentially before
displaying choices.
The Lightbox opens Export with Ctrl/Command+Shift+E, displayed as `Ctrl-Shift-E`
in its action menu. Ctrl/Command+E opens Edit. Export shortcuts defer to focused
text editors and remain disabled during face-marker mode and export transitions.
The dialog explicitly uses the selected application theme, is 560 px wide on
desktop, and becomes fullscreen on mobile. Its header and actions stay fixed
while choices scroll. Desktop gutters are 24 px and fullscreen gutters are 16 px.
Header and option text share a 44 px preview/control column with a 12 px gap;
primary actions are 44 px high. File headers use 44 px matched-raster thumbnails or media
icons, original filenames, format/media labels, and known metadata. Options show
their label and metadata with known size on the right. Presentation orders Original, Compatible
Video/Video Part, Full-Size Preview, then JPEG choices in their existing size order;
JPEG sizes share one labeled Resized JPEGs section under each original. Compatible
Video choices show a small Needs Conversion note until their output is prepared.
This does not change queue order. Selection counts appear in chips and the Export
button; no selection disables Export. The multiple-download hint appears only
when more than one output is selected. Confirmation replaces the selection list
with compact progress rows for unfinished or failed outputs. Rows show the output
filename, metadata, status, indeterminate preparation, and individual Retry for
failures. Successful handoffs leave the list. The dialog closes automatically
after all selected files are handed to the browser, including successful retries.
Pending work offers Cancel; failed-only results offer Close. The footer reports
errors. The browser's actual disk-save completion is not observable.
Cancel, Escape, a changed photo,
Lightbox closure, and permission loss invalidate client work and late replies.
Started server work may finish publishing its sidecar. Reopening restores initial
selections and reprobes capabilities. Lightbox and PhotoSwipe shortcuts remain suspended
through the closing transition.

#### Video API Integration Status

The client consumes JSON from
`/api/v1/videos/{hash}/{downloadToken}/export?export=probe|prepare|download`;
preparation/download requests also select `kind=avc|video-part`. Probe data uses
`codec`, `duration` (seconds), `appleKey`, and `avc`/`videoPart` capabilities
(`ready`, `preparable`, `format`, `codec`, `width`, `height`, `bytes`). Prepare
returns a verified output with `ready: true` and the same format, codec, dimensions,
and byte fields. Download is an attachment with an optional sanitized `filename`.
JSON bodies do not carry paths or tokens.

The backend handles these modes before playback token validation and stack
fallback. Each request uses download admission and per-file visibility. The path
download token is authoritative; query credentials cannot replace it. Pairing keys
are omitted for sessions without full file or photo details. Responses are not cached
by browsers or shared proxies. Export feature and menu-role gating remain client-side.

Probe reads metadata and verified sidecars without creating files. Prepare shares
same-source work and serializes writers of the same standard AVC destination with
playback. Conversion stages and verifies an output before atomic publication;
a failed replacement preserves the existing sidecar. At most four distinct Export
conversions and four source probes or hash validations run concurrently; excess
work returns 429 with Retry-After. Started conversion can finish after client
cancellation. Download never starts conversion.

AVC outputs use the full source filename, such as `clip.mov.avc` or
`animation.gif.mp4`; ambiguous basename siblings are not used by Export.
Verified sidecars are reused even after conversion settings change or conversion
is disabled. This associates existing output by its standard source path, without
checksum receipts or a guarantee that it was generated from the current source
bytes. Source changes require the normal sidecar maintenance policy. An invalid
nonempty sidecar is unavailable and is not automatically overwritten. No separate
Export cache is read, written, migrated, or cleaned up.

Native motion video is already ready: prepare returns its metadata, and download
serves its bounded segment directly from the source, preserving bytes, codec, and
container with Range support. Preparation and download check the indexed source
hash; conversion retains the observed source version through waiting and encoder
fallback. Frontend unit tests mock this contract; backend tests exercise real
conversion and segment delivery.

Tests live in `frontend/tests/vitest/common/export.test.js`,
`frontend/tests/vitest/common/export-media.test.js`,
`frontend/tests/vitest/common/export-download.test.js`,
`frontend/tests/vitest/component/photo/export/dialog.test.js`, and
`frontend/tests/vitest/component/lightbox.export.test.js`.

### Focus Management

PhotoPrism uses a shared view helper to maintain predictable focus across pages and dialogs:

- [`frontend/src/common/view.js`](https://github.com/photoprism/photoprism/blob/develop/frontend/src/common/view.js)

This helper tracks the currently active component, applies focus when views change, and traps focus inside open dialogs, ensuring that tabbing never leaks into the page behind an overlay. The following guidelines explain how to work with the helper when building UI functionality.

#### Session Storage & Web View Notes

When integrating third-party clients such as native mobile apps that embed the PhotoPrism web UI in a web view and pre-populate browser storage, keep the following behavior in mind:

- Storage keys are namespaced per app site via `storageNamespace` and use the format `pp:<storageNamespace>:<key>`. When no namespace is available, the fallback prefix is `pp:root:`.
- Session restore requires both `session.token` and `session.id`. A token-only write does not create an authenticated session state during startup.
- Optional session metadata keys are `session.provider`, `session.user`, `session.scope`, `session.data`, and `session.error`.
- The login-flow keys `login.next` and `login.logout` live in namespaced `localStorage` outside the `session.*` prefix so they survive the OIDC callback's clear list. `login.next` is the deep-link target captured by the global router guard via `$session.setLoginRedirectUrl(to.href)`; the `/login` route reads it through `$session.hasLoginRedirectUrl()` (deep-link arrival signal) and consumes it with `$session.followLoginRedirectUrl()` after authentication. `login.logout` is a one-shot raised by every logout path so the very next deep-link arrival skips the auto-OIDC bounce, letting users sign in locally instead.
- The storage preference flag is always stored in namespaced `localStorage` under `session`. A value of `"true"` tells the app to use namespaced `sessionStorage` for the active session; any other value falls back to namespaced `localStorage`.
- `Session` must resolve `storageNamespace` from the actual client config shape used at runtime. In production, that value may be present on `config.values.storageNamespace` even when a direct `config.storageNamespace` property is absent.
- The OIDC callback template clears legacy global auth/session payload keys and namespaced session data keys from `localStorage` and `sessionStorage`, but preserves the `session` storage-preference flag so OIDC flows using ephemeral `sessionStorage` remain logged in after redirect.
- During migration, the OIDC callback also honors the legacy unnamespaced `localStorage["session"] === "true"` preference when choosing the preferred write target before rewriting the session into namespaced storage.
- Legacy compatibility keys (`authToken`, `sessionId`, `sessionData`, `user`, and `provider`) are only migrated by the startup bridge when both the legacy token and legacy session id are present. New integrations should not rely on those keys.
- Malformed JSON stored in `session.user` or `session.data` is discarded during startup instead of being reused.
- In the standard login UI, this preference is exposed through the `Stay signed in on this device` toggle. Checked means persistent namespaced `localStorage`; unchecked means ephemeral namespaced `sessionStorage`.
- The login page initializes that toggle from the current session storage mode when available, and otherwise falls back to the stored namespaced `session` preference flag.
- Logout and session reset clear the current app namespace from both `localStorage` and `sessionStorage`, and also remove deprecated raw legacy auth keys from both stores, without touching other namespaced instances on the same origin.
- When `PHOTOPRISM_OIDC_LOGOUT` is enabled, `DELETE /api/v1/session` returns a `providerLogoutUri`; `logout()` passes it to `onLogout()`, which redirects the browser to the provider's end-session endpoint (RP-initiated logout) instead of the local login page, so the provider's SSO session is ended too. `onLogout()` also resolves to that landing URL, so callers that pass `noRedirect` and navigate themselves — the `/logout` route guard and the Portal instance-chooser Sign-Out — follow the same provider URL rather than discarding it. Without the flag, sign-out lands on the local/Portal login page as before.
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

   Only add local `@focusout` handlers if a dialog needs custom behavior. If you do, always call `ev.preventDefault()` when you redirect focus so you do not fight the global handler.

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
- Persistent dialogs that must suppress Vuetify’s rejection animation should still attach a direct `@keydown.esc.exact` handler; `onShortCut(ev)` alone does not override the built-in dialog behavior.
- Return `true` from `onShortCut(ev)` after handling a shortcut to signal `preventDefault()`. Return `false` to fall back to the browser’s native behavior.

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
- Consider removing the custom handler if the global trap already matches the desired behavior.

##### Nested Dialogs (Dialog Inside Dialog)
- Each dialog must have `ref="dialog"` so the helper can distinguish them.
- The helper chooses the currently active component (`this.$view.getCurrent()`) as the trap owner, so opening a second dialog automatically pauses the first one’s trap.
