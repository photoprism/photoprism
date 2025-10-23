// Utility functions for handling album selection logic

export function processAlbumSelection(selectedAlbums, availableAlbums) {
  if (!Array.isArray(selectedAlbums)) {
    return { processed: [], changed: false };
  }

  let changed = false;
  const processed = [];
  const seenUids = new Set();

  selectedAlbums.forEach((item) => {
    // If it's a string, try to match it with existing albums
    if (typeof item === "string" && item.trim().length > 0) {
      const matchedAlbum = availableAlbums.find(
        (album) => album.Title && album.Title.toLowerCase() === item.trim().toLowerCase()
      );

      if (matchedAlbum && !seenUids.has(matchedAlbum.UID)) {
        // Replace string with actual album object
        processed.push(matchedAlbum);
        seenUids.add(matchedAlbum.UID);
        changed = true;
      } else if (!matchedAlbum) {
        // Keep as string for new album creation
        processed.push(item.trim());
      }
    } else if (typeof item === "object" && item?.UID && !seenUids.has(item.UID)) {
      // Keep existing album objects, but prevent duplicates
      processed.push(item);
      seenUids.add(item.UID);
    } else if (typeof item === "object" && item?.UID && seenUids.has(item.UID)) {
      // Skip duplicate album objects
      changed = true;
    }
  });

  return {
    processed,
    changed: changed || processed.length !== selectedAlbums.length
  };
}

// Creates a selectedAlbums watcher for Vue components
export function createAlbumSelectionWatcher(albumsProperty) {
  return {
    handler(newVal) {
      const availableAlbums = this[albumsProperty] || [];
      const { processed, changed } = processAlbumSelection(newVal, availableAlbums);

      if (changed) {
        this.$nextTick(() => {
          this.selectedAlbums = processed;
        }).catch((error) => {
          console.error('Error updating selectedAlbums:', error);
        });
      }
    }
  };
}
