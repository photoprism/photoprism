import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.openImportFolderSelect = Selector(".input-import-folder input", { timeout: 15000 });
    this.import = Selector(".action-import");
    this.indexFolderSelect = Selector(".input-index-folder input", { timeout: 15000 });
    this.index = Selector(".action-index");
    this.importTab = Selector("#tab-library-import", { timeout: 15000 });
    this.indexTab = Selector("#tab-library-index", { timeout: 15000 });
  }
}
