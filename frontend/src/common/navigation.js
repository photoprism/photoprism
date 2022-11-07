window.backwardsNavigationDetected = false;

window.addEventListener("popstate", () => {
  window.backwardsNavigationDetected = true;
  // give components time to react to backwardsNavigationDetected in `created` or '$route'-watcher
  setTimeout(() => {
    window.backwardsNavigationDetected = false;
  });
});
