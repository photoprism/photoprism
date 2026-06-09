import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import { helperBeforeFixture, helperBeforeEach, helperAfterEach, logMessage } from "../page-model/helpers";

fixture`Test helper`
.page`${testcafeconfig.url}`
.before(async ctx => {
  await helperBeforeFixture(ctx);
});

function cleanAlbumsFilesAndLabels(jsonBody) {
  let items = JSON.parse(JSON.stringify(jsonBody.Albums));
  jsonBody.Albums.length = 0;
  for (let item of items) {
    delete item.UpdatedAt;
    delete item.ThumbSrc;
    jsonBody.Albums.push(item);
  }
  items = JSON.parse(JSON.stringify(jsonBody.Files));
  jsonBody.Files.length = 0;
  for (let item of items) {
    delete item.UpdatedAt;
    jsonBody.Files.push(item);
  }
  items = JSON.parse(JSON.stringify(jsonBody.Labels));
  jsonBody.Labels.length = 0;
  for (let item of items) {
    delete item.Label.UpdatedAt;
    delete item.LabelSrc; // can't be set by API
    delete item.Uncertainty; // when Src changes, the uncertainty changes.
    jsonBody.Labels.push(item);
  }
}

test.meta("testID", "cleanup-001").meta({ type: "short", mode: "api" })("Common: Cleanup Album remove album cover photo and change caption/description of existing album", async (t) => {
    await helperBeforeEach(t);
    let beforeAlbumResponse = await t.request({
        url: `${testcafeconfig.api}albums`,
        method: 'get',
        params: {
          count: 1,
          q: `Christmas type:album`
        }
      });
    await t.expect(beforeAlbumResponse.status).eql(200);
    const albumUID = beforeAlbumResponse.body[0].UID;
    beforeAlbumResponse = await t.request(`${testcafeconfig.api}albums/${albumUID}`);
    await t.expect(beforeAlbumResponse.status).eql(200);
    // Change the name on the album
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${albumUID}`,
      method: 'put',
      body: {
        "Caption": "Cleanup test data",
        "Description": "This should be removed"
      }
    });
    await t.expect(apiResponse.status).eql(200);

    // remove pqmxlr7188hz4bih from album "Christmas", which is the album cover photo
    apiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${albumUID}/photos`,
      method: 'delete',
      body: {
        "photos": [ "pqmxlr7188hz4bih" ]  // This should be ok.
      }
    });
    await t.expect(apiResponse.status).eql(200);

    await helperAfterEach(t);
    const afterAlbumResponse = await t.request(`${testcafeconfig.api}albums/${albumUID}`);
    await t.expect(beforeAlbumResponse).notEql(afterAlbumResponse);
    // Remove the fields that are impacted by changes
    delete beforeAlbumResponse.body.UpdatedAt; // Will change
    delete afterAlbumResponse.body.UpdatedAt;
    delete beforeAlbumResponse.body.ThumbSrc; // Will change on 1st run, wont on subsequent
    delete afterAlbumResponse.body.ThumbSrc;
    delete beforeAlbumResponse.headers["content-length"]; // Will change (timestamp)
    delete afterAlbumResponse.headers["content-length"];
    delete beforeAlbumResponse.headers.date; // May change if second ticks over
    delete afterAlbumResponse.headers.date;
    await t.expect(afterAlbumResponse).eql(beforeAlbumResponse);
});

