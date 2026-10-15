import { ClientFunction, Selector, t } from "testcafe";

// dispatchExportShortcut supplies the code omitted by TestCafe's native key events.
const dispatchExportShortcut = ClientFunction((metaKey) => {
  const target = document.querySelector(".p-lightbox__pswp");
  if (!target) {
    throw new Error("The lightbox must be open before using Export shortcuts");
  }
  target.dispatchEvent(new KeyboardEvent("keydown", {
    key: "E", code: "KeyE", keyCode: 69, which: 69,
    ctrlKey: !metaKey, metaKey, shiftKey: true, bubbles: true, cancelable: true,
  }));
});

// triggerExportShortcut waits for focus restoration before exercising the shared shortcut router.
export async function triggerExportShortcut(metaKey = false) {
  await t.expect(Selector(".p-photo-export-dialog .v-overlay__content").visible).notOk();
  await dispatchExportShortcut(metaKey);
}
