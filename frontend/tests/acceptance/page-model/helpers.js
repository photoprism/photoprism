import { ClientFunction } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";

export const showLogs = process.env.SHOW_LOGS === "true";

// getTopElement will return details on what is on top of a selector.
// Useful when the standard output from testcafe warning is insufficient to identify the obstruction.
export const getTopElement = ClientFunction((selectorFn) => {
  const el = selectorFn();
  const rect = el.getBoundingClientRect();
  const centerX = rect.left + rect.width / 2;
  const centerY = rect.top + rect.height / 2;

  // Returns the topmost element at the exact center of the target
  const topEl = document.elementFromPoint(centerX, centerY);

  return {
    tagName: topEl.tagName.toLowerCase(),
    className: topEl.className,
    id: topEl.id,
    innerText: topEl.innerText,
    innerHTML: topEl.innerHTML,
    outerHTML: topEl.outerHTML
  };
});

// clickIfVisible speeds up testcafe's click by shortening the wait time when an item is visible.
// Should only be used for cases where an overlay obscures a selection, but it will not impact the testing.
export async function clickIfVisible(t, sel) {
  if (await sel.visible) {
    await t.click(sel.with({ timeout: 150 }));
  } else {
    await t.click(sel);
  }
}

export function logMessage(message) {
    if (showLogs) {
        const now = new Date();
        console.log(now.toISOString() + " " + message);
    }
}

export function logTime(key) {
  showLogs && console.time(key);
}

export function logTimeEnd(key) {
  showLogs && console.timeEnd(key);
}

// helperRequest performs in a similar way to t.request, but uses fetch instead.
// Currently supports public tests.
// Must be provided with url and method objects
// Example :
// {
//   url: `${testcafeconfig.api}albums`,
//   method: 'get',
//   params: {
//     count: Limit,
//     offset: xOffset
//   }
// }
async function helperRequest(requestOptions) {
  const rURL = new URL(requestOptions.url);
  if (requestOptions.params) {
    Object.entries(requestOptions.params).forEach(([key, value]) => {
      rURL.searchParams.append(key, value);
    });
  }
  
  let rOption = {
    method: requestOptions.method
  }

  if (requestOptions.body) {
    rOption.body = requestOptions.body;
  }

  const response = await fetch(rURL, rOption);

  let headers = {};
  if (response.headers) {
    response.headers.forEach((value, key) => {
      headers[key] = value;
    });
  }

  if (response.ok) {
    const result = await response.json();
    return {
      status: response.status,
      statusText: response.statusText,
      headers: headers,
      body: result
    };
  } else {
    try {
      const result = await response.json();
      return {
        status: response.status,
        statusText: response.statusText,
        headers: headers,
        body: result
      };
    } catch {
      return response;
    }
  }
}

// getAllAlbumPhotos uses Fetch for get all the basic photo information of photos that are in the album identified by it's UID (albumUID).
async function getAllAlbumPhotos(albumUID) {
  let photos = [];
  const Limit = 50;
  let xCount = Limit;
  let xOffset = 0;

  while (xCount === Limit) {
    const photoApiResponse = await helperRequest({
      url: `${testcafeconfig.api}photos`,
      params: {
        count: Limit,
        offset: xOffset,
        s: albumUID
      }
    });
    xOffset += Limit;
    if (photoApiResponse.status === 200) {
      xCount = Number(photoApiResponse.headers["x-count"]);
      photos.push(...photoApiResponse.body);
    } else {
      xCount = 0;
      throw new Error(`getAllPhotos failed with status ${photoApiResponse.status} and ${photoApiResponse.statusText}`);
    }
  }
  return photos;
}

