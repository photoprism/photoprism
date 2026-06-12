import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Toolbar from "../page-model/toolbar";
import Photo from "../page-model/photo";
import PhotoViewer from "../page-model/photoviewer";

// Multi-page PDF viewer acceptance tests. These exercise the lightbox wiring
// (document slide -> interactive viewer overlay, non-document slides untouched,
// arrow keys keep navigating media). Full page-render assertions additionally
// require the original PDF to be present in the acceptance storage archive.
fixture`Test multi-page PDF viewer`.page`${testcafeconfig.url}`;

const toolbar = new Toolbar();
const photo = new Photo();
const photoviewer = new PhotoViewer();

const pdfViewer = Selector("div.p-pdf-viewer", { timeout: 15000 });
const pdfControls = Selector(".p-pdf-viewer__controls", { timeout: 15000 });
const pdfThumbs = Selector(".p-pdf-viewer__thumbs", { timeout: 15000 });

test.meta("testID", "pdf-001").meta({ type: "short", mode: "public" })(
  "Common: Open a multi-page PDF in the interactive viewer",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    await toolbar.search("type:document");
    const uid = await photo.getNthPhotoUid("document", 0);
    await t.expect(uid).ok("expected an indexed PDF document in the test library");
    await photoviewer.openPhotoViewer("uid", uid);
    await t
      .expect(pdfViewer.visible)
      .ok("the PDF viewer overlay should replace the static cover")
      .expect(pdfControls.exists)
      .ok()
      .expect(pdfThumbs.exists)
      .ok()
      // The viewer owns its own zoom, so the default PhotoSwipe zoom button is hidden.
      .expect(Selector("button.pswp__button--zoom").visible)
      .notOk();
  }
);

test.meta("testID", "pdf-002").meta({ type: "short", mode: "public" })(
  "Common: Standard images do not load the PDF viewer",
  async (t) => {
    await t.click(toolbar.cardsViewAction);
    await toolbar.search("type:image");
    const uid = await photo.getNthPhotoUid("image", 0);
    await photoviewer.openPhotoViewer("uid", uid);
    await t.expect(photoviewer.viewer.visible).ok().expect(pdfViewer.exists).notOk();
  }
);

test.meta("testID", "pdf-003").meta({ type: "short", mode: "public" })(
  "Common: Arrow keys navigate media items while a PDF is open",
  async (t) => {
    // Open the document from an unfiltered view so it has neighboring media items.
    await t.click(toolbar.cardsViewAction);
    await toolbar.search("type:document");
    const uid = await photo.getNthPhotoUid("document", 0);
    await photoviewer.openPhotoViewer("uid", uid);
    await t.expect(pdfViewer.visible).ok();
    // The Right arrow advances to the next media item (not the next PDF page);
    // when the document is the only result it simply stays on the same slide.
    await t.pressKey("right").expect(photoviewer.viewer.visible).ok();
  }
);
