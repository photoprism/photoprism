export const virtualizationTools = {
  updateVisibleElementIndices: (visibleElementIndices, entries, elementIndexFromEntry) => {
    entries.forEach((entry) => {
      const inView = entry.isIntersecting && entry.intersectionRatio >= 0;
      const elementIndex = elementIndexFromEntry(entry);
      if (elementIndex === undefined || elementIndex < 0) {
        return;
      }

      if (inView) {
        visibleElementIndices.add(elementIndex)
      } else {
        visibleElementIndices.delete(elementIndex)
      }
    });

    /**
     * There are many things that can influence what elements are currently
     * visible on the screen, like scrolling, resizing, menu-opening etc.
     *
     * We therefore cannot make assumptions about our new first- and last
     * visible index, even if it is tempting to initialize these values
     * with this.firstVisibleElementIndex and this.lastVisibleElementIndex.
     *
     * Doing so would break the virtualization though. this.firstVisibleElementIndex
     * would for example always stay at 0
     */
    let firstVisibleElementIndex, lastVisibileElementIndex;
    for (const visibleElementIndex of visibleElementIndices.values()) {
      if (firstVisibleElementIndex === undefined || visibleElementIndex < firstVisibleElementIndex) {
        firstVisibleElementIndex = visibleElementIndex;
      }
      if (lastVisibileElementIndex === undefined || visibleElementIndex > lastVisibileElementIndex) {
        lastVisibileElementIndex = visibleElementIndex;
      }
    }

    return [firstVisibleElementIndex, lastVisibileElementIndex];
  },
}