test.meta("testID", "cleanup-002").meta({ type: "short", mode: "api" })("Common: Cleanup Album remove Garden album", async (t) => {
    await helperBeforeEach(t);
    const albumUID = "arkgush1tdwk4fsy";
    let beforeAlbumResponse = await t.request(`${testcafeconfig.api}albums/${albumUID}`);
    await t.expect(beforeAlbumResponse.status).eql(200);
    // Soft Delete the album
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${albumUID}`,
      method: 'delete',
      params: {
        force: false
      }      
    });
    await t.expect(apiResponse.status).eql(200);

    await helperAfterEach(t);
    let afterAlbumResponse = await t.request(`${testcafeconfig.api}albums/${albumUID}`);
    await t.expect(beforeAlbumResponse).notEql(afterAlbumResponse);
    await t.expect(afterAlbumResponse.status).eql(200); // A deleted album with the SAME CreatedBy as the current user will be undeleted
    // Remove the fields that are impacted by changes
    delete beforeAlbumResponse.body.UpdatedAt;
    delete afterAlbumResponse.body.UpdatedAt;
    delete beforeAlbumResponse.body.ThumbSrc;
    delete afterAlbumResponse.body.ThumbSrc;
    delete beforeAlbumResponse.headers["content-length"];
    delete afterAlbumResponse.headers["content-length"];
    delete beforeAlbumResponse.headers.date;
    delete afterAlbumResponse.headers.date;
    await t.expect(afterAlbumResponse).eql(beforeAlbumResponse);
});

test.meta("testID", "cleanup-003").meta({ type: "short", mode: "api" })("Common: Cleanup Album remove Holiday album owned by other user", async (t) => {
    await helperBeforeEach(t);
    let beforeAlbumResponse = await t.request({
        url: `${testcafeconfig.api}albums`,
        method: 'get',
        params: {
          count: 1,
          q: `Holiday type:album`
        }
      });
    await t.expect(beforeAlbumResponse.status).eql(200);
    const albumUID = beforeAlbumResponse.body[0].UID;
    await t.expect(albumUID).eql("aqmxlt22ilujuxux", "Test requires Holiday album to not have been deleted already");
    beforeAlbumResponse = await t.request(`${testcafeconfig.api}albums/${albumUID}`);
    await t.expect(beforeAlbumResponse.status).eql(200);
    // Soft Delete the album
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${albumUID}`,
      method: 'delete',
      params: {
        force: false
      }      
    });
    await t.expect(apiResponse.status).eql(200);

    await helperAfterEach(t);
    let afterAlbumResponse = await t.request(`${testcafeconfig.api}albums/${albumUID}`);
    await t.expect(beforeAlbumResponse).notEql(afterAlbumResponse);
    await t.expect(afterAlbumResponse.status).eql(404); // A deleted album with a null or different CreatedBy will change UID
    afterAlbumResponse = await t.request({
      url: `${testcafeconfig.api}albums`,
      method: 'get',
      params: {
        count: 1,
        q: `Holiday type:album`
      }
    });
    await t.expect(afterAlbumResponse.status).eql(200);
    afterAlbumResponse = await t.request(`${testcafeconfig.api}albums/${afterAlbumResponse.body[0].UID}`);
    await t.expect(afterAlbumResponse.status).eql(200);
    // Remove the fields that are impacted by changes
    delete beforeAlbumResponse.body.UID;
    delete afterAlbumResponse.body.UID;
    delete beforeAlbumResponse.body.ID;
    delete afterAlbumResponse.body.ID;
    delete beforeAlbumResponse.body.UpdatedAt;
    delete afterAlbumResponse.body.UpdatedAt;
    delete beforeAlbumResponse.body.CreatedAt;
    delete afterAlbumResponse.body.CreatedAt;
    delete beforeAlbumResponse.body.CreatedBy;
    delete afterAlbumResponse.body.CreatedBy;
    delete beforeAlbumResponse.body.ThumbSrc;
    delete afterAlbumResponse.body.ThumbSrc;
    delete beforeAlbumResponse.headers["content-length"];
    delete afterAlbumResponse.headers["content-length"];
    delete beforeAlbumResponse.headers.date;
    delete afterAlbumResponse.headers.date;
    await t.expect(afterAlbumResponse).eql(beforeAlbumResponse);
});

test.meta("testID", "cleanup-004").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert titles and details", async (t) => {
    await helperBeforeEach(t);
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `geo:true camera:apple`
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);
    // Change the name and other stuff on the photo
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}`,
      method: 'put',
      body: {
        "Title": "Cleanup test data",
        "Description": "This should be removed",
        "CameraID": 7,
        "LensID": 10,
        "CellID": "s2:47a85a634bcc",
        "PlaceID": "de:ukLS8nroIoB7"
      }
    });
    await t.expect(apiResponse.status).eql(200);

    await helperAfterEach(t);

    let afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);
})

test.meta("testID", "cleanup-005").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert Labels", async (t) => {
    await helperBeforeEach(t);
    const stamp = Date.now();
    const labelTitle = `CleanupLabel-${stamp}`;
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `label:cat`
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);
    const labelID = 11; // The Cat label! beforePhotoResponse.body.Labels[0].ID;
    // Remove a label from the photo
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label/${labelID}`,
      method: 'delete'
    });
    await t.expect(apiResponse.status).eql(200);

    // Add a manual label
    const labelApiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label`,
      method: 'post',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": labelTitle,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(200);

    console.log(JSON.stringify(labelApiResponse.body));

    await helperAfterEach(t);

    const afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    if (!beforePhotoResponse.body.Title.includes("Cat")) {
      // If the title didn't include cat, then the label change to Manual 100% certainty MAY add cat to the Title.
      delete beforePhotoResponse.body.Title;
      delete afterPhotoResponse.body.Title;
    }
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);
})

// This test will leave a junk label behind if it fails, as it's testing that manual labels are reverted correctly.
// This test is no longer possible due to beforeFixture caching state before this test initiates.
test.meta("testID", "cleanup-006").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert manual deleted Label", async (t) => {
  // This test is not possible as there is no manual labels in the acceptance database.
  // This prevents the required conditions from being there when beforeFixture runs.
  return;
    await helperBeforeEach(t);
    const stamp = Date.now();
    const labelTitle = `CleanupLabel-${stamp}`;
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    // Add a manual label
    let labelApiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label`,
      method: 'post',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": labelTitle,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(200);
    labelApiResponse = await t.request({
      url: `${testcafeconfig.api}labels`,
      method: 'get',
      params: {
        count: 1,
        q: labelTitle
      }
    });
    await t.expect(labelApiResponse.status).eql(200);
    const labelID = labelApiResponse.body[0].ID;

    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);

    // Remove the manual label from the photo
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label/${labelID}`,
      method: 'delete'
    });
    await t.expect(apiResponse.status).eql(200);
    await helperAfterEach(t);

    const afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);

    // Remove the manual label from the photo
    apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label/${labelID}`,
      method: 'delete'
    });
    await t.expect(apiResponse.status).eql(200);
})