// getAllAlbums populates the passed in array with all the albums details
// returns empty string on success, or error message.
async function getAllAlbums(albums) {
  albums.length = 0;
  let searchApiResponse;
  const Limit = 50;
  let xCount = Limit;
  let xOffset = 0;
  // Snapshot all albums, and the photos in them
  while (xCount === Limit) {
    searchApiResponse = await helperRequest({
      url: `${testcafeconfig.api}albums`,
      method: 'get',
      params: {
        count: Limit,
        offset: xOffset
      }
    });
    xOffset += Limit;
    if (searchApiResponse.status === 200) {
      xCount = Number(searchApiResponse.headers["x-count"]);
      for (const album of searchApiResponse.body) {
        const photos = await getAllAlbumPhotos(album.UID);
        const rAlbum = {
          "uid": album.UID,
          "data": album,
          "photos": photos
        }
        albums.push(rAlbum);
      }
    } else {
      const msg = "getAllAlbums gather albums " + JSON.stringify(searchApiResponse);
      logMessage(msg);
      return msg;
    }
  }
  return '';
}

// getAllLabels populates the passed in array with all the label details
// returns empty string on success, or error message.
async function getAllLabels(labels) {
  labels.length = 0;
  let searchApiResponse;
  const Limit = 50;
  let xCount = Limit;
  let xOffset = 0;
  // Snapshot all labels (the photo snapshot captures which labels are for which photos)
  while (xCount === Limit) {
    searchApiResponse = await helperRequest({
      url: `${testcafeconfig.api}labels`,
      method: 'get',
      params: {
        count: Limit,
        offset: xOffset,
        all: true
      }
    });
    xOffset += Limit;
    if (searchApiResponse.status === 200) {
      xCount = searchApiResponse.body.length;  // labels does not return an x-count as at 2026-06-04.
      for (const label of searchApiResponse.body) {
        const rLabel = {
          "uid": label.UID,
          "data": label
        }
        labels.push(rLabel);
      }
    } else {
      const msg = "getAllLabels gather labels " + JSON.stringify(searchApiResponse);
      logMessage(msg);
      return msg;
    }
  }
  return '';
}

// getAllPhotos populates the passed in array with all the photo details
// returns empty string on success, or error message.
async function getAllPhotos(photos) {
  photos.length = 0;
  let searchApiResponse;
  const Limit = 50;
  let xCount = Limit;
  let xOffset = 0;
  // There are 110 photos with files in Acceptance database
  // primary:true public:false - 104 (includes review photos)
  // primary:true archived:true - 6 (archived:true overrides public:false, so only archived photos are returned.)
  // Snapshot non archived photos, getting all their details (2nd call per photo required)
  while (xCount === Limit) {
    searchApiResponse = await helperRequest({
      url: `${testcafeconfig.api}photos`,
      method: 'get',
      params: {
        count: Limit,
        offset: xOffset,
        q: "primary:true public:false"
      }
    });
    xOffset += Limit;
    if (searchApiResponse.status === 200) {
      xCount = Number(searchApiResponse.headers["x-count"]);
      for (const photo of searchApiResponse.body) {
        const photoApiResponse = await helperRequest({
          url: `${testcafeconfig.api}photos/${photo.UID}`,
          method: 'get'
        });
        if (photoApiResponse.status === 200) {
          const rPhoto = {
            "uid": photo.UID,
            "data": photoApiResponse.body
          }          
          photos.push(rPhoto);
        } else {
          const msg = "getAllPhotos query photo " + JSON.stringify(photoApiResponse);
          logMessage(msg);
          return msg;
        }
      }
    } else {
      const msg = "getAllPhotos gather photos " + JSON.stringify(searchApiResponse);
      logMessage(msg);
      return msg;
    }
  }
  // Snapshot archived photos, getting all their details (2nd call per photo required)
  xCount = Limit;
  xOffset = 0;
  while (xCount === Limit) {
    searchApiResponse = await helperRequest({
      url: `${testcafeconfig.api}photos`,
      method: 'get',
      params: {
        count: Limit,
        offset: xOffset,
        q: "primary:true archived:true"
      }
    });
    xOffset += Limit;
    if (searchApiResponse.status === 200) {
      xCount = Number(searchApiResponse.headers["x-count"]);
      for (const photo of searchApiResponse.body) {
        const photoApiResponse = await helperRequest({
          url: `${testcafeconfig.api}photos/${photo.UID}`,
          method: 'get'
        });
        if (photoApiResponse.status === 200) {
          const rPhoto = {
            "uid": photo.UID,
            "data": photoApiResponse.body
          }          
          photos.push(rPhoto);
        } else {
          const msg = "getAllPhotos query archived photo " + JSON.stringify(photoApiResponse);
          logMessage(msg);
          return msg;
        }
      }
    } else {
      const msg = "getAllPhotos gather archived photos " + JSON.stringify(searchApiResponse);
      logMessage(msg);
      return msg;
    }
  }
  return '';
}


