import { Selector } from "testcafe";
import testcafeconfig from "./testcafeconfig";
import Page from "./page-model";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import PhotoViews from "../page-model/photo-views";
import Label from "../page-model/label";
import Album from "../page-model/album";
import Subject from "../page-model/subject";
import NewPage from "../page-model/page";
import Originals from "../page-model/originals";

fixture`Test files`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const album = new Album();
const originals = new Originals();

test.meta("testID", "originals-001")("Add original files to album", async (t) => {
  await menu.openPage("albums");
  await toolbar.search("KanadaVacation");
  await t.expect(Selector("div.no-results").visible).ok();
  await menu.openPage("originals");
  await t.click(Selector("button").withText("Vacation"));
  await t.wait(15000);
  const FirstItemInVacation = await Selector("div.result", { timeout: 15000 }).nth(0).innerText;
  const KanadaUid = await originals.getNthFolderUid(0);
  const SecondItemInVacation = await Selector("div.result").nth(1).innerText;
  await t
    .expect(FirstItemInVacation)
    .contains("Kanada")
    .expect(SecondItemInVacation)
    .contains("Korsika");
  await originals.openFolderWithUid(KanadaUid);
  const FirstItemInKanada = await Selector("div.result").nth(0).innerText;
  const SecondItemInKanada = await Selector("div.result").nth(1).innerText;
  await t
    .expect(FirstItemInKanada)
    .contains("BotanicalGarden")
    .expect(SecondItemInKanada)
    .contains("originals-001_2.jpg")
    .click(Selector("button").withText("BotanicalGarden"))
    .click(Selector('a[href="/library/files/Vacation"]'));
  await originals.triggerHoverAction("is-folder","uid", KanadaUid, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("album", "KanadaVacation", "");
  await menu.openPage("albums");
  await toolbar.search("KanadaVacation");
  const AlbumUid = await album.getNthAlbumUid("all", 0);
  await album.openAlbumWithUid(AlbumUid);
  const PhotoCountAfterAdd = await photo.getPhotoCount("all");
  await t.expect(PhotoCountAfterAdd).eql(2);
  await menu.openPage("albums");
  await album.triggerHoverAction("uid", AlbumUid, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.triggerContextMenuAction("delete", "", "");
});

test.meta("testID", "originals-002")("Download original files", async (t) => {
  await menu.openPage("originals");
  const FirstFile = await originals.getNthFileUid(0);
  await originals.triggerHoverAction("is-file", "uid", FirstFile, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.clearSelection();
  const FirstFolder = await originals.getNthFolderUid(0);
  await originals.triggerHoverAction("is-folder","uid", FirstFolder, "select");
  await contextmenu.checkContextMenuCount("1");
  await contextmenu.checkContextMenuActionAvailability("download", true);
  await contextmenu.clearSelection();
});