// This test will leave a junk album behind if it fails, as it's testing that manual albums are reverted correctly from a photo.
test.meta("testID", "cleanup-007").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert Albums", async (t) => {
    await helperBeforeEach(t);
    const stamp = Date.now();
    const albumTitle = `CleanupAlbum-${stamp}`;
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `album:garden private:false`
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);

    const albumUID = "arkgush1tdwk4fsy";
    // Remove an existing album from the photo
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${albumUID}/photos`,
      method: 'delete',
      body: {
        "photos": [ photoUID ]
      }
    });
    await t.expect(apiResponse.status).eql(200);

    // Add a manual album
    let albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums`,
      method: 'post',
      body: {
          "Caption": "cleanup-007",
          "Title": albumTitle
      }
    });
    await t.expect(albumApiResponse.status).eql(201);

    // Add photo to the new album
    const newAlbumUID = albumApiResponse.body.UID;

    albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${newAlbumUID}/photos`,
      method: 'post',
      body: {
        "photos": [ photoUID ]
      }
    });
    await t.expect(albumApiResponse.status).eql(200);
    logMessage(JSON.stringify(albumApiResponse));

    await helperAfterEach(t);

    const afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);

    // await helperBeforeEach(t);
    // await helperRemoveAlbum(t, "name", albumTitle);
    // await helperAfterEach(t);

    albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${newAlbumUID}`,
      method: 'get'
    });
    await t.expect(albumApiResponse.status).eql(404);
})

// This test can fail due to order of label results changing (which testcafe thinks is an actual change),
// or if a named person is chosen, then the title can change, assuming the original title isn't up to date.
test.meta("testID", "cleanup-008").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert Markers", async (t) => {
    await helperBeforeEach(t);
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `faces:new` // faces:new so that the title can't change, and labels are less likely to change order.
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);

    // Loop files and markers.  Invalidate the 1st marker found.  Store the FileUID and MarkerUID
    let fileUID = "";
    let markerUID = "";

    for (const file of beforePhotoResponse.body.Files){
      for (const marker of file.Markers) {
        if (!marker.Invalid) {
          markerUID = marker.UID;
          fileUID = marker.FileUID;
          break;
        }
      }
      if (markerUID != "") {
        break;
      }
    }

    // Remove an existing marker from the photo
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}markers/${markerUID}`,
      method: 'put',
      body: {
        "Invalid":true
      }
    });
    await t.expect(apiResponse.status).eql(200);

    // Add a manual marker
    let markerApiResponse = await t.request({
      url: `${testcafeconfig.api}markers`,
      method: 'post',
      body: {
        "FileUID":fileUID,
        "Type":"face",
        "Src":"manual",
        "X":0.42172298011844334,
        "Y":0.35562737944162437,
        "W":0.1810358502538071,
        "H":0.13577688769035534
      }
    });
    await t.expect(markerApiResponse.status).eql(201);
    const newMarkerUID = markerApiResponse.body.UID;

    await helperAfterEach(t);

    const afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);

    // need to remove the new marker as it will be invalid=true, not removed.
    for (const file of afterPhotoResponse.body.Files){
      let items = JSON.parse(JSON.stringify(file.Markers));
      file.Markers.length = 0;
      for (let item of items) {
        if (item.UID != newMarkerUID)
        {
          delete item.UpdatedAt;
          file.Markers.push(item);
        } else {
          await t.expect(item.Invalid).eql(true);  // Make sure it was marked invalid.
        }
      }
    }
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);

})

// This test will leave a junk album behind if it fails, as it's testing that manual albums are reverted correctly from a photo.
test.meta("testID", "cleanup-009").meta({ type: "short", mode: "api" })("Common: Cleanup Album", async (t) => {
    await helperBeforeEach(t);
    const stamp = Date.now();
    const albumTitle1 = `CleanupAlbum1-${stamp}`;
    const albumTitle2 = `CleanupAlbum2-${stamp}`;

    // Add a manual album
    let albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums`,
      method: 'post',
      body: {
          "Caption": "cleanup-009",
          "Title": albumTitle1
      }
    });
    await t.expect(albumApiResponse.status).eql(201);
    const newAlbum1UID = albumApiResponse.body.UID;

    // Add a manual album
    albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums`,
      method: 'post',
      body: {
          "Caption": "cleanup-009",
          "Title": albumTitle2
      }
    });
    await t.expect(albumApiResponse.status).eql(201);
    const newAlbum2UID = albumApiResponse.body.UID;

    await helperAfterEach(t);

    albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${newAlbum1UID}`,
      method: 'get'
    });
    await t.expect(albumApiResponse.status).eql(404);
    albumApiResponse = await t.request({
      url: `${testcafeconfig.api}albums/${newAlbum2UID}`,
      method: 'get'
    });
    await t.expect(albumApiResponse.status).eql(404);
})

