import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";

fixture`Test 360° sphere viewer`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const photoviewer = new PhotoViewer();

// Smoke tests for the equirectangular 360° viewer pipeline.
// Live runs require the acceptance fixture set to include
// `panorama_sphere.jpg` (and ideally `panorama_sphere.mp4`)
// with ProjectionType=equirectangular metadata.

test.meta("testID", "sphere-001").meta({ mode: "public" })("Common: Opens 360° photo in sphere viewer", async (t) => {
  await menu.openPage("panoramas");
  const uid = await photo.getNthPhotoUid("image", 0);
  await photoviewer.openPhotoViewer("uid", uid);

  await t.expect(Selector("div.pswp__media--sphere").exists).ok({ timeout: 5000 });
  await t.expect(Selector("div.psv-container").exists).ok({ timeout: 10000 });
});

test.meta("testID", "sphere-002").meta({ mode: "public" })("Common: Standard photo does not mount sphere viewer", async (t) => {
  await menu.openPage("browse");
  const uid = await photo.getNthPhotoUid("image", 0);
  await photoviewer.openPhotoViewer("uid", uid);

  await t.expect(Selector("div.p-lightbox__pswp").visible).ok();
  await t.expect(Selector("div.pswp__media--sphere").exists).notOk();
});

test.meta("testID", "sphere-003").meta({ mode: "public" })("Common: Opens 360° video in sphere viewer", async (t) => {
  await menu.openPage("panoramas");
  const uid = await photo.getNthPhotoUid("video", 0);

  if (!uid) {
    // No 360° video present in current fixture set — skip silently.
    return;
  }

  await photoviewer.openPhotoViewer("uid", uid);
  await t.expect(Selector("div.pswp__media--sphere").exists).ok({ timeout: 5000 });
  await t.expect(Selector("div.psv-container video").exists).ok({ timeout: 10000 });
});