// helperBeforeFixture will setup the context for all the tests, by taking a snapshot of the before fixture state.
export async function helperBeforeFixture(ctx) {
  logMessage("helperBeforeFixture");
  let snapshotAlbums = [];
  let snapshotLabels = [];
  let snapshotPhotos = [];

  // Snapshot all albums, and the photos in them
  let result = await getAllAlbums(snapshotAlbums);
  if (result !== '') {
    throw new Error(result);
  }

  // Snapshot all labels
  result = await getAllLabels(snapshotLabels);
  if (result !== '') {
    throw new Error(result);
  }

  // Snapshot all photos
  result = await getAllPhotos(snapshotPhotos);
  if (result !== '') {
    throw new Error(result);
  }

  // Store all the snapshots for use in the beforeEach/afterEach functions
  ctx.snapshots = {
    "snapshotAlbums": snapshotAlbums,
    "snapshotLabels": snapshotLabels,
    "snapshotPhotos": snapshotPhotos
  }
}

export async function helperBeforeEach(t) {
  logMessage("helperBeforeEach");

    let startTimestamp = new Date();
  startTimestamp.setMilliseconds(0);
  let helperFailures = [];

  t.ctx.testChanges = {
    "startTimestamp": startTimestamp.toISOString(),
    "revertAlbums": [],
    "revertPhotos": [],
    "removeAlbums": [],
    "removeLabels": [],
    "removePhotos": []
  }
  await t.expect(helperFailures).eql([]);
}

// deepEqual attempts to determine if 2 json objects are different.
function deepEqual(x, y) {
  const ignoreKeys = ["UpdatedAt", "EditedAt", "CheckedAt", "EstimatedAt", "IndexedAt", "ThumbSrc", "LabelSrc"]; // Ignore these keys as we are looking for real changes.
  return x && y && typeof x === 'object' && typeof x === typeof y ? (
    Object.keys(x).filter(key => !ignoreKeys.includes(key)).length === Object.keys(y).filter(key => !ignoreKeys.includes(key)).length &&
      Object.keys(x).every(key => {
        if (ignoreKeys.includes(key)) {
          return true;
        } else {
          return deepEqual(x[key], y[key]);
        }
      })
  ) : (x === y);
}

// determineChangedAlbums works out which albums have been changed since the beforeFixture was run.
async function determineChangedAlbums(t) {
  const beforeTimestamp = new Date(t.ctx.testChanges.startTimestamp);
  let currentAlbums = [];

  // Snapshot all albums, and the photos in them
  const result = await getAllAlbums(currentAlbums);
  if (result !== '') {
    throw new Error(result);
  }

  for (const currentAlbum of currentAlbums) {
    if (new Date(currentAlbum.data.CreatedAt) >= beforeTimestamp) {
      if (!t.ctx.testChanges.removeAlbums.some(ra => ra.uid === currentAlbum.uid)) {
        const removeAlbum = {
          "name": "uid",
          "uid": currentAlbum.uid
        }
        t.ctx.testChanges.removeAlbums.push(removeAlbum);
      }      
    } else {
      const rAlbum = t.fixtureCtx.snapshots.snapshotAlbums.find((a) => a.uid === currentAlbum.uid);
      if (rAlbum) {
        if (!deepEqual(currentAlbum, rAlbum)) {
          t.ctx.testChanges.revertAlbums.push(rAlbum);
        }
      } else {
        const msg = `helperDetermineChangedItems revert albums (1) ${currentAlbum.uid} is missing from snapshot albums`;
        logMessage(msg);
        return msg;
      }
    }
  }
  let foundAlbums = [];
  foundAlbums.push(...currentAlbums.map(album => album.uid));
  // Find albums that have been deleted, and flag them for reversion.
  for (const album of t.fixtureCtx.snapshots.snapshotAlbums) {
    if (!foundAlbums.includes(album.uid)) {
      t.ctx.testChanges.revertAlbums.push(album);
    }
  }
  return '';
}

