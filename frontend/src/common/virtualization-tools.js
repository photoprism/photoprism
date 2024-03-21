export const virtualizationTools = {
  updateVisibleElementIndices: (visibleElementIndices, entries, elementIndexFromEntry) => {
    entries.forEach((entry) => {
      const inView = entry.isIntersecting && entry.intersectionRatio >= 0;
      const elementIndex = elementIndexFromEntry(entry);
      if (elementIndex === undefined || elementIndex < 0) {
        return;
      }

      if (inView) {
        visibleElementIndices.add(elementIndex);
      } else {
        /**
         * If the target has no parent-node, it's no longer in the dom-tree.
         * If the element is no longer inView because it was removed from the
         * dom-tree, then this says nothing about the visible indices.
         * If you remove a picture from the grid, the space where the picture
         * was is still visible.
         *
         * We therefore must ignore entries that became invisible that no longer
         * exists
         */
        const entryIsStillMounted = entry.target.parentNode !== null;
        if (entryIsStillMounted) {
          visibleElementIndices.delete(elementIndex);
        }
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
};
