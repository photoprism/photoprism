<script>
// v-longpress: Bind to a function to be executed after long-pressing.
// Set arg = true or false to enable or disable long-pressing.
export default {
  name: "PLongPress",
  bind: function (el, binding, vNode) {
    if (typeof (binding.value) !== 'function') {
      console.warn(`v-longpress: Must be bound to a function. Expression in error: '${binding.expression}'.`);
    }

    el.longPressActive = (binding.arg === true);
    let longPressTimer = null

    el.startLongPressTimer = (event) => {
      if (!event.isPrimary) return;
      if (!el.longPressActive) return;
      if (longPressTimer !== null) return; // Already counting down
      longPressTimer = setTimeout(() => {
        el.style.boxShadow = "";
        binding.value(event);
      }, 1500); // Call the bound function after 1.5 seconds
      setTimeout(() => {
        if (longPressTimer !== null) {
          el.style.boxShadow = "0px 0px 70px 40px " + window.getComputedStyle(el).backgroundColor;
        }
      }, 300);  // Visual feedback for long-press after 0.3 seconds
    }

    el.cancelLongPressTimer = (event) => {
      el.style.boxShadow = "";
      if (longPressTimer !== null) {
        clearTimeout(longPressTimer);
        longPressTimer = null;
      }
    }

    // Start the countdown on pointer down
    el.addEventListener("pointerdown", el.startLongPressTimer);
    // Cancel the countdown if the pointer does not stay down over the element for 1.5 seconds
    el.addEventListener("click", el.cancelLongPressTimer);
    el.addEventListener("pointerup", el.cancelLongPressTimer);
    el.addEventListener("pointercancel", el.cancelLongPressTimer);
    el.addEventListener("pointerleave", el.cancelLongPressTimer);
    el.addEventListener("pointerout", el.cancelLongPressTimer);
    el.addEventListener("mouseup", el.cancelLongPressTimer);
  },
  update: function (el, binding, vNode) {
    el.longPressActive = (binding.arg === true);
  },
  unbind: function (el, binding, vNode) {
    el.removeEventListener("pointerdown", el.startLongPressTimer);
    el.removeEventListener("click", el.cancelLongPressTimer);
    el.removeEventListener("pointerup", el.cancelLongPressTimer);
    el.removeEventListener("pointercancel", el.cancelLongPressTimer);
    el.removeEventListener("pointerleave", el.cancelLongPressTimer);
    el.removeEventListener("pointerout", el.cancelLongPressTimer);
    el.removeEventListener("mouseup", el.cancelLongPressTimer);
  },
}
</script>