// determineChangedLabels works out which labels have been changed since the beforeFixture was run.
async function determineChangedLabels(t) {
  const beforeTimestamp = new Date(t.ctx.testChanges.startTimestamp);
  let currentLabels = [];

  // Snapshot all labels
  const result = await getAllLabels(currentLabels);
  if (result !== '') {
    throw new Error(result);
  }

  for (const currentLabel of currentLabels) {
    if (new Date(currentLabel.data.CreatedAt) >= beforeTimestamp) {
      if (!t.ctx.testChanges.removeLabels.some(rl => rl.uid === currentLabel.uid)) {
        const removeLabel = {
          "uid": currentLabel.uid
        }
        t.ctx.testChanges.removeLabels.push(removeLabel);
      }      
    }
  }
  return '';
}

// determineChangedPhotos works out which photos have been changed since the beforeFixture was run.
// this function stores the snapshot information about a photo that will need to be reverted into the context.
// Known issues after executing helperAfterEach:
// Automatically generated titles may be updated to match new format (over old in acceptance data), or reflect labels reverted
// Labels will change type and uncertainty if they were not manual and need to be reverted
// Updated timestamps will change
// Can not restack a file that has been unstacked from a photo
// Can not undelete a file that has been deleted from a photo
// Can NOT revert a photo that has Quality < 3, because the quality will be 3 after reversion due to edits applied.
//   Please note that photos of quality < 3 will not be selected for reversion.

async function determineChangedPhotos(t) {
  const beforeTimestamp = new Date(t.ctx.testChanges.startTimestamp);
  let currentPhotos = [];

  // Snapshot all photos
  const result = await getAllPhotos(currentPhotos);
  if (result !== '') {
    throw new Error(result);
  }

  for (const currentPhoto of currentPhotos) {
    if (new Date(currentPhoto.data.CreatedAt) >= beforeTimestamp) {
      t.ctx.testChanges.removePhotos.push(currentPhoto);
    } else {
      const rPhoto = t.fixtureCtx.snapshots.snapshotPhotos.find((a) => a.uid === currentPhoto.uid);
      if (rPhoto) {
        if (!deepEqual(currentPhoto, rPhoto)) {
          if (rPhoto.data.Quality > 2) {  // all photos have Quality, so null check is not required
            t.ctx.testChanges.revertPhotos.push(rPhoto);
          }
        }
      } else {
        // This is the result of an unstack call /api/v1/photos/{uid}/files/{fileuid}/unstack
        // There is no API to revert this :-(
        const msg = `Unable to revert photo ${currentPhoto.uid} as it has been unstacked from another photo.`;
        logMessage(msg);
      }
    }
  }
  return '';
}