test.meta("testID", "cleanup-010").meta({ type: "short", mode: "api" })("Common: Cleanup Labels 2 manual labels", async (t) => {
    await helperBeforeEach(t);
    const stamp = Date.now();
    const label1Title = `CleanupLabel1-${stamp}`;
    const label2Title = `CleanupLabel2-${stamp}`;
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `label:beach`
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);

    // Add a manual label
    let labelApiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label`,
      method: 'post',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": label1Title,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(200);

    // Add a manual label
    labelApiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/label`,
      method: 'post',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": label2Title,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(200);

    await helperAfterEach(t);

    const afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse).notEql(afterPhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);
})

test.meta("testID", "cleanup-010").meta({ type: "short", mode: "api" })("Common: Cleanup Labels remove 1 label from 2 photos", async (t) => {
    await helperBeforeEach(t);
    const stamp = Date.now();
    const labelTitle = `CleanupLabel-${stamp}`;
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 2,
          q: `label:zebra`
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    await t.expect(beforePhotoResponse.body.length).eql(2);
    const photo1UID = beforePhotoResponse.body[0].UID;
    const photo2UID = beforePhotoResponse.body[1].UID;

    // Add a manual label
    let labelApiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photo1UID}/label`,
      method: 'post',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": labelTitle,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(200);
    const label1UID = labelApiResponse.body.Labels.find((element) => element.Label.Name == labelTitle).Label.UID;

    // Add a manual label
    labelApiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photo2UID}/label`,
      method: 'post',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": labelTitle,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(200);
    const label2UID = labelApiResponse.body.Labels.find((element) => element.Label.Name == labelTitle).Label.UID;

    await helperAfterEach(t);

    await t.expect(label1UID).eql(label2UID);
    await t.expect(label1UID.length).eql(16);

    labelApiResponse = await t.request({
      url: `${testcafeconfig.api}labels/${label1UID}`,
      method: 'put',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": labelTitle,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(404);

    labelApiResponse = await t.request({
      url: `${testcafeconfig.api}labels/${label2UID}`,
      method: 'put',
      body: {
          "Description": "Testing Label",
          "Favorite": false,
          "Name": labelTitle,
          "Priority": 0,
          "Uncertainty": 0
      }
    });
    await t.expect(labelApiResponse.status).eql(404);

})

test.meta("testID", "cleanup-011").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert archive and restore", async (t) => {
    await helperBeforeEach(t);
    let beforeRestorePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `archived:true photo:yes`
        }
      });
    await t.expect(beforeRestorePhotoResponse.status).eql(200);
    const archivedPhotoUID = beforeRestorePhotoResponse.body[0].UID;
    beforeRestorePhotoResponse = await t.request(`${testcafeconfig.api}photos/${archivedPhotoUID}`);
    await t.expect(beforeRestorePhotoResponse.status).eql(200);

    // Change the name and other stuff on the photo
    let apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${archivedPhotoUID}`,
      method: 'put',
      body: {
        "Title": "Cleanup test data",
        "Description": "This should be removed",
        "CameraID": 7,
        "LensID": 10,
        "CellID": "s2:47a85a634bcc",
        "PlaceID": "de:ukLS8nroIoB7"
      }
    });
    await t.expect(apiResponse.status).eql(200);

    apiResponse = await t.request({
      url: `${testcafeconfig.api}batch/photos/restore`,
      method: 'post',
      body: {
        "photos": [ archivedPhotoUID ]
      }
    });
    await t.expect(apiResponse.status).eql(200);

    let beforeArchivePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `archived:false photo:yes`
        }
      });
    await t.expect(beforeArchivePhotoResponse.status).eql(200);
    const restoredPhotoUID = beforeArchivePhotoResponse.body[0].UID;
    beforeArchivePhotoResponse = await t.request(`${testcafeconfig.api}photos/${restoredPhotoUID}`);
    await t.expect(beforeArchivePhotoResponse.status).eql(200);

    // Change the name and other stuff on the photo
    apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${restoredPhotoUID}`,
      method: 'put',
      body: {
        "Title": "Cleanup test data",
        "Description": "This should be removed",
        "CameraID": 7,
        "LensID": 10,
        "CellID": "s2:47a85a634bcc",
        "PlaceID": "de:ukLS8nroIoB7"
      }
    });
    await t.expect(apiResponse.status).eql(200);

    apiResponse = await t.request({
      url: `${testcafeconfig.api}batch/photos/archive`,
      method: 'post',
      body: {
        "photos": [ restoredPhotoUID ]
      }
    });
    await t.expect(apiResponse.status).eql(200);


    await helperAfterEach(t);

    let afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${archivedPhotoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforeRestorePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforeRestorePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforeRestorePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforeRestorePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforeRestorePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforeRestorePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforeRestorePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    // DeletedAt will be different.
    await t.expect(beforeRestorePhotoResponse.body.DeletedAt).contains("Z");
    await t.expect(afterPhotoResponse.body.DeletedAt).contains("Z");
    delete beforeRestorePhotoResponse.body.DeletedAt;
    delete afterPhotoResponse.body.DeletedAt;
    await t.expect(afterPhotoResponse).eql(beforeRestorePhotoResponse);

    afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${restoredPhotoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforeArchivePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforeArchivePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforeArchivePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforeArchivePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforeArchivePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforeArchivePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforeArchivePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforeArchivePhotoResponse);
})

