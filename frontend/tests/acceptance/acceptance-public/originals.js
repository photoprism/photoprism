import { Selector } from "testcafe";
import testcafeconfig from "../../testcafeconfig.json";
import Menu from "../page-model/menu";
import Photo from "../page-model/photo";
import Toolbar from "../page-model/toolbar";
import ContextMenu from "../page-model/context-menu";
import Album from "../page-model/album";
import Originals from "../page-model/originals";

fixture`Test files`.page`${testcafeconfig.url}`;

const menu = new Menu();
const photo = new Photo();
const toolbar = new Toolbar();
const contextmenu = new ContextMenu();
const album = new Album();
const originals = new Originals();

test.meta("testID", "originals-001").meta({ type: "short", mode: "public" })(
  "Common: Navigate in originals",
  async (t) => {
    await menu.openPage("originals");
    await t.click(Selector("button").withText("Vacation"));
    const FirstItemInVacationName = await Selector("div.result", { timeout: 15000 }).nth(0)
      .innerText;
    const KanadaFolderUid = await originals.getNthFolderUid(0);
    const SecondItemInVacationName = await Selector("div.result").nth(1).innerText;

    await t
      .expect(FirstItemInVacationName)
      .contains("Kanada")
      .expect(SecondItemInVacationName)
      .contains("Korsika");

    await originals.openFolderWithUid(KanadaFolderUid);

    const FirstItemInKanadaName = await Selector("div.result").nth(0).innerText;
    const SecondItemInKanadaName = await Selector("div.result").nth(1).innerText;

    await t
      .expect(FirstItemInKanadaName)
      .contains("BotanicalGarden")
      .expect(SecondItemInKanadaName)
      .contains("originals-001_2.jpg");

    await t.click(Selector("button").withText("BotanicalGarden"));
    const FirstItemInBotanicalGardenName = await Selector("div.result", { timeout: 15000 }).nth(0)
      .innerText;
    await t.expect(FirstItemInBotanicalGardenName).contains("originals-001_1.jpg");
    await t.click(Selector('a[href="/library/index/files/Vacation"]'));
    const FolderCount = await originals.getFolderCount();

    await t.expect(FolderCount).eql(2);
  }
);

test.meta("testID", "originals-002").meta({ type: "short", mode: "public" })(
  "Common: Add original files to album",
  async (t) => {
    await menu.openPage("albums");
    await toolbar.search("KanadaVacation");

    await t.expect(Selector("div.no-results").visible).ok();

    await menu.openPage("originals");
    await t.click(Selector("button").withText("Vacation"));
    const KanadaFolderUid = await originals.getNthFolderUid(0);
    await originals.openFolderWithUid(KanadaFolderUid);
    const FilesCountInKanada = await originals.getFileCount();
    await t.click(Selector("button").withText("BotanicalGarden"));
    const FilesCountInKanadaSubfolder = await originals.getFileCount();
    await t.navigateTo("/library/index/files/Vacation");
    await originals.triggerHoverAction("is-folder", "uid", KanadaFolderUid, "select");
    await contextmenu.checkContextMenuCount("1");
    await contextmenu.triggerContextMenuAction("album", "KanadaVacation");
    await menu.openPage("albums");
    await toolbar.search("KanadaVacation");
    const AlbumUid = await album.getNthAlbumUid("all", 0);
    await album.openAlbumWithUid(AlbumUid);
    const PhotoCountAfterAdd = await photo.getPhotoCount("all");

    await t.expect(PhotoCountAfterAdd).eql(FilesCountInKanada + FilesCountInKanadaSubfolder);

    await menu.openPage("albums");
    await album.triggerHoverAction("uid", AlbumUid, "select");
    await contextmenu.checkContextMenuCount("1");
    await contextmenu.triggerContextMenuAction("delete", "");
  }
);

test.meta("testID", "originals-003").meta({ mode: "public" })(
  "Common: Download available in originals",
  async (t) => {
    await menu.openPage("originals");
    const FirstFile = await originals.getNthFileUid(0);
    await originals.triggerHoverAction("is-file", "uid", FirstFile, "select");
    await contextmenu.checkContextMenuCount("1");
    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.clearSelection();
    const FirstFolder = await originals.getNthFolderUid(0);
    await originals.triggerHoverAction("is-folder", "uid", FirstFolder, "select");
    await contextmenu.checkContextMenuCount("1");
    await contextmenu.checkContextMenuActionAvailability("download", true);
    await contextmenu.clearSelection();
  }
);