// This function will undo what the test has done (to the best of it's ability)
// as requested by the helperRemove and helperRevert functions.
export async function helperAfterEach(t) {
  logMessage("helperAfterEach");

  let helperFailures = [];
  
  let result = await determineChangedLabels(t);
  if (result !== ""){
    throw new Error(result);
  }

  // Remove Labels
  try {
    if (t.ctx.testChanges.removeLabels.length > 0) {
      let labels = [];
      for (const removeLabel of t.ctx.testChanges.removeLabels) {
        labels.push(removeLabel.uid);
      }
      if (labels.length > 0) {
        const apiResponse = await t.request({
          url: `${testcafeconfig.api}batch/labels/delete`,
          method: 'post',
          body: {
            "labels": labels
          }
        });
        if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach delete labels " + JSON.stringify(apiResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }
    }
  } catch (e) {
    const errorText = e.errmsg || e.message || "An unknown error occurred";
    helperFailures.push(`removeLabels threw ${errorText}`);
  }

  result = await determineChangedPhotos(t);
  if (result !== ""){
    throw new Error(result);
  }

  // Revert Photos state
  // this can not fully restore labels to a photo.
  // if the label has been fully removed, and it matches a keyword, then it will 
  // be restored as a keyword based label.  Otherwise it will be a manual
  // style label.
  try {
    for (const revertPhoto of t.ctx.testChanges.revertPhotos) {
      // Get current photo status
      let apiResponse = await t.request({
        url: `${testcafeconfig.api}photos/${revertPhoto.uid}`,
        method: 'get'
      });
      if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
        const msg = "helperAfterEach revert photo (1) " + JSON.stringify(apiResponse);
        logMessage(msg);
        helperFailures.push(msg);
      }

      if (!revertPhoto.data.DeletedAt && apiResponse.body.DeletedAt) {
        // Need to restore the photo
        const restoreResponse = await t.request({
          url: `${testcafeconfig.api}batch/photos/restore`,
          method: 'post',
          body: {
            "photos": [ revertPhoto.uid ]
          }
        });
        if (restoreResponse.status !== 200 || restoreResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach revert restore photo " + JSON.stringify(restoreResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }

      // Revert the photo
      apiResponse = await t.request({
        url: `${testcafeconfig.api}photos/${revertPhoto.uid}`,
        method: 'put',
        body: revertPhoto.data
      });
      if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
        const msg = "helperAfterEach revert photo (2) " + JSON.stringify(apiResponse);
        logMessage(msg);
        helperFailures.push(msg);
      }

      // Loop through the labels in revertPhoto.data and apiResponse.body to add/remove as needed.
      // Remove
      for (const label of apiResponse.body.Labels) {
        const exists = revertPhoto.data.Labels.find(slug => slug.Label.Slug === label.Label.Slug);
        if (!exists) {
          let labelApiResponse = await t.request({
            url: `${testcafeconfig.api}photos/${revertPhoto.uid}/label/${label.LabelID}`,
            method: 'put',
            body: {
                "Uncertainty": 10, // Need to set Uncertainty < 100, and LabelSrc to manual so it gets soft deleted!
                "LabelSrc": "manual"
            }
          });
          if (labelApiResponse.status !== 200 || labelApiResponse.status === null) { // Ignore Ok
            const msg = "helperAfterEach manual label " + JSON.stringify(labelApiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
          labelApiResponse = await t.request({
            url: `${testcafeconfig.api}photos/${revertPhoto.uid}/label/${label.LabelID}`,
            method: 'delete'
          });
          if ((labelApiResponse.status !== 200 && labelApiResponse.status !== 404) || labelApiResponse.status === null ) { // Ignore Ok and not found
            const msg = "helperAfterEach remove label from photo " + JSON.stringify(labelApiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
        }
      }
      // Add
      for (const label of revertPhoto.data.Labels) {
        const exists = apiResponse.body.Labels.find(slug => slug.Label.Slug === label.Label.Slug);
        if (!exists) {
          const labelApiResponse = await t.request({
            url: `${testcafeconfig.api}photos/${revertPhoto.uid}/label`,
            method: 'post',
            body: {
                "Description": label.Label.Description,
                "Favorite": label.Label.Favorite,
                "Name": label.Label.Name,
                "Notes": label.Label.Notes,
                "Priority": label.Label.Priority,
                "Thumb": label.Label.Thumb,
                // "ThumbSrc": label.Label.ThumbSrc,
                "Uncertainty": label.Uncertainty,
                "LabelSrc": "manual" // Reverted ThumbSrc will always be manual due to API.
            }
          });
          if (labelApiResponse.status !== 200 || labelApiResponse.status === null) { // Ignore Ok
            const msg = "helperAfterEach add label " + JSON.stringify(labelApiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
        } else {
          let labelApiResponse;
          labelApiResponse = await t.request({
            url: `${testcafeconfig.api}photos/${revertPhoto.uid}/label/${label.LabelID}`,
            method: 'put',
            body: {
                "Description": label.Label.Description,
                "Favorite": label.Label.Favorite,
                "Name": label.Label.Name,
                "Notes": label.Label.Notes,
                "Priority": label.Label.Priority,
                "Thumb": label.Label.Thumb,
                // "ThumbSrc": label.Label.ThumbSrc,
                "Uncertainty": label.Uncertainty // Although this doesn't match the previous number, it forces a manual label back into place.  All that can be done.
            }
          });
          if (labelApiResponse.status !== 200 || labelApiResponse.status === null) { // Ignore Ok
            const msg = "helperAfterEach reset label " + JSON.stringify(labelApiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
        }
      }

      // Loop through the Albums in revertPhoto.data and apiResponse.body to add/remove as needed.
      // Remove
      for (const album of apiResponse.body.Albums) {
        const exists = revertPhoto.data.Albums.some(slug => slug.Slug === album.Slug);
        if (!exists) {
          const albumApiResponse = await t.request({
            url: `${testcafeconfig.api}albums/${album.UID}/photos`,
            method: 'delete',
            body: {
              "photos": [ revertPhoto.uid ]
            }
          });
          if (albumApiResponse.status !== 200 || albumApiResponse.status === null) { // Ignore Ok
            const msg = "helperAfterEach delete from album " + JSON.stringify(albumApiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
        }
      }
      // Add
      for (const album of revertPhoto.data.Albums) {
        const exists = apiResponse.body.Albums.some(slug => slug.UID === album.UID);
        if (exists) {
          const albumApiResponse = await t.request({
            url: `${testcafeconfig.api}albums/${album.UID}/photos`,
            method: 'post',
            body: {
              "photos": [ revertPhoto.uid ]
            }
          });
          if (albumApiResponse.status !== 200 || albumApiResponse.status === null) { // Ignore Ok
            const msg = `helperAfterEach add to album ${album.UID} ${JSON.stringify(albumApiResponse)}`;
            logMessage(msg);
            helperFailures.push(msg);
          }
        }
      }

      // Loop through the files and markers to update as required
      // Invalidate any that shouldn't be there.
      for (const file of apiResponse.body.Files) {
        const rFile = revertPhoto.data.Files.find(fileI => fileI.UID === file.UID)
        if (rFile) {
          for (const marker of file.Markers) {
            const rMarker = rFile.Markers.find(m => m.UID === marker.UID && m.FileUID === marker.FileUID);
            let markerApiResponse;
            if (rMarker) {
              // reset
              markerApiResponse = await t.request({
                url: `${testcafeconfig.api}markers/${rMarker.UID}`,
                method: 'put',
                body: rMarker
              });
            } else {
              // inactivate
              markerApiResponse = await t.request({
                url: `${testcafeconfig.api}markers/${marker.UID}`,
                method: 'put',
                body: {
                  "Invalid":true
                }
              });
            }
            if (markerApiResponse.status !== 200 || markerApiResponse.status === null) { // Ignore Ok
              const msg = "helperAfterEach sync markers (1) file " + marker.FileUID + " marker " + marker.UID + " " + JSON.stringify(markerApiResponse);
              logMessage(msg);
              helperFailures.push(msg);
            }
          }
        } else {
          const msg = `Choosing not to remove file ${file.UID} which has been added, as that will break future tests as the file is physically deleted.`;
          logMessage(msg);
          helperFailures.push(msg);
        }
      }
      for (const file of revertPhoto.data.Files) {
        const cFile = apiResponse.body.Files.find(fileI => fileI.UID === file.UID)
        if (cFile) {
          for (const marker of file.Markers) {
            // Restore the marker whether it is there or not.
            const markerApiResponse = await t.request({
                url: `${testcafeconfig.api}markers/${marker.UID}`,
                method: 'put',
                body: marker
              });
            if (markerApiResponse.status !== 200 || markerApiResponse.status === null) { // Ignore Ok
              const msg = "helperAfterEach sync markers (2)" + JSON.stringify(markerApiResponse);
              logMessage(msg);
              helperFailures.push(msg);
            }
          }
        } else {
          // This is the result of an unstack call /api/v1/photos/{uid}/files/{fileuid}/unstack
          // There is no API to revert this :-(
          const msg = `Unable to restore file ${file.UID} which has been unstacked from photo ${revertPhoto.uid}.`;
          logMessage(msg);
        }
      }

      // Revert any changes to Primary file.
      const originalPrimary = revertPhoto.data.Files.find((element) => element.Primary === true)
      const currentPrimary = apiResponse.body.Files.find((element) => element.Primary === true)
      if (originalPrimary && currentPrimary) {
        const originalUID = originalPrimary.UID
        const currentUID = currentPrimary.UID
        if (originalUID !== currentUID) {
          const primaryApiResponse = await t.request({
            url: `${testcafeconfig.api}photos/${revertPhoto.uid}/files/${originalUID}/primary`,
            method: 'post'
          });
          if (primaryApiResponse.status !== 200 || primaryApiResponse.status === null) { // Ignore Ok
            const msg = "helperAfterEach revert photo primary " + JSON.stringify(primaryApiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
        }
      }

      // Do the photo again to try the Title again.
      apiResponse = await t.request({
        url: `${testcafeconfig.api}photos/${revertPhoto.uid}`,
        method: 'put',
        body: revertPhoto.data
      });
      if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
        const msg = "helperAfterEach revert photo again " + JSON.stringify(apiResponse);
        logMessage(msg);
        helperFailures.push(msg);
      }
      if (revertPhoto.data.DeletedAt && !apiResponse.body.DeletedAt) {
        // Need to archive the photo
        const archiveResponse = await t.request({
          url: `${testcafeconfig.api}batch/photos/archive`,
          method: 'post',
          body: {
            "photos": [ revertPhoto.uid ]
          }
        });
        if (archiveResponse.status !== 200 || archiveResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach revert archive photo " + JSON.stringify(archiveResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }

    }
  } catch (e) {
    const errorText = e.errmsg || e.message || "An unknown error occurred";
    helperFailures.push(`revertPhoto threw ${errorText}`);
  }

  // Archive any new photos
  let deletePhotos = [];
  try {
    for (const removePhoto of t.ctx.testChanges.removePhotos) {
      // Get current photo status
      let apiResponse = await t.request({
        url: `${testcafeconfig.api}photos/${removePhoto.uid}`,
        method: 'get'
      });
      if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
        const msg = "helperAfterEach removePhoto get " + JSON.stringify(apiResponse);
        logMessage(msg);
        helperFailures.push(msg);
      }
      if (!apiResponse.body.DeletedAt) {
        deletePhotos.push(removePhoto.uid);
      }
    }
  
    if (deletePhotos.length > 0) {
        let apiResponse = await t.request({
          url: `${testcafeconfig.api}batch/photos/archive`, // Can NOT use delete, as this will remove the associated file from the file system, which will break all subsequent test suite executions.
          method: 'post',
          body: { "photos": deletePhotos }
        });
        if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach removePhoto archive " + JSON.stringify(apiResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
    }
  } catch (e) {
    const errorText = e.errmsg || e.message || "An unknown error occurred";
    helperFailures.push(`removePhoto threw ${errorText}`);
  }

  result = await determineChangedAlbums(t);
  if (result !== ""){
    throw new Error(result);
  }

  // Revert Albums state
  // This MAY result in a different UID if the album has been deleted, and it wasn't created by the current user.
  try {
    for (let revertAlbum of t.ctx.testChanges.revertAlbums) {
      let apiResponse = await t.request({
        url: `${testcafeconfig.api}albums/${revertAlbum.uid}`,
        method: 'put',
        body: revertAlbum.data
      });
      if (apiResponse.status === 404) {
        let apiPostResponse = await t.request({
          url: `${testcafeconfig.api}albums`,
          method: 'post',
          body: revertAlbum.data
        });

        if (apiPostResponse.status === 201) { // The Album has been created with a different UID!
          revertAlbum.data.UID = apiPostResponse.body.UID;
          revertAlbum.data.ID = apiPostResponse.body.ID;
          revertAlbum.uid = apiPostResponse.body.UID;
          if (revertAlbum.data.Thumb) {
            revertAlbum.data.ThumbSrc = "manual"; // To revert a thumb it must be manual
          }
          // Updating the NEW album
          apiResponse = await t.request({
            url: `${testcafeconfig.api}albums/${revertAlbum.uid}`,
            method: 'put',
            body: revertAlbum.data
          });
          if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
            const msg = "helperAfterEach revert albums (1) " + JSON.stringify(apiResponse);
            logMessage(msg);
            helperFailures.push(msg);
          }
        }
      } else {
        if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach revert albums (2) " + JSON.stringify(apiResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }
      // Restore the photos connections
      const albumPhotoApiResponse = await t.request(`${testcafeconfig.api}photos?count=50&offset=0&s=${revertAlbum.uid}`);

      let photos = [];
      for (const photo of revertAlbum.photos) {
        if (!albumPhotoApiResponse.body.find(ap => ap.UID === photo.UID))
        {
          photos.push(photo.UID);
        }
      }
      if (photos.length > 0){
        const photoApiResponse = await t.request({
          url: `${testcafeconfig.api}albums/${revertAlbum.uid}/photos`,
          method: 'post',
          body: { "photos": photos }
        });
        if (photoApiResponse.status !== 200 || photoApiResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach revert albums photos " + JSON.stringify(photoApiResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }
      apiResponse = await t.request({
        url: `${testcafeconfig.api}albums/${revertAlbum.uid}`,
        method: 'get'
      });
      if (revertAlbum.data.Thumb !== apiResponse.body.Thumb) {
        // Try updating the album again in case the thumb was from a removed photo.
        if (revertAlbum.data.Thumb) {
          revertAlbum.data.ThumbSrc = "manual"; // To revert a thumb it must be manual
        }

        let apiResponse = await t.request({
          url: `${testcafeconfig.api}albums/${revertAlbum.uid}`,
          method: 'put',
          body: revertAlbum.data
        });
        if (apiResponse.status !== 200 || apiResponse.status === null) { // Ignore Ok
          const msg = "helperAfterEach revert albums with thumb manual " + JSON.stringify(apiResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }
    }
  } catch (e) {
    const errorText = e.errmsg || e.message || "An unknown error occurred";
    helperFailures.push(`revertAlbum threw ${errorText}`);
  }


  // Remove albums
  try {
    for (const removeAlbum of t.ctx.testChanges.removeAlbums) {
      let listApiResponse;
      if (removeAlbum.uid === "name") {
        listApiResponse = await t.request({
          url: `${testcafeconfig.api}albums`,
          method: 'get',
          params: {
            count: 10,
            q: `${removeAlbum.name} type:album`
          }
        });
      } else {
        listApiResponse = await t.request({
          url: `${testcafeconfig.api}albums`,
          method: 'get',
          params: {
            count: 10,
            q: `uid:${removeAlbum.uid}`
          }
        });
      }
      if (listApiResponse.status !== 200 || listApiResponse.status === null) { // Ignore Ok
        const msg = "helperAfterEach list albums " + JSON.stringify(listApiResponse);
        logMessage(msg);
        helperFailures.push(msg);
      }
      for (const album of listApiResponse.body) {
        const apiResponse = await t.request({
          url: `${testcafeconfig.api}albums/${album.UID}`,
          method: 'delete',
          params: {
              force: true
          }
        });
        if (apiResponse.status !== 200 || apiResponse.status === null && apiResponse.status !== 404) { // Ignore Ok and not found
          const msg = "helperAfterEach delete album " + JSON.stringify(apiResponse);
          logMessage(msg);
          helperFailures.push(msg);
        }
      }
    }
  } catch (e) {
    const errorText = e.errmsg || e.message || "An unknown error occurred";
    helperFailures.push(`removeAlbums threw ${errorText}`);
  }

  // Error if there were any API or try/catch failures.
  await t.expect(helperFailures).eql([]);
}