test.meta("testID", "cleanup-012").meta({ type: "short", mode: "api" })("Common: Cleanup Photo revert primary", async (t) => {
    await helperBeforeEach(t);
    let beforePhotoResponse = await t.request({
        url: `${testcafeconfig.api}photos`,
        method: 'get',
        params: {
          count: 1,
          q: `stacks`
        }
      });
    await t.expect(beforePhotoResponse.status).eql(200);
    const photoUID = beforePhotoResponse.body[0].UID;
    beforePhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(beforePhotoResponse.status).eql(200);

    const currentPrimary = beforePhotoResponse.body.Files.find((element) => element.Primary == true).UID
    const targetPrimary = beforePhotoResponse.body.Files.find((element) => element.Primary == false).UID

    let apiResponse = await t.request({
      url: `${testcafeconfig.api}photos/${photoUID}/files/${targetPrimary}/primary`,
      method: 'post'
    });
    await t.expect(apiResponse.status).eql(200);

    await helperAfterEach(t);

    let afterPhotoResponse = await t.request(`${testcafeconfig.api}photos/${photoUID}`);
    await t.expect(afterPhotoResponse).notEql(beforePhotoResponse);
    // Remove the fields that are impacted by changes
    delete beforePhotoResponse.headers["content-length"]; // Will change (timestamp)
    delete afterPhotoResponse.headers["content-length"];
    delete beforePhotoResponse.headers.date; // May change if second ticks over
    delete afterPhotoResponse.headers.date;
    delete beforePhotoResponse.body.UpdatedAt;
    delete afterPhotoResponse.body.UpdatedAt;
    delete beforePhotoResponse.body.EditedAt;
    delete afterPhotoResponse.body.EditedAt;
    delete beforePhotoResponse.body.Details.UpdatedAt;
    delete afterPhotoResponse.body.Details.UpdatedAt;
    cleanAlbumsFilesAndLabels(beforePhotoResponse.body);
    cleanAlbumsFilesAndLabels(afterPhotoResponse.body);
    await t.expect(afterPhotoResponse).eql(beforePhotoResponse);
